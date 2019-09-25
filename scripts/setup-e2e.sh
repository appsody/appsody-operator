#!/bin/bash

set -o errexit
set -o nounset

oc create sa robot
oc policy add-role-to-user registry-viewer system:serviceaccounts
oc policy add-role-to-user registry-editor system:serviceaccounts
oc login -u developer
