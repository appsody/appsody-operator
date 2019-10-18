#!/bin/bash

login_cluster(){
    # Install kubectl and oc
    curl -L https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz | tar xvz
    cd openshift-origin-clien*
    sudo mv oc kubectl /usr/local/bin/
    cd ..
    # Start a cluster and login
    oc login $CLUSTER_URL --token=$CLUSTER_TOKEN
    # Set variables for rest of script to use
    export DEFAULT_REGISTRY=$(oc get route docker-registry -o jsonpath="{ .spec.host }" -n default)
    export BUILD_IMAGE=$DEFAULT_REGISTRY/openshift/application-operator-$TRAVIS_BUILD_NUMBER:daily
}

cleanup() {
    # Remove image from the local registry after test has finished
    oc delete imagestream application-operator-$TRAVIS_BUILD_NUMBER -n openshift
    # ---- Extend cleanup as needed below ----
}

main() {
    echo "****** Logging into remote cluster..."
    login_cluster
    echo "****** Logging into local registry..."
    docker login -u unused -p $CLUSTER_TOKEN $DEFAULT_REGISTRY
    echo "****** Building image"
    operator-sdk build $BUILD_IMAGE
    echo "****** Pushing image into registry..."
    docker push $BUILD_IMAGE
    ## Use internal registry address as the pull will happen internally
    echo "****** Starting e2e tests..."
    operator-sdk test local github.com/appsody/appsody-operator/test/e2e --go-test-flags "-timeout 35m" --image $(oc registry info)/openshift/application-operator-$TRAVIS_BUILD_NUMBER:daily --verbose
    echo "****** Cleaning up tests..."
    cleanup
}

main
