#!/bin/bash


# Restart docker daemon for insecure registry access
restart_daemon() {
cat << EOF  | sudo tee /etc/docker/daemon.json
  {
      "insecure-registries" : [ "172.30.0.0/16" ]
  }
EOF

sudo systemctl restart docker
}

setup_cluster(){
    # Install kubectl and oc
    curl -L https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz | tar xvz
    cd openshift-origin-clien*
    sudo mv oc kubectl /usr/local/bin/
    cd ..
    # Start a cluster and login
    oc login $CLUSTER_URL --token=$CLUSTER_TOKEN
    export DEFAULT_REGISTRY=$(oc get route docker-registry -o jsonpath="{ .spec.host }" -n default)
    export BUILD_IMAGE=$DEFAULT_REGISTRY/openshift/application-operator:daily
}

# Log in to docker daemon with openshift cluster registry
docker_login() {
    i=0
    # Cluster up doesn't wait for registry so have to poll for ready state
    until docker login -u unused -p $CLUSTER_TOKEN $DEFAULT_REGISTRY &> /dev/null
    do
        echo "> Waiting for oc registry pods to initialize ..."
        sleep 1
        # Timeout if registry has run into an issue of some sort.
        ((i++))
        if [[ "$i" == "30" ]]; then
            echo "> Failed to connect to registry, logging state of default namespace: "
            echo "Default pods:"
            oc get pods -n default
            echo "Default services:"
            oc get svc -n default
            break;
        fi
    done

    echo "> Logged into oc registry."
}

main() {
    echo "****** Restarting daemon for insecure registry..."
    restart_daemon
    echo "****** Setting up cluster..."
    setup_cluster
    echo "****** Logging into local registry..."
    docker_login
    echo "****** Building image"
    operator-sdk build $BUILD_IMAGE
    echo "****** Pushing image into registry..."
    docker push $BUILD_IMAGE
    echo "****** Starting e2e tests..."
    operator-sdk test local github.com/appsody/appsody-operator/test/e2e --go-test-flags "-timeout 25m" --image $BUILD_IMAGE --verbose
}

main
