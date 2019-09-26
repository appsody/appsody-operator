#!/bin/bash

set -o errexit
set -o nounset

REGISTRY_IP=${1}
IMAGE=${2}

echo $REGISTRY_IP/$IMAGE

oc create sa robot
oc policy add-role-to-user registry-viewer system:serviceaccounts
oc policy add-role-to-user registry-editor system:serviceaccounts

oc login -u developer
docker login -u robot -p $(oc sa get-token robot) $REGISTRY_IP:5000
docker push $REGISTRY_IP:5000/$IMAGE