# Appsody Application Operator

## Operator Installation

You can install the Appsody Application Operator by running the following `kubectl` commands:

```bash
kubectl apply -f https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/daily/appsody-app-crd.yaml
kubectl apply -f https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/daily/appsody-app-operator.yaml
```

In above `appsody-app-operator.yaml`, the `WATCH_NAMESPACE` environment variable is set to watch all namespaces in the cluster. If you want to watch a specific namespace, then set that namespace (`jane` for example) as following:

```yaml
          env:
            - name: WATCH_NAMESPACE
              value: jane
```

## Current Limitations:

- Knative support is limited. Values specified for `autoscaling`, `resources` and `replicas` parameters would not apply for Knative, when enabled using `createKnativeService` parameter.