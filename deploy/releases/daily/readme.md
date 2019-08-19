# Appsody Application Operator

## Operator Installation

Install the Appsody Application Operator in your cluster by following these steps:

**Important: In Step 1, ensure that you replace  `<SPECIFY_OPERATOR_NAMESPACE_HERE>` and `<SPECIFY_WATCH_NAMESPACE_HERE>` with proper values:**

1. Set operator namespace and the namespace to watch:

   - To watch all namespaces in the cluster, set `WATCH_NAMESPACE="''"`

    ```console
    $ OPERATOR_NAMESPACE=`<SPECIFY_OPERATOR_NAMESPACE_HERE>`
    $ WATCH_NAMESPACE=`<SPECIFY_WATCH_NAMESPACE_HERE>`
    ```

2. Install Custom Resource Definition (CRD) and operator:

    ```console
    $ kubectl apply -f https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/daily/appsody-app-crd.yaml

    $ curl -L https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/daily/appsody-app-operator.yaml | sed -e "s/APPSODY_OPERATOR_NAMESPACE/$OPERATOR_NAMESPACE/" -e "s/APPSODY_WATCH_NAMESPACE/$WATCH_NAMESPACE/" | kubectl apply -f -
    ```

## Current Limitations

- Knative support is limited. Values specified for `autoscaling`, `resources` and `replicas` parameters would not apply for Knative, when enabled using `createKnativeService` parameter.