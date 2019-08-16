#!/bin/bash
oc cluster up --skip-registry-check=true
oc login -u system:admin
oc create -f deploy/stack_defaults.yaml
oc create -f deploy/crds/appsody_v1alpha1_appsodyapplication_crd.yaml
oc create -f deploy/crds/appsody_v1alpha1_appsodyapplication_cr.yaml
operator-sdk test ./test/e2e --namespace myproject