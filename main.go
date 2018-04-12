package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	// TODO build our own cluster-controller informers and clientset
	"k8s.io/sample-controller/pkg/signals"
)

// Controller object
type Controller struct {
	// add cluster queue
	clientset    kubernetes.Interface
}

var (
	masterURL  string
	kubeconfig string
)


func (c *Controller) run(threadiness int, stopCh <-chan struct{}) error {
	// don't let panics crash the process
	defer utilruntime.HandleCrash()

	glog.Infof("Starting cluster-controller")

	for i := 0; i < threadiness; i++ {
		fmt.Println("check cluster status here\n")
	}

	// wait until we're told to stop
	<-stopCh
	glog.Infof("Shutting down cluster-controller")


	return nil
}


// return rest config, if path not specified assume in cluster config
func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

// create new controller
func newController(
	kubeclientset kubernetes.Interface,
	) *Controller {

	c := &Controller{
		clientset: kubeclientset,
	}
	return c
}


func main() {
	// pass kubeconfig like: -kubeconfig=$HOME/.kube/config
	kubeconf := flag.String("kubeconf", "admin.conf", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	config, err := getClientConfig(*kubeconf)
	if err != nil {
		panic(err.Error())
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	controller := newController(kubeClient)

	if err = controller.run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}