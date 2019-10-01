#!/bin/bash

set -o errexit
set -o nounset

i=0

until docker login -u appsody -p $(oc whoami -t) $(oc registry info) &> /dev/null
do
    echo "Waiting for oc registry pods to initialize ..."
    sleep 1
    # Timeout if registry has run into an issue of some sort.
    ((i++))
    if [[ "$i" == "30" ]]; then
        break;
    fi
done

echo "Logged into oc registry."
