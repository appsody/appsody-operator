#!/bin/bash

cat << EOF  | sudo tee /etc/docker/daemon.json
  {
      "insecure-registries" : [ "172.30.0.0/16" ]
  }
EOF
