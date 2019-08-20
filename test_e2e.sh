#!/bin/bash
oc cluster up --skip-registry-check=true
oc login -u system:admin
oc apply -f deploy/cluster_role_binding.yaml
oc apply -f deploy/cluster_role.yaml
oc apply -f deploy/service_account.yaml
oc apply -f deploy/crds/appsody_v1alpha1_appsodyapplication_crd.yaml
oc apply -f deploy/crds/appsody_v1alpha1_appsodyapplication_cr.yaml
operator-sdk up local
operator-sdk test local ./test/e2e --up-local --namespace myproject