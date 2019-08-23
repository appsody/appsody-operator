# Appsody Application Operator

The Appsody Application Operator can be used to deploy applications created by [Appsody Application Stacks](https://appsody.dev/) into [OKD](https://www.okd.io/) or [OpenShift](https://www.openshift.com/) clusters.

## Operator installation

Use the instructions for one of the [releases](../deploy/releases) to install the operator into a Kubernetes cluster.

## Overview

The architecture of the Appsody Application Operator follows the basic controller pattern:  the Operator container with the controller is deployed into a Pod and listens for incoming resources with `Kind: AppsodyApplication`. Creating a `AppsodyApplication` custom resource (CR) triggers the Appsody Application Operator to create, update or delete Kubernetes resources needed by the application to run on your cluster.

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

## `AppsodyApplication` configuration

### Custom Resource Definition (CRD)

The following table lists configurable parameters of the `AppsodyApplication` CRD. For complete OpenAPI v3 representation of these values please see [`AppsodyApplication` CRD](../deploy/crds/appsody_v1beta1_appsodyapplication_crd.yaml).

Each `AppsodyApplication` CR must specify `applicationImage` and `stack` parameters. Specifying other parameters is optional.

| Parameter | Description |
|---|---|
| `stack` | The name of the Appsody Application Stack that produced this application image. |
| `serviceAccountName` | The name of the OpenShift service account to be used during deployment. |
| `applicationImage` | The absolute name of the image to be deployed, containing the registry and the tag. |
| `pullPolicy` | The policy used when pulling the image.  One of: `Always`, `Never`, and `IfNotPresent`. |
| `pullSecret` | If using a registry that requires authentication, the name of the secret containing credentials. |
| `architecture` | An array of architectures to be considered for deployment. Their position in the array indicates preference. |
| `service.port` | The port exposed by the container. |
| `service.type` | The Kubernetes [Service Type](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types). |
| `service.annotations` | Annotations to be added to the service. |
| `createKnativeService`   | A boolean to toggle the creation of Knative resources and usage of Knative serving. |
| `expose`   | A boolean that toggles the external exposure of this deployment via a Route resource.|
| `replicas` | The static number of desired replica pods that run simultaneously. |
| `autoscaling.maxReplicas` | Required field for autoscaling. Upper limit for the number of pods that can be set by the autoscaler. Cannot be lower than the minimum number of replicas. |
| `autoscaling.minReplicas`   | Lower limit for the number of pods that can be set by the autoscaler. |
| `autoscaling.targetCPUUtilizationPercentage`   | Target average CPU utilization (represented as a percentage of requested CPU) over all the pods. |
| `resourceConstraints.requests.cpu` | The minimum required CPU core. Specify integers, fractions (e.g. 0.5), or millicore values(e.g. 100m, where 100m is equivalent to .1 core). Required field for autoscaling. |
| `resourceConstraints.requests.memory` | The minimum memory in bytes. Specify integers with one of these suffixes: E, P, T, G, M, K, or power-of-two equivalents: Ei, Pi, Ti, Gi, Mi, Ki.|
| `resourceConstraints.limits.cpu` | The upper limit of CPU core. Specify integers, fractions (e.g. 0.5), or millicores values(e.g. 100m, where 100m is equivalent to .1 core). |
| `resourceConstraints.limits.memory` | The memory upper limit in bytes. Specify integers with suffixes: E, P, T, G, M, K, or power-of-two equivalents: Ei, Pi, Ti, Gi, Mi, Ki.|
| `env`   | An array of environment variables following the format of `{name, value}`, where value is a simple string. |
| `envFrom`   | An array of environment variables following the format of `{name, valueFrom}`, where `valueFrom` is YAML object containing a property named either `secretKeyRef` or `configMapKeyRef`, which in turn contain the properties `name` and `key`.|
| `readinessProbe`   | A YAML object configuring the [Kubernetes readiness probe](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/#define-readiness-probes) that controls when the pod is ready to receive traffic. |
| `livenessProbe` | A YAML object configuring the [Kubernetes liveness probe](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/#define-a-liveness-http-request) that controls when Kubernetes needs to restart the pod.|
| `volumes` | A YAML object representing a [pod volume](https://kubernetes.io/docs/concepts/storage/volumes). |
| `volumeMounts` | A YAML object representing a [pod volumeMount](https://kubernetes.io/docs/concepts/storage/volumes/). |
| `storage.size` | A convenient field to set the size of the persisted storage. Can be overriden by the `storage.volumeClaimTemplate` property. |
| `storage.mountPath` | The directory inside the container where this persisted storage will be bound to. |
| `storage.volumeClaimTemplate` | A YAML object representing a [volumeClaimTemplate](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#components) component of a `StatefulSet`. |

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

Both `stack` and `applicationImage` values are required to be defined in an `AppsodyApplication` CR. `stack` should be the same value as the [Appsody application stack](https://github.com/appsody/stacks) you used to created your application.

### `ServiceAccount` configuration

The operator can creates a `ServiceAccount` resource when deploying an Appsody based application. If `serviceAccountName` is not specified in a CR, the operator creates a service account with the same name as the CR (e.g. `my-appsody-app`) in the namespace the CR is created.

Users can also specify `serviceAccountName` when they want to create a service account manually.

If applications require specific permissions but still want the operator to create a `ServiceAccount`, users can still manually create a role binding to bind a role to the service account created by the operator. To learn more about Role-based access control (RBAC), see Kubernetes [documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

### Environment variables

You can set environment variables for your application container. To set environment variables, specify `env` and/or `envFrom` fields in your CR. The environment variables can come directly from key/value pairs, `ConfigMap`s or `Secret`s.

 ```yaml
apiVersion: appsody.dev/v1beta1
kind: AppsodyApplication
metadata:
  name: my-appsody-app
spec:
  stack: java-microprofile
  applicationImage: quay.io/my-repo/my-app:1.0
  env:
    - name: DB_PORT
      value: "6379"
    - name: DB_USERNAME
      valueFrom:
        secretKeyRef:
          name: db-credential
          key: adminUsername
    - name: DB_PASSWORD
      valueFrom:
        secretKeyRef:
          name: db-credential
          key: adminPassword
  envFrom:
    - configMapRef:
        name: env-configmap
    - secretRef:
        name: env-secrets
```

Use `envFrom` to define all data in a `ConfigMap` or a `Secret` as environment variables in a container. Keys from `ConfigMap` or `Secret` resources, become environment variable name in your container.

### Persistence

Appsody Operator is capable of creating a `StatefulSet` and `PersistentVolumeClaim` for each pod if storage is specified in the `AppsodyApplication` CR.

Users also can provide mount points for their application. There are 2 ways to enable storage.

#### Basic Storage

With the `AppsodyApplication` CR definition below the operator will create `PersistentVolumeClaim` called `pvc` with the size of `1Gi` and `ReadWriteOnce` access mode.

Operator will also create a volume mount for the `StatefulSet` mounting to `/data` folder. You can use `volumeMounts` field instead of `storage.mountPath` if you require to persist more then one folder.

```yaml
apiVersion: appsody.dev/v1beta1
kind: AppsodyApplication
metadata:
  name: my-appsody-app
spec:
  stack: java-microprofile
  applicationImage: quay.io/my-repo/my-app:1.0
  storage:
    size: 1Gi
    mountPath: "/data"
```

#### Advanced Storage

Operator allows users to provide entire `volumeClaimTemplate` for full control over automatically created `PersistentVolumeClaim`.

It is also possible to create multiple volume mount points for persistent volume using `volumeMounts` field as shown below. You can still use `storage.mountPath` if you require only a single mount point.

```yaml
apiVersion: appsody.dev/v1beta1
kind: AppsodyApplication
metadata:
  name: my-appsody-app
spec:
  stack: java-microprofile
  applicationImage: quay.io/my-repo/my-app:1.0
  volumeMounts:
  - name: pvc
    mountPath: /data_1
    subPath: data_1
  - name: pvc
    mountPath: /data_2
    subPath: data_2
  storage:
    volumeClaimTemplate:
      metadata:
        name: pvc
      spec:
        accessModes:
        - "ReadWriteMany"
        storageClassName: 'glusterfs'
        resources:
          requests:
            storage: 1Gi
```

### Knative support

Appsody Operator can deploy serverless applications with [Knative](https://knative.dev/docs/) on a Kubernetes cluster. To achieve this, the operator creates a [Knative `Service`](https://github.com/knative/serving/blob/master/docs/spec/spec.md#service) resource which manages the whole life cycle of a workload.

To create `Knative Service`, set `createKnativeService` to `true`:

```yaml
apiVersion: appsody.dev/v1beta1
kind: AppsodyApplication
metadata:
  name: my-appsody-app
spec:
  stack: java-microprofile
  applicationImage: quay.io/my-repo/my-app:1.0
  createKnativeService: true
```

By setting this parameter, the operator creates a `Knative Service` in the cluster and populates the resource with applicable `AppsodyApplication` CRD fields. Also it ensures non-Knative resources including Kubernetes `Service`, `Route`, `Deployment` and etc. are deleted.

The CRD fields that are used to populate the `Knative Service` resource includes `applicationImage`, `serviceAccountName`, `livenessProbe`, `readinessProbe`, `service.Port`, `volumes`, `volumeMounts`, `env`, `envFrom`, `pullSecret` and `pullPolicy`.

For more details on how to configure Knative for tasks such as enabling HTTPS connections and setting up a custom domain, checkout [Knative Documentation](https://knative.dev/docs/serving/).

_This feature is only available if you have Knative installed on your cluster._

### Exposing service externally

To expose your application externally, set `expose` to `true`:

```yaml
apiVersion: appsody.dev/v1beta1
kind: AppsodyApplication
metadata:
  name: my-appsody-app
spec:
  stack: java-microprofile
  applicationImage: quay.io/my-repo/my-app:1.0
  expose: true
```

By setting this parameter, the operator creates an unsecured route based on your application service. Setting this parameter is the same as running `oc expose service <service-name>`.

To create a secured HTTPS route, see [secured routes](https://docs.openshift.com/container-platform/3.11/architecture/networking/routes.html#secured-routes) for more information.

_This feature is only available if you are running on OKD or OpenShift._

### Troubleshooting

See the [troubleshooting guide](troubleshooting.md) for information on how to investigate and resolve deployment problems.
