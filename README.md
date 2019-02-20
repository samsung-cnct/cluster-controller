# cluster-controller

This repository implements a controller for watching Cluster resources as
defined with a CustomResourceDefinition (CRD).

This is a WIP repo for a POC, TODO alter directory structure and create a CI job to handle codegen, see: https://blog.openshift.com/kubernetes-deep-dive-code-generation-customresources/

## Purpose

This custom controller manages a custom resource of type `KrakenCluster`.

## Deploy CRD and create a resource

If your cluster has RBAC enabled you will need to create a clusterrole and clusterbinding:
```
$ kubectl create -f assets/clusterrole.yaml
$ kubectl create -f assets/clusterrolebinding.yaml
```

```
$ helm install chartmuseum/cluster-controller --version 0.1.3
```

You may then create a sample resource by updating either `assets/test-aws.yaml` or `assets/test-maas.yaml`.

For an **aws** cluster make sure you update the following fields in `assets/test-aws.yaml`:
```
metadata
  name : # name of your resource, will show up with $ kubectl get kc ,  should be unique for namespace
spec:
  cloudProvider:
    credentials:
      username: # aws username or iam profile name to be used
      accesskey: # aws access key id
      password: # aws secret access key
    region: # optional , but ideally would be set. ex = "aws/us-west-2"
  provisioner:
    bundle: # juju bundle desired for kubernetes cluster creation. ex - cs:bundle/kubernetes-core-306
  cluster:
    clusterName: # name of your cluster
```
For a **maas** cluster make sure you update the following fields in `assets/test-maas.yaml`:
```
metadata
  name : # name of your resource, will show up with $ kubectl get kc, should be unique for namespace
spec:
  cloudProvider:
    credentials:
      username: # maas username
      password: # maas api key
  provisioner:
    bundle: # juju bundle desired for kubernetes cluster creation. ex - cs:bundle/kubernetes-core-306
    maasEndpoint: # ex-"http://your-ip/MAAS/api/2.0"
  cluster:
    clusterName: # name of your cluster
```

Then run:
```
kubectl create -f assets/test-aws.yaml
```
or
```
kubectl create -f assets/test-maas.yaml
```

Monitor the `KrakenCluster` resources with:
```
kubectl get kc
kubectl describe kc <metadata.name>
```

## Running Locally Without a Container

```sh
# assumes you have a working kubeconfig, not required if operating in-cluster
$ go run *.go -kubeconfig=$HOME/.kube/config -logtostderr=true
```

## Changing the Specification for KrakenCluster
```sh
vi pkg/apis/clustercontroller/v1alpha1/types.go
hack/update-codegen.sh
```

## To Test Locally With Minikube and the Helm Chart

Start a minikube cluster - set your Kubernetes version to v1.10.0, a [bug](https://github.com/kubernetes/kubernetes/issues/61178) in v1.9.2 (default) will cause you some volume grief.

`$ minikube start --kubernetes-version  v1.10.0`

Start tiller:
`$ helm init`

Make sure you have the `chartmuseum` repo with `$ helm repo list`. If not:
`$ helm repo add chartmuseum https://charts.cnct.io`.

Once `chartmuseum` is added run `$ helm repo update` for good measure. You can check you have the most current version of our chart `cluster-controller` by running `$ helm search chartmuseum`. The most current helm chart version is pinned in the `.versionfile` in this repo's home directory.

Install the clusterrole and clusterbindings for the controller with kubectl:
`$ kubectl create -f assets/clusterrole.yaml`
`$ kubectl create -f assets/clusterrolebinding.yaml`

Ok! Tiller pods are running, clusterrole and clusterbindings are created, Helm has the most current version of our chart - let's install the chart!

`$ helm install chartmuseum/cluster-controller`

If there is not installation error, let's see if everything is running.  

Helm chart status:
`$ helm list`
Check the controller pod:
`$ kubectl get po`
You can check the controller logs with `$ kubectl logs <controller pod>` to make sure the controller is running and check your cluster status.

Let's make sure the CRD is installed:
`$ kubectl get crd`

Nothing blowing up? Let's try to deploy a new KrakenCluster object! In `./assets` you will find a `test-aws.yaml`. Put your AWS or Maas creds ( specified above ) in that file, rename the cluster something meaningful and create it:
`$ kubectl create -f assets/test-aws.yaml`

You can see/describe your new resource with `$ kubectl get kc` and watch the logs from the controller pod to check the latest status. You can also monitor the cluster with the AWS UI.

To delete your cluster run:

`$ kubectl delete kc <my-kc-resource>`

NOTE: Delete your `kc` resources BEFORE you delete the cluster-controller. If you don't you will not be able to delete your `kc` resource or your `KrakenCluster` crd resource.
