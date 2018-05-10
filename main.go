package main

import (
	clusterErrors "errors"
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	// cluster-controller pkg libraries
	samsungv1alpha1 "github.com/samsung-cnct/cluster-controller/pkg/apis/clustercontroller/v1alpha1"
	clientset "github.com/samsung-cnct/cluster-controller/pkg/client/clientset/versioned"
	samsungscheme "github.com/samsung-cnct/cluster-controller/pkg/client/clientset/versioned/scheme"
	informers "github.com/samsung-cnct/cluster-controller/pkg/client/informers/externalversions"
	listers "github.com/samsung-cnct/cluster-controller/pkg/client/listers/clustercontroller/v1alpha1"
	"github.com/samsung-cnct/cluster-controller/pkg/signals"

	"github.com/dstorck/gogo"
)

const controllerAgentName = "cluster-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a KrakenCluster is synced
	SuccessSynced = "Synced"
	// MessageResourceSynced is the message used for an Event fired when a KrakenCluster
	// is synced successfully
	MessageResourceSynced = "KrakenCluster synced successfully"
	// CreatedStatus will show the juju status
	CreatedStatus = "CreatedStatus"
	// JujuBootstrapStatus will show the juju bootstrap status in an Event
	JujuBootstrapStatus = "JujuBootstrapStatus"
	// ClusterReady displays the Ready status in the resource status field
	ClusterReady = "Ready"
	// JujuBootstrapping displays the status in the resource status field
	JujuBootstrapping = "JujuBootstrapping"
	// JujuBootstrapReady displays this in the resource status field
	JujuBootstrapReady = "JujuBootstrapReady"
	// JujuBootstrapError displays this in the resource status field
	JujuBootstrapError = "JujuBootstrapError"
	// DeleteComplete is displayed in the resource status field when a delete is complete
	DeleteComplete = "DeleteComplete"
	// MaasEndpoint will show the endpoint to use for Maas - This will be input in the future
	MaasEndpoint = "http://192.168.2.24/MAAS/api/2.0"
	// JujuBundle is the bundle used to create the cluster - This will be input in the future
	JujuBundle = "cs:bundle/kubernetes-core-306"
)

// Controller object
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// samsungclientset is a clientset for our own api group
	samsungclientset clientset.Interface

	krakenclusterLister  listers.KrakenClusterLister
	krakenclustersSynced cache.InformerSynced
	workqueue            workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

var (
	masterURL  string
	kubeconfig string
)

