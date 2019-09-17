#!/bin/bash

set -o errexit
set -o nounset

oc policy add-role-to-user registry-viewer developer
oc policy add-role-to-user registry-editor developer

oc login -u developer

docker login -u developer -p $(oc whoami -t) $(oc registry info)
