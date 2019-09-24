#!/bin/bash

set -o errexit
set -o nounset

# sudo - E env "PATH=$PATH" echo "{ \"insecure-registries\":[\"172.30.0.0/16\"] }" >> /etc/docker/daemon.json
echo "{ \"insecure-registries\":[\"172.30.0.0/16\"] }" | sudo tee -a /etc/docker/daemon.json
sudo service docker restart

oc create sa robot
oc login -u developer
oc policy add-role-to-user registry-viewer system:serviceaccounts
oc policy add-role-to-user registry-editor system:serviceaccounts
oc get routes
docker login -u robot -p $(oc sa get-token robot) $(oc registry info)
