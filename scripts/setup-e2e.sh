#!/bin/bash

set -o errexit
set -o nounset

# oc create sa robot -n myproject
# oc policy add-role-to-user registry-viewer system:serviceaccounts
# oc policy add-role-to-user registry-editor system:serviceaccounts
# oc login -u developer
docker login -u openshift -p $(oc whoami -t) 172.30.1.1:5000