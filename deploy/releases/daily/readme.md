# Appsody Operator

## Installation

The Appsody Operator can be installed to:

- watch its own namespace
- watch another namespace
- watch all namespaces in the cluster

Appropriate cluster role and binding are required to watch another namespace or to watch all namespaces.

_Limitation: Operator cannot be installed to watch multiple namespaces_

---

1. Install `AppsodyApplication` Custom Resource Definition (CRD). This needs to be done only ONCE per cluster:

    ```console
    kubectl apply -f https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/daily/appsody-app-crd.yaml
    ```

2. Install the Appsody Operator:

    **Important: In Step 2.1, ensure that you replace  `<SPECIFY_OPERATOR_NAMESPACE_HERE>` and `<SPECIFY_WATCH_NAMESPACE_HERE>` with proper values:**

    2.1. Set operator namespace and the namespace to watch:

    - To watch all namespaces in the cluster, set `WATCH_NAMESPACE='""'`

    ```console
    OPERATOR_NAMESPACE=`<SPECIFY_OPERATOR_NAMESPACE_HERE>`
    WATCH_NAMESPACE=`<SPECIFY_WATCH_NAMESPACE_HERE>`
    ```

    2.2. _Optional_: Install cluster-level role based access. This step can be skipped if the operator is only watching its own namespace:
    ```
    curl -L https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/daily/appsody-app-cluster-rbac.yaml \
      | sed -e "s/APPSODY_OPERATOR_NAMESPACE/${OPERATOR_NAMESPACE}/" \
      | kubectl apply -f -
    ```

    2.3. Install the operator:

    ```console
    curl -L https://raw.githubusercontent.com/appsody/appsody-operator/master/deploy/releases/daily/appsody-app-operator.yaml \
      | sed -e "s/APPSODY_WATCH_NAMESPACE/${WATCH_NAMESPACE}/" \
      | kubectl apply -n ${OPERATOR_NAMESPACE} -f -
    ```

## Uninstallation

To uninstall the operator, run commands from Step 2.3 first and then Step 2.2 (if applicable), but after replacing `kubectl apply` with `kubectl delete`.

To delete the CRD, run command from Step 1, but after replacing `kubectl apply` with `kubectl delete`.

_Deleting the CRD will also delete all `AppsodyApplication` in the cluster_

## Current Limitations

- Operator cannot be installed to watch multiple namespaces
- Knative support is limited. Values specified for `autoscaling`, `resources` and `replicas` parameters would not apply for Knative, when enabled using `createKnativeService` parameter.