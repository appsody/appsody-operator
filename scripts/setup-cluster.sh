#!/bin/bash

set -o errexit
set -o nounset

# Install kubectl and oc
curl -L https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz | tar xvz
cd openshift-origin-clien*
sudo mv oc kubectl /usr/local/bin/
cd ..

# Start a cluster and login
oc cluster up --skip-registry-check=true
oc login -u system:admin
oc apply -f scripts/servicemonitor.crd.yaml