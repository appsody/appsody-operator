#!/bin/bash

set -o errexit
set -o nounset

until docker login -u appsody -p $(oc whoami -t) $(oc registry info) &> /dev/null
do
    echo "Waiting for oc registry pods to initialize ...\n"
    oc whoami -t
    oc registry info
    sleep 1
done

echo "Logged into oc registry.\n"
