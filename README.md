# cluster-controller

This repository implements a controller for watching Cluster resources as
defined with a CustomResourceDefinition (CRD).

This is a WIP repo for a POC, TODO alter directory structure and create a CI job to handle codegen, see: https://blog.openshift.com/kubernetes-deep-dive-code-generation-customresources/

## Purpose

This custom controller manages a custom resource of type `Cluster`.

## Deploy CRD and create a resource
Creating the KrakenCluster CRD object that defines the schema of a kraken cluster
and the resource to be consumed by the controller:
```sh
kubectl create -f crd/KrakenClusterCRD.yaml
```

You may then create a sample resource by running
```sh
kubectl create-f crd/test-cluster.yaml
```

## Running

```sh
# assumes you have a working kubeconfig, not required if operating in-cluster
$ go run *.go -kubeconfig=$HOME/.kube/config
```