func (c *Controller) run(threadiness int, stopCh <-chan struct{}) error {
	// don't let panics crash the process
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	glog.Info("Starting cluster-controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.krakenclustersSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	glog.Info("Started workers")
	// wait until we're told to stop
	<-stopCh
	glog.Info("Shutting down cluster-controller")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the KrakenCluster resource with this namespace/name
	kc, err := c.krakenclusterLister.KrakenClusters(namespace).Get(name)
	if err != nil {
		// The KrakenCluster resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("krakencluster '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	// Calls to Business logic go here
	err = c.processStates(kc)
	if err != nil {
		return err
	}
	c.recorder.Event(kc, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) processStates(kc *samsungv1alpha1.KrakenCluster) error {
	switch kc.Status.State {
	case samsungv1alpha1.Unknown:
		err := c.processUnknownState(kc)
		if err != nil {
			return err
		}
	case samsungv1alpha1.Creating:
		err := c.processCreatingState(kc)
		if err != nil {
			return err
		}
	case samsungv1alpha1.Created:
		err := c.processCreatedState(kc)
		if err != nil {
			return err
		}
	case samsungv1alpha1.Deleting:
		err := c.processDeletingState(kc)
		if err != nil {
			return err
		}
	case samsungv1alpha1.Deleted:
		err := c.processDeletedState(kc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) processUnknownState(kc *samsungv1alpha1.KrakenCluster) error {
	glog.Infof("Processing Unknown state for %s", kc.Spec.Cluster.ClusterName)
	// set initial State and Status
	err := c.updateKrakenClusterStatus(kc, samsungv1alpha1.Creating, JujuBootstrapping, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) processCreatingState(kc *samsungv1alpha1.KrakenCluster) error {
	glog.Infof("Processing Creating state for %s", kc.Spec.Cluster.ClusterName)
	// check for delete
	if kc.DeletionTimestamp != nil && (kc.Status.Status == JujuBootstrapReady || kc.Status.Status == JujuBootstrapError) {
		glog.Infof("Processing Delete for %s", kc.Spec.Cluster.ClusterName)
		err := c.updateKrakenClusterStatus(kc, samsungv1alpha1.Deleting, kc.Status.Status, nil)
		if err != nil {
			return err
		}
	} else if kc.Status.Status == JujuBootstrapping {
		// this will block
		err := c.createCluster(kc)
		if err != nil {
			return err
		}
	} else if kc.Status.Status == JujuBootstrapReady && c.isClusterReady(kc) {
		cluster := gogo.Juju{Name: string(kc.UID)}
		karray, err := cluster.GetKubeConfig()
		kubeconf := ""
		if err != nil {
			glog.Error(err)
		} else {
			kubeconf = string(karray)
		}
		jstatus, err := cluster.GetStatus()
		if err != nil {
			glog.Error(err)
			c.recorder.Event(kc, corev1.EventTypeWarning, CreatedStatus, err.Error())
		} else {
			c.recorder.Event(kc, corev1.EventTypeNormal, CreatedStatus, jstatus)
		}
		err = c.updateKrakenClusterStatus(kc, samsungv1alpha1.Created, ClusterReady, &kubeconf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) processCreatedState(kc *samsungv1alpha1.KrakenCluster) error {
	glog.Infof("Processing Created state for '%s'", kc.Spec.Cluster.ClusterName)
	// check for delete
	if kc.DeletionTimestamp != nil {
		glog.Infof("Processing Delete for %s", kc.Spec.Cluster.ClusterName)
		err := c.updateKrakenClusterStatus(kc, samsungv1alpha1.Deleting, JujuBootstrapReady, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) processDeletingState(kc *samsungv1alpha1.KrakenCluster) error {
	glog.Infof("Processing Deleting state for '%s'", kc.Spec.Cluster.ClusterName)
	status := DeleteComplete
	if kc.Status.Status == JujuBootstrapReady {
		// this call blocks
		c.deleteCluster(kc)
		err := c.updateKrakenClusterStatus(kc, samsungv1alpha1.Deleting, status, nil)
		if err != nil {
			return err
		}
	} else if kc.Status.Status == DeleteComplete {
		if c.isDestroyComplete(kc) {
			// allow the delete to finish
			err := c.updateKrakenClusterStatus(kc, samsungv1alpha1.Deleted, string(samsungv1alpha1.Deleted), nil)
			if err != nil {
				return err
			}
		}
	} else if kc.Status.Status == JujuBootstrapError {
		// allow the delete to finish
		err := c.updateKrakenClusterStatus(kc, samsungv1alpha1.Deleted, string(samsungv1alpha1.Deleted), nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) processDeletedState(kc *samsungv1alpha1.KrakenCluster) error {
	glog.Infof("Processing Deleted state for %s", kc.Spec.Cluster.ClusterName)
	// remove the Finalizer field so the resource can be deleted
	kc.SetFinalizers(nil)
	err := c.updateKrakenClusterStatus(kc, samsungv1alpha1.Deleted, string(samsungv1alpha1.Deleted), nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) updateKrakenClusterStatus(kc *samsungv1alpha1.KrakenCluster, state samsungv1alpha1.KrakenClusterState, status string, kubeconf *string) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	kcCopy := kc.DeepCopy()
	kcCopy.Status.State = state
	kcCopy.Status.Status = status
	if kubeconf != nil {
		kcCopy.Status.Kubeconfig = *kubeconf
	}
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.samsungclientset.SamsungV1alpha1().KrakenClusters(kc.Namespace).Update(kcCopy)
	return err
}

func (c *Controller) isClusterReady(kc *samsungv1alpha1.KrakenCluster) bool {
	cluster := gogo.Juju{Name: string(kc.UID)}
	ready, err := cluster.ClusterReady()
	if err != nil {
		glog.Warning(err)
	}
	return ready
}

func (c *Controller) isDestroyComplete(kc *samsungv1alpha1.KrakenCluster) bool {
	cluster := gogo.Juju{Name: string(kc.UID)}
	destroyed, err := cluster.DestroyComplete()
	if err != nil {
		glog.Warning(err)
	}
	return destroyed
}

func (c *Controller) initGogoInstance(kc *samsungv1alpha1.KrakenCluster) *gogo.Juju {
	switch kc.Spec.CloudProvider.Name {
	case samsungv1alpha1.MaasProvider:
		cluster := &gogo.Juju{
			Name:   string(kc.UID),
			Kind:   gogo.Maas,
			Bundle: JujuBundle,
			MaasCl: gogo.MaasCloud{
				Endpoint: MaasEndpoint,
			},
			MaasCr: gogo.MaasCredentials{
				Username:  kc.Spec.CloudProvider.Credentials.Username,
				MaasOauth: kc.Spec.CloudProvider.Credentials.Password,
			},
		}
		return cluster
	case samsungv1alpha1.AwsProvider:
		cluster := &gogo.Juju{
			Name:   string(kc.UID),
			Kind:   gogo.Aws,
			Bundle: JujuBundle,
			AwsCr: gogo.AWSCredentials{
				Username:  kc.Spec.CloudProvider.Credentials.Username,
				AccessKey: kc.Spec.CloudProvider.Credentials.Accesskey,
				SecretKey: kc.Spec.CloudProvider.Credentials.Password,
			},
			AwsCl: gogo.AWSCloud{
				Region: kc.Spec.CloudProvider.Region,
			},
		}
		return cluster
	}
	return nil
}

func (c *Controller) spinUp(kc *samsungv1alpha1.KrakenCluster) {
	cluster := c.initGogoInstance(kc)
	// this call blocks while bootstrapping juju
	status := JujuBootstrapReady
	spinupError := false
	err := cluster.Spinup()
	if err != nil {
		glog.Error(err.Error())
		spinupError = true
	}
	if spinupError {
		glog.Info("spinupError = true")
	}
	exists, err := cluster.ControllerReady()
	if err != nil {
		glog.Error(err.Error())
	}
	if !exists || spinupError {
		if !exists {
			glog.Error("Juju controller was not ready after call to ControllerReady")
		}
		status = JujuBootstrapError
	}
	glog.Info("status = ", status)
	// add Finalizer so the resource won't be deleted immediately on delete kc
	kc.SetFinalizers([]string{"samsung.cnct.com/finalizer"})
	err = c.updateKrakenClusterStatus(kc, samsungv1alpha1.Creating, status, nil)
	if err != nil {
		glog.Error(err.Error())
	}
}

func (c *Controller) createCluster(kc *samsungv1alpha1.KrakenCluster) error {
	// TODO Use apiMachinary errors instead
	if kc.Spec.CloudProvider.Name != samsungv1alpha1.MaasProvider && kc.Spec.CloudProvider.Name != samsungv1alpha1.AwsProvider {
		return clusterErrors.New("Invalid Cloudprovider.  Valid providers are: maas, aws")
	}
	c.spinUp(kc)
	return nil
}

func (c *Controller) deleteCluster(kc *samsungv1alpha1.KrakenCluster) error {
	cluster := new(gogo.Juju)
	cluster.Name = string(kc.UID)
	if kc.Spec.CloudProvider.Name == samsungv1alpha1.AwsProvider {
		cluster.Kind = gogo.Aws
		cluster.AwsCl = gogo.AWSCloud{
			Region: kc.Spec.CloudProvider.Region}
	} else if kc.Spec.CloudProvider.Name == samsungv1alpha1.MaasProvider {
		cluster.Kind = gogo.Maas
	}
	err := cluster.DestroyCluster()
	if err != nil {
		glog.Error(err)
		return err
	}
	return nil
}

// enqueueKrakenCluster takes a KrakenCluster resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than KrakenCluster.
func (c *Controller) enqueueKrakenCluster(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// return rest config, if path not specified assume in cluster config
func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	}
	return rest.InClusterConfig()
}

// create new controller
func newController(
	kubeclientset kubernetes.Interface,
	samsungclientset clientset.Interface,
	samsungInformerFactory informers.SharedInformerFactory,
) *Controller {
	// obtain references to the shared index informer for KrakenCluster types
	krakenclusterInformer := samsungInformerFactory.Samsung().V1alpha1().KrakenClusters()

	// create event broadcaster so events can be logged for krakencluster types
	samsungscheme.AddToScheme(scheme.Scheme)
	glog.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	c := &Controller{
		kubeclientset:        kubeclientset,
		samsungclientset:     samsungclientset,
		krakenclusterLister:  krakenclusterInformer.Lister(),
		krakenclustersSynced: krakenclusterInformer.Informer().HasSynced,
		workqueue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "KrakenClusters"),
		recorder:             recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when KrakenCluster resources change
	krakenclusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueKrakenCluster,
		UpdateFunc: func(old, new interface{}) {
			glog.Info("Update called")
			c.enqueueKrakenCluster(new)
		},
	})

	return c
}

func main() {
	// pass kubeconfig like: -kubeconfig=$HOME/.kube/config
	// incluster config: -kubeconfig=""
	kubeconf := flag.String("kubeconfig", "admin.conf", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	config, err := getClientConfig(*kubeconf)
	if err != nil {
		glog.Fatalf("Error loading cluster config: %s", err.Error())
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	samsungClient, err := clientset.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Error building example clientset: %s", err.Error())
	}
	glog.Info("Constructing informer factory")
	samsungInformerFactory := informers.NewSharedInformerFactory(samsungClient, time.Second*30)

	glog.Info("Constructing controller")
	controller := newController(kubeClient, samsungClient, samsungInformerFactory)
	go samsungInformerFactory.Start(stopCh)

	if err = controller.run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}
