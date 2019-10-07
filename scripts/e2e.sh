#!/bin/bash

DEFAULT_REGISTRY=172.30.1.1:5000
BUILD_IMAGE=$DEFAULT_REGISTRY/openshift/appsody-operator:daily
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
    oc cluster up
    oc create secret docker-registry internal-registry --docker-email=unused --docker-password=$(oc whoami -t) --docker-username=unused --docker-server=$DEFAULT_REGISTRY
    oc login -u system:admin
    # Link accounts to internal docker registry
    oc create -f deploy/service_account.yaml
    if oc secrets link appsody-operator internal-registry; then
        echo "> Successfully linked registry secret to service account"
    else
        echo "> Failed to link registry secret to service account"
    fi
    # Add necessary roles to the developer user so he can interact with internal registry
    oc adm policy add-role-to-user registry-viewer developer
    oc adm policy add-role-to-user registry-editor developer
    oc adm policy add-role-to-user system:image-builder developer -n openshift
    oc adm policy add-role-to-user system:image-builder system:serviceaccounts
    oc login -u developer
}

# Log in to docker daemon with openshift cluster registry
docker_login() {
    i=0
    # Cluster up doesn't wait for registry so have to poll for ready state
    until docker login -u appsody -p $(oc whoami -t) $DEFAULT_REGISTRY &> /dev/null
    do
        echo "> Waiting for oc registry pods to initialize ..."
        sleep 1
        # Timeout if registry has run into an issue of some sort.
        ((i++))
        if [[ "$i" == "30" ]]; then
            echo "> Failed to connect to registry, logging state of default namespace: "
            oc login -u system:admin
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
    echo "****** Building image..."
    operator-sdk build $BUILD_IMAGE
    echo "****** Setting up cluster..."
    setup_cluster
    echo "****** Logging into local registry..."
    docker_login
    echo "****** Pushing image into registry..."
    docker push $BUILD_IMAGE
    echo "****** Starting e2e tests..."
    oc login -u system:admin
    operator-sdk test local github.com/appsody/appsody-operator/test/e2e --go-test-flags "-timeout 25m" --image $BUILD_IMAGE --verbose
}

main