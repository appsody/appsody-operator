#!/bin/bash

set -o errexit
set -o nounset

echo " {\"insecure-registries\": [\"172.30.0.0/16\"] } " >> /etc/docker/daemon.json

oc create sa robot
oc policy add-role-to-user registry-viewer system:serviceaccounts
oc policy add-role-to-user registry-editor system:serviceaccounts
docker login -u robot -p $(oc sa get-token robot) $(oc registry info)

