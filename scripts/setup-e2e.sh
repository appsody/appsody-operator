#!/bin/bash

set -o errexit
set -o nounset

# oc adm ca create-server-cert \
#  --signer-cert=/etc/origin/master/ca.crt \
#  --signer-key=/etc/origin/master/ca.key  \
#  --signer-serial=/etc/origin/master/ca.serial.txt  \
#  --hostnames=$(oc registry info) \
#  --cert=/etc/secrets/registry.crt  \
#  --key=/etc/secrets/registry.key


oc create sa robot
oc policy add-role-to-user registry-viewer system:serviceaccounts
oc policy add-role-to-user registry-editor system:serviceaccounts
docker login -u robot -p $(oc sa get-token robot) $(oc registry info)

