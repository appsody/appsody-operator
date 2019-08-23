package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	appsodyv1beta1 "github.com/appsody-operator/pkg/apis/appsody/v1beta1"
	"github.com/appsody-operator/test/util"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"k8s.io/apimachinery/pkg/types"
)

// AppsodyBasicTest barebones deployment test that makes sure applications will deploy and scale.
func AppsodyBasicTest(t *testing.T) {
	ctx, err := util.InitializeContext(t, cleanupTimeout, retryInterval)
	defer ctx.Cleanup()
	if err != nil {
		t.Fatal(err)
	}

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("Couldn't get namespace: %v", err)
	}

	t.Logf("Namespace: %s", namespace)

	f := framework.Global

	// create one replica of the operator deployment in current namespace with provided name
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, operatorTimeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = appsodyBasicScaleTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}

func appsodyBasicScaleTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	helper := int32(1)

	exampleAppsody := util.MakeBasicAppsodyApplication(t, f, "example-appsody", namespace, helper)

	timestamp := time.Now().UTC()
	t.Logf("%s - Creating basic appsody application for scaling test...", timestamp)
	// Create application deployment and wait
	err = f.Client.Create(goctx.TODO(), exampleAppsody, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second, RetryInterval: time.Second})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody", 1, retryInterval, timeout)
	if err != nil {
		return err
	}
	// -- Run all scaling tests below based on the above example deployment of 1 pods ---
	// update the number of replicas and return if failure occurs
	if err = appsodyUpdateScaleTest(t, f, namespace, exampleAppsody); err != nil {
		return err
	}
	timestamp = time.Now().UTC()
	t.Logf("%s - Completed basic appsody scale test", timestamp)
	return err
}

func appsodyUpdateScaleTest(t *testing.T, f *framework.Framework, namespace string, exampleAppsody *appsodyv1beta1.AppsodyApplication) error {
	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody", Namespace: namespace}, exampleAppsody)
	if err != nil {
		return err
	}

	helper2 := int32(2)
	exampleAppsody.Spec.Replicas = &helper2
	err = f.Client.Update(goctx.TODO(), exampleAppsody)
	if err != nil {
		return err
	}

	// wait for example-memcached to reach 2 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody", 2, retryInterval, timeout)
	if err != nil {
		return err
	}
	return err
}
