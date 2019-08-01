# Appsody Application Operator

## Operator Installation

You can install the Appsody Application Operator by running the following `kubectl` commands:

```bash
kubectl apply -f https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/daily/appsody-app-dependencies.yaml
kubectl apply -f https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/daily/appsody-app-operator.yaml
```

## Current Limitations:

- The ConfigMap is specified in JSON format
- Knative support is limited. Values specified for `autoscaling` and `replicas` parameters would not apply for Knative, when enabled using `createKnativeService` parameter.