# Appsody Operator

The Appsody Operator can be used to deploy applications created by [Appsody Application Stacks](https://appsody.dev/) into [OKD](https://www.okd.io/) or [OpenShift](https://www.openshift.com/) clusters.

This documentation refers to the latest codebase.  For documentation and samples of older releases, please check out the [main releases](https://github.com/appsody/appsody-operator/releases) page and navigate the corresponding tag.

## Operator installation

Use the instructions for one of the [releases](../deploy/releases) to install the operator into a Kubernetes cluster.

The Appsody Operator can be installed to:

- watch own namespace
- watch another namespace
- watch multiple namespaces
- watch all namespaces in the cluster

Appropriate cluster roles and bindings are required to watch another namespace, watch multiple namespaces or watch all namespaces.

NOTE: The Appsody Operator can only interact with resources it is given permission to interact through [Role-based access control (RBAC)](https://kubernetes.io/docs/reference/access-authn-authz/rbac). Some of the operator features require interacting with resources in other namespaces. In that case, the operator must be installed with correct `ClusterRole` definitions.

## Overview

The architecture of the Appsody Operator follows the basic controller pattern:  the Operator container with the controller is deployed into a Pod and listens for incoming resources with `Kind: AppsodyApplication`. Creating an `AppsodyApplication` custom resource (CR) triggers the Appsody Operator to create, update or delete Kubernetes resources needed by the application to run on your cluster.

Each instance of `AppsodyApplication` CR represents the application to be deployed on the cluster:

```yaml
apiVersion: appsody.dev/v1beta1
kind: AppsodyApplication
metadata:
  name: my-appsody-app
spec:
  stack: java-microprofile
  applicationImage: quay.io/my-repo/my-app:1.0
  service:
    type: ClusterIP
    port: 9080
  expose: true
  storage:
    size: 2Gi
    mountPath: "/logs"
```

## Configuration

### Custom Resource Definition (CRD)

The following table lists configurable parameters of the `AppsodyApplication` CRD. For complete OpenAPI v3 representation of these values please see [`AppsodyApplication` CRD](../deploy/crds/appsody.dev_appsodyapplications_crd.yaml).

Each `AppsodyApplication` CR must at least specify the `applicationImage` parameter. Specifying other parameters is optional.

| Parameter                                    | Description                                                                                                                                                                                                                                                                                                                                                                                                |
|----------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `stack`                                      | The name of the Appsody Application Stack that produced this application image.                                                                                                                                                                                                                                                                                                                            |
| `version`                                    | The current version of the application. Label `app.kubernetes.io/version` will be added to all resources when the version is defined.                                                                                                                                                                                                                                                                      |
| `serviceAccountName`                         | The name of the OpenShift service account to be used during deployment.                                                                                                                                                                                                                                                                                                                                    |
| `applicationImage`                           | The Docker image name to be deployed. On OpenShift, it can also be set to `<project name>/<image stream name>[:<tag>]` to reference an image from an image stream. If `<project name>` and `<tag>` values are not defined, they default to the namespace of the CR and the value of `latest`, respectively.                                                                                                |
| `applicationName`                            | The name of the application this resource is part of. If not specified, it defaults to the name of the CR.                                                                                                                                                                                                                                                                                                 |
| `createAppDefinition`                        | A boolean to toggle the automatic configuration of Kubernetes resources for the `AppsodyApplication` CR to allow creation of an application definition by [kAppNav](https://kappnav.io/). The default value is `true`. See [Application Navigator](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#kubernetes-application-navigator-kappnav-support) for more information.                                                                                |
| `pullPolicy`                                 | The policy used when pulling the image.  One of: `Always`, `Never`, and `IfNotPresent`.                                                                                                                                                                                                                                                                                                                    |
| `pullSecret`                                 | If using a registry that requires authentication, the name of the secret containing credentials.                                                                                                                                                                                                                                                                                                           |
| `initContainers`                             | The list of [Init Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#container-v1-core) definitions.                                                                                                                                                                                                                                                                          |
| `sidecarContainers`                          | The list of `sidecar` containers. These are additional containers to be added to the pods. Note: Sidecar containers should not be named `app`.                                                                                                                                                                                                                                                             |
| `architecture`                               | An array of architectures to be considered for deployment. Their position in the array indicates preference.                                                                                                                                                                                                                                                                                               |
| `bindings.embedded`                          | A YAML object that represents a `ServiceBindingRequest` custom resource.                                                                                                                                                                                                                                                                                                                                   |
| `bindings.autoDetect`                        | A boolean to toggle whether the operator should automatically detect and use a `ServiceBindingRequest` resource with `<CR_NAME>-binding` naming format. The default value for this parameter is `true`.                                                                                                                                                                                                    |
| `bindings.resourceRef`                       | The name of a `ServiceBindingRequest` custom resource created manually in the same namespace as the application.                                                                                                                                                                                                                                                                                           |
| `service.portName`                           | The name for the port exposed by the container.                                                                                                                                                                                                                                                                                                                                                            |
| `service.targetPort`                         | The port that the appsody application uses within the container. Defaults to the value of `service.port`.                                                                                                                                                                                                                                                                                                  |
| `service.ports`                              | An array consisting of service ports.                                                                                                                                                                                                                                                                                                                                                                      |
| `service.type`                               | The Kubernetes [Service Type](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types).                                                                                                                                                                                                                                                                         |
| `service.nodePort`                           | Node proxies this port into your service. Please note once this port is set to a non-zero value it cannot be reset to zero.                                                                                                                                                                                                                                                                                |
| `service.annotations`                        | Annotations to be added to the service.                                                                                                                                                                                                                                                                                                                                                                    |
| `service.certificate`                        | A YAML object representing a [Certificate](https://cert-manager.io/docs/reference/api-docs/#cert-manager.io/v1alpha2.CertificateSpec).                                                                                                                                                                                                                                                                     |
| `service.certificateSecretRef`               | A name of a secret that already contains TLS key, certificate and CA to be mounted in the pod.                                                                                                                                                                                                                                                                                                             |
| `service.provides.category`                  | Service binding type to be provided by this CR. At this time, the only allowed value is `openapi`.                                                                                                                                                                                                                                                                                                         |
| `service.provides.protocol`                  | Protocol of the provided service. Defaults to `http`.                                                                                                                                                                                                                                                                                                                                                      |
| `service.provides.context`                   | Specifies the context root of the service.                                                                                                                                                                                                                                                                                                                                                                 |
| `service.provides.auth.username`             | Optional value to specify username as [SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#secretkeyselector-v1-core).                                                                                                                                                                                                                                                 |
| `service.provides.auth.password`             | Optional value to specify password as [SecretKeySelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#secretkeyselector-v1-core).                                                                                                                                                                                                                                                 |
| `service.consumes`                           | An array consisting of services to be consumed by the `AppsodyApplication`.                                                                                                                                                                                                                                                                                                                                |
| `service.consumes[].category`                | The type of service binding to be consumed. At this time, the only allowed value is `openapi`.                                                                                                                                                                                                                                                                                                             |
| `service.consumes[].name`                    | The name of the service to be consumed. If binding to an `AppsodyApplication`, then this would be the provider's CR name.                                                                                                                                                                                                                                                                                  |
| `service.consumes[].namespace`               | The namespace of the service to be consumed. If binding to an `AppsodyApplication`, then this would be the provider's CR namespace.                                                                                                                                                                                                                                                                        |
| `service.consumes[].mountPath`               | Optional field to specify which location in the pod, service binding secret should be mounted. If not specified, the secret keys would be injected as environment variables.                                                                                                                                                                                                                               |
| `createKnativeService`                       | A boolean to toggle the creation of Knative resources and usage of Knative serving.                                                                                                                                                                                                                                                                                                                        |
| `expose`                                     | A boolean that toggles the external exposure of this deployment via a Route or a Knative Route resource.                                                                                                                                                                                                                                                                                                   |
| `replicas`                                   | The static number of desired replica pods that run simultaneously.                                                                                                                                                                                                                                                                                                                                         |
| `autoscaling.maxReplicas`                    | Required field for autoscaling. Upper limit for the number of pods that can be set by the autoscaler. It cannot be lower than the minimum number of replicas.                                                                                                                                                                                                                                              |
| `autoscaling.minReplicas`                    | Lower limit for the number of pods that can be set by the autoscaler.                                                                                                                                                                                                                                                                                                                                      |
| `autoscaling.targetCPUUtilizationPercentage` | Target average CPU utilization (represented as a percentage of requested CPU) over all the pods.                                                                                                                                                                                                                                                                                                           |
| `resourceConstraints.requests.cpu`           | The minimum required CPU core. Specify integers, fractions (e.g. 0.5), or millicore values(e.g. 100m, where 100m is equivalent to .1 core). Required field for autoscaling.                                                                                                                                                                                                                                |
| `resourceConstraints.requests.memory`        | The minimum memory in bytes. Specify integers with one of these suffixes: E, P, T, G, M, K, or power-of-two equivalents: Ei, Pi, Ti, Gi, Mi, Ki.                                                                                                                                                                                                                                                           |
| `resourceConstraints.limits.cpu`             | The upper limit of CPU core. Specify integers, fractions (e.g. 0.5), or millicores values(e.g. 100m, where 100m is equivalent to .1 core).                                                                                                                                                                                                                                                                 |
| `resourceConstraints.limits.memory`          | The memory upper limit in bytes. Specify integers with suffixes: E, P, T, G, M, K, or power-of-two equivalents: Ei, Pi, Ti, Gi, Mi, Ki.                                                                                                                                                                                                                                                                    |
| `env`                                        | An array of environment variables following the format of `{name, value}`, where value is a simple string. It may also follow the format of `{name, valueFrom}`, where `valueFrom` refers to a value in a `ConfigMap` or `Secret` resource. See [Environment variables](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#environment-variables) for more info. |
| `envFrom`                                    | An array of references to `ConfigMap` or `Secret` resources containing environment variables. Keys from `ConfigMap` or `Secret` resources become environment variable names in your container. See [Environment variables](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#environment-variables) for more info.                                            |
| `readinessProbe`                             | A YAML object configuring the [Kubernetes readiness probe](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/#define-readiness-probes) that controls when the pod is ready to receive traffic.                                                                                                                                                                  |
| `livenessProbe`                              | A YAML object configuring the [Kubernetes liveness probe](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/#define-a-liveness-http-request) that controls when Kubernetes needs to restart the pod.                                                                                                                                                            |
| `volumes`                                    | A YAML object representing a [pod volume](https://kubernetes.io/docs/concepts/storage/volumes).                                                                                                                                                                                                                                                                                                            |
| `volumeMounts`                               | A YAML object representing a [pod volumeMount](https://kubernetes.io/docs/concepts/storage/volumes/).                                                                                                                                                                                                                                                                                                      |
| `storage.size`                               | A convenient field to set the size of the persisted storage. Can be overridden by the `storage.volumeClaimTemplate` property.                                                                                                                                                                                                                                                                              |
| `storage.mountPath`                          | The directory inside the container where this persisted storage will be bound to.                                                                                                                                                                                                                                                                                                                          |
| `storage.volumeClaimTemplate`                | A YAML object representing a [volumeClaimTemplate](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#components) component of a `StatefulSet`.                                                                                                                                                                                                                                        |
| `monitoring.labels`                          | Labels to set on [ServiceMonitor](https://github.com/coreos/prometheus-operator/blob/master/Documentation/api.md#servicemonitor).                                                                                                                                                                                                                                                                          |
| `monitoring.endpoints`                       | A YAML snippet representing an array of [Endpoint](https://github.com/coreos/prometheus-operator/blob/master/Documentation/api.md#endpoint) component from ServiceMonitor.                                                                                                                                                                                                                                 |
| `route.annotations`                          | Annotations to be added to the service.                                                                                                                                                                                                                                                                                                                                                                    |
| `route.host`                                 | Hostname to be used for the Route.                                                                                                                                                                                                                                                                                                                                                                         |
| `route.path`                                 | Path to be used for Route.                                                                                                                                                                                                                                                                                                                                                                                 |
| `route.termination`                          | TLS termination policy. Can be one of `edge`, `reencrypt` and `passthrough`.                                                                                                                                                                                                                                                                                                                               |
| `route.insecureEdgeTerminationPolicy`        | HTTP traffic policy with TLS enabled. Can be one of `Allow`, `Redirect` and `None`.                                                                                                                                                                                                                                                                                                                        |
| `route.certificate`                          | A YAML object representing a [Certificate](https://cert-manager.io/docs/reference/api-docs/#cert-manager.io/v1alpha2.CertificateSpec).                                                                                                                                                                                                                                                                     |
| `route.certificateSecretRef`                 | A name of a secret that already contains TLS key, certificate and CA to be used in the route. Also can contain destination CA certificate.                                                                                                                                                                                                                                                                 |

### Basic usage

To deploy a Docker image containing an Appsody based application to a Kubernetes environment you can use the following CR:

 ```yaml
apiVersion: appsody.dev/v1beta1
kind: AppsodyApplication
metadata:
  name: my-appsody-app
spec:
  stack: java-microprofile
  applicationImage: quay.io/my-repo/my-app:1.0
```

The `applicationImage` value must be defined in the `AppsodyApplication` CR. On OpenShift, the operator tries to find an image stream name with the `applicationImage` value. The operator falls back to the registry lookup if it is not able to find any image stream that matches the value. If you want to distinguish an image stream called `my-company/my-app` (project: `my-company`, image stream name: `my-app`) from the Docker Hub `my-company/my-app` image, you can use the full image reference as `docker.io/my-company/my-app`.

The `stack` parameter must be the same value as the [Appsody application stack](https://github.com/appsody/stacks) you used to create your application.

To get information on the deployed CR, use either of the following:

```sh
oc get appsodyapplication my-appsody-app
oc get app my-appsody-app
```

The short name for `appsodyapplication` is `app`.

### Common Component Documentation

Appsody Application Operator is based on the generic [Runtime Component Operator](https://github.com/application-stacks/runtime-component-operator). To see more
information on the usage of common functionality, see the Runtime Component Operator
documentation below. Note that, in the samples from the links below, the instances of `Kind:
RuntimeComponent` must be replaced with `Kind: AppsodyApplication`.

- [Image Streams](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#Image-streams)
- [Service Account](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#Service-account)
- [Labels](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#Labels)
- [Annotations](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#Annotations)
- [Environment Variables](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#Environment-variables)
- [High Availability](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#High-availability)
- [Service Ports](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#Service-ports)
- [Persistence](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#Persistence)
- [Service Binding](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#Service-binding)
- [Monitoring](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#Monitoring)
- [Knative Support](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#Knative-support)
- [Exposing Service](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#Exposing-service-externally)
- [Kubernetes Application Navigator](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#kubernetes-application-navigator-kappnav-support)
- [Certificate Manager Support](https://github.com/application-stacks/runtime-component-operator/blob/master/doc/user-guide.adoc#certificate-manager-integration)

For functionality that is unique to the Appsody Application Operator, see the
following sections.


### Operator Configuration

When the operator starts, it creates two `ConfigMap` objects that contain default and constant values for individual stacks in `AppsodyApplication`.

#### Stack defaults

ConfigMap [`appsody-operator-defaults`](../deploy/stack_defaults.yaml) contains the default values for each stack. When users do not provide values inside their `AppsodyApplication` resource, the operator will look up default values inside
this [stack defaults map](../deploy/stack_defaults.yaml).

Input resource:

```yaml
apiVersion: appsody.dev/v1beta1
kind: AppsodyApplication
metadata:
  name: my-appsody-app
spec:
  stack: java-microprofile
  applicationImage: quay.io/my-repo/my-app:1.0
```

Since in the `AppsodyApplication` resource service `port` and `type` are not set, they will be looked up in the default `ConfigMap` and added to the resource. It will be set according to the `stack` field. If the `appsody-operator-defaults` doesn't have the `stack` with a particular name defined then the operator will use `generic` stack's default values.

After defaults are applied:

```yaml
apiVersion: appsody.dev/v1beta1
kind: AppsodyApplication
metadata:
  name: my-appsody-app
spec:
  stack: java-microprofile
  applicationImage: quay.io/my-repo/my-app:1.0
  ....
  service:
    port: 9080
    type: ClusterIP  
```
 
#### Stack Constants ConfigMap

[`appsody-operator-constants`](../deploy/stack_constants.yaml) ConfigMap contains the constant values for each stack. These values will always be used over the ones that users provide. This can be used to limit the user's ability to control certain fields such as `expose`. It also provides the ability to set environment variables that are always required.

Input resource:

```yaml
apiVersion: appsody.dev/v1beta1
kind: AppsodyApplication
metadata:
  name: my-appsody-app
spec:
  stack: java-microprofile
  applicationImage: quay.io/my-repo/my-app:1.0
  expose: true
  env:
  -  name: DB_URL
     value: url
```

After constants are applied:

```yaml
apiVersion: appsody.dev/v1beta1
kind: AppsodyApplication
metadata:
  name: my-appsody-app
spec:
  stack: java-microprofile
  applicationImage: quay.io/my-repo/my-app:1.0
  ....
  expose: false
  env:
  -  name: VENDOR
     value: COMPANY
  -  name: DB_URL
     value: url     
```


### Troubleshooting

See the [troubleshooting guide](troubleshooting.md) for information on how to investigate and resolve deployment problems.
