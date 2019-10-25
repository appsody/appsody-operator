package e2e

import (
	goctx "context"
	"testing"
	"time"

	"github.com/appsody/appsody-operator/test/util"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// AppsodyKnativeTest : Create application with knative service enabled to verify feature
func AppsodyKnativeTest(t *testing.T) {
	ctx, err := util.InitializeContext(t, cleanupTimeout, retryInterval)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Cleanup()
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("Couldn't get namespace: %v", err)
	}

	t.Logf("Namespace: %s", namespace)

	f := framework.Global

	if isKnativeInstalled(t, f) {
		err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, operatorTimeout)
		if err != nil {
			util.FailureCleanup(t, f, namespace, err)
		}
		knativeBool := true
		applicationName := "example-appsody-knative"

		exampleAppsody := util.MakeBasicAppsodyApplication(t, f, applicationName, namespace, 1)
		exampleAppsody.Spec.CreateKnativeService = &knativeBool

		// Create application deployment and wait
		err = f.Client.Create(goctx.TODO(), exampleAppsody, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second, RetryInterval: time.Second})
		if err != nil {
			util.FailureCleanup(t, f, namespace, err)
		}

		err = util.WaitForKnativeDeployment(t, f, namespace, applicationName, retryInterval, timeout)
		if err != nil {
			util.FailureCleanup(t, f, namespace, err)
		}
	} else {
		t.Log("Knative is not installed on this cluster, skipping AppsodyKnativeTest...")
	}
}

func isKnativeInstalled(t *testing.T, f *framework.Framework) bool {
	deployments := &corev1.PodList{}
	options := &dynclient.ListOptions{
		Namespace: "knative-serving",
	}
	err := f.Client.List(goctx.TODO(), options, deployments)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false
		}
		t.Fatalf("Error occurred while trying to find knative-serving %v", err)
	}
	return true
}
