# Appsody Application Operator

## Operator Installation

You can install the Appsody Application Operator by running the following `kubectl` commands:

**Important: In the following commands, make sure you replace  `<SPECIFY_OPERATOR_NAMESPACE_HERE>` and `<SPECIFY_WATCH_NAMESPACE_HERE>` with proper values:**

```console
$ kubectl apply -f https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/daily/appsody-app-crd.yaml

$ OPERATOR_NAMESPACE=`<SPECIFY_OPERATOR_NAMESPACE_HERE>`
$ WATCH_NAMESPACE=`<SPECIFY_WATCH_NAMESPACE_HERE>`

$ curl -L https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/daily/appsody-app-operator.yaml | sed -e 's/APPSODY_OPERATOR_NAMESPACE/$OPERATOR_NAMESPACE/' -e 's/APPSODY_WATCH_NAMESPACE/$WATCH_NAMESPACE/' | kubectl apply -f -
```

## Current Limitations:

- Knative support is limited. Values specified for `autoscaling`, `resources` and `replicas` parameters would not apply for Knative, when enabled using `createKnativeService` parameter.