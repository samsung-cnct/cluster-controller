#!/bin/bash

# vars defined in controller-statefulset
# we will need to create these secrets in the CMC
echo "$SECRET_USERNAME"

# We need to create use the env vars to create the
# config files outlined in: https://github.com/DStorck/gogo/blob/master/README.md
# once we figure out how gogo and the controller will integrate we can finish this script



# start controller
/bin/controller/main -kubeconfig= -logtostderr=true