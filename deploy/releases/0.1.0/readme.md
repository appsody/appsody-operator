# Appsody Application Operator

## Operator Installation

You can install the Appsody Application Operator by running the following `kubectl` commands:

```bash
kubectl apply -f https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/0.1.0/appsody-app-crd.yaml
kubectl apply -f https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/0.1.0/appsody-app-operator.yaml
```

## Current Limitations:

- Knative support is limited. Values specified for `autoscaling`, `resources` and `replicas` parameters would not apply for Knative, when enabled using `createKnativeService` parameter.