#!/bin/bash

set -o errexit
set -o nounset



# echo "{ \"insecure-registries\":[\"$\"] }" >> /etc/docker/daemon.json
# service docker restart

oc create sa robot
oc login -u developer
oc policy add-role-to-user registry-viewer system:serviceaccounts
oc policy add-role-to-user registry-editor system:serviceaccounts
docker login -u robot -p $(oc sa get-token robot) $(oc registry info)

