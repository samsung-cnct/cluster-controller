# cluster-controller

This repository implements a controller for watching Cluster resources as
defined with a CustomResourceDefinition (CRD).

## Purpose

This custom controller manages a custom resource of type `Cluster`.

## Running

```sh
# assumes you have a working kubeconfig, not required if operating in-cluster
$ go run *.go -kubeconfig=$HOME/.kube/config
```
