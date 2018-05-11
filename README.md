# cluster-controller

This repository implements a controller for watching Cluster resources as
defined with a CustomResourceDefinition (CRD).

This is a WIP repo for a POC, TODO alter directory structure and create a CI job to handle codegen, see: https://blog.openshift.com/kubernetes-deep-dive-code-generation-customresources/

## Purpose

This custom controller manages a custom resource of type `Cluster`.

## Deploy CRD and create a resource

Make sure your cluster has the clusterrole and clusterbindings created. If not:
`$ kubectl create -f assets/clusterrole.yaml`
`$ kubectl create -f assets/clusterrolebinding.yaml`

`helm install chartmuseum/cluster-controller --version 0.1.3`

You may then create a sample resource by running
```sh
kubectl create -f assets/test-aws.yaml
```
or
`kubectl create -f assets/test-maas.yaml`

Be sure to plug your credentials into those files. Monitor the `KrakenCluster` resources with:
`kubectl get kc`
`kubectl describe kc <name>`

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
`$ helm repo add chartmuseum https://charts.migrations.cnct.io`.

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

Nothing blowing up? Let's try to deploy a new KrakenCluster object! In `./assets` you will find a `test-aws.yaml`. Put your AWS creds and username in that file, rename the cluster something meaningful and create it: 
`$ kubectl create -f assets/test-aws.yaml`

You can see/describe your new resource with `$ kubectl get kc` and watch the logs from the controller pod to check the latest status. You can also monitor the cluster with the AWS UI.

To delete your cluster run:

`$ kubectl delete <my-kc-resource>`

NOTE: Delete your `kc` resources BEFORE you delete the cluster-controller. If you don't you will not be able to delete your `kc` resource or your `KrakenCluster` crd resource.