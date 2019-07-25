# Appsody Application Operator

The Appsody Application Operator has been designed to deploy applications created by [Appsody Application Stacks](https://appsody.dev/) into [OKD](https://www.okd.io/) clusters.  The goal of this project is to iterative grow the operator's set of day-2 capabilities.  If there's an certain functionality you would like to see or a bug you would like to report, please use our [issues tab](https://github.com/appsody/appsody-operator/issues) to get in contact with us.

## Operator Installation

You can install the Appsody Application Operator via a single `kubectl` command or assisted by the [Operator Lifecycle Manager](https://github.com/operator-framework/operator-lifecycle-manager).

### Direct installation

Run the following command to install the operator:  

* `kubectl apply -f <URL>`


### OLM-assisted installation

*Note:* OLM is labelled as a tech preview for OKD / OpenShift 3.11.  

* install OLM as described in [here](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/Documentation/install/install.md#installing-olm).
* install the operator via this command:

## Overview

The architecture of the Appsody Application Operator follows the basic controller pattern:  the Operator container with the controller is deployed into a Pod and listens for incoming resources with `Kind: AppsodyApplication`.  The 

![ICP OS](architecture.png)

## Application Deployments

Each application deployment will have a YAML file that specifies its configuration.  Here's an example:

```
apiVersion: appsody.dev/v1alpha1
kind: AppsodyApplication
metadata:
  name: example-appsodyapplication
spec:
  # Add fields here
  version: 1.0.0
  applicationImage: quay.io/arthurdm/myApp:1.0
  service:
    type: ClusterIP
    port: 9080
  expose: true
  storage:
    size: 2Gi
    mountPath: "/etc/websphere"
```

### Application deployment configuration

These are the available keys under the `spec` section of the Custom Resource file.  For the complete OpenAPI v3 representation of these values (including their types, etc), please see [this part](https://github.com/appsody/appsody-operator/blob/master/deploy/crds/appsody_v1alpha1_appsodyapplication_crd.yaml#L25) of the Custom Resource Definition.

The only required field is `applicationImage`. 

| Parameter | Description |
|---|---|
| `applicationImage`   | The absolute name of the image to be deployed, containing the registry and the tag. |
| `architecture`   | An array of architectures to be considered for deployment.  Their position in the array indicates preference. |
| `autoscaling.maxReplicas`   | Upper limit for the number of pods that can be set by the autoscaler.  Cannot be lower than the minimum number of replicas.|
| `autoscaling.minReplicas`   | Lower limit for the number of pods that can be set by the autoscaler.  Can only be 0 if `createKnativeService` is set to true. |
| `autoscaling.targetCPUUtilizationPercentage`   | Target average CPU utilization (represented as a percentage of requested CPU) over all the pods. |
| `createKnativeService`   | A boolean to toggle the creation of Knative resources and usage of Knative serving. |
| `env`   | An array of environment variables following the format of `{name, value}`, where value is a simple string. |
| `envFrom`   | An array of environment variables following the format of `{name, valueFrom}`, where `valueFrom` is an object containing a property named either `secretKeyRef` or `configMapKeyRef`, which in turn contain the properties name and key. |
| ` `   | ` ` |
| ` `   | ` ` |
| ` `   | ` ` |
| ` `   | ` ` |
| ` `   | ` ` |
| ` `   | ` ` |
| ` `   | ` ` |
| ` `   | ` ` |
| ` `   | ` ` |
