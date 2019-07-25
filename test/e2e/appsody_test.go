package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	"github.com/appsody-operator/pkg/apis"
	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	"github.com/appsody-operator/test/util"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	retryInterval        = time.Second * 5
	operatorTimeout      = time.Minute * 3
	timeout              = time.Second * 30
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestAppsodyApplication(t *testing.T) {
	appsodyApplicationList := &appsodyv1alpha1.AppsodyApplicationList{
		TypeMeta: metav1.TypeMeta{
			Kind: "AppsodyApplication",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, appsodyApplicationList)
	if err != nil {
		t.Fatalf("Failed to add CR scheme to framework: %v", err)
	}

	t.Run("AppsodyBasicTest", appsodyBasicTest)
}

// --- Test Functions ----

func appsodyBasicTest(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       cleanupTimeout,
		RetryInterval: retryInterval,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Cluster resource intialized.")

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
	// if err = appsodyBasicStorageTest(t, f, ctx); err != nil {
	// 	t.Fatal(err)
	// }
}

func appsodyBasicStorageTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	exampleAppsody := util.MakeBasicAppsodyApplication(t, f, "example-appsody-storage", namespace, 1)
	exampleAppsody.Spec.Storage = &appsodyv1alpha1.AppsodyApplicationStorage{
		Size:      "10Mi",
		MountPath: "/mnt/data",
	}

	err = f.Client.Create(goctx.TODO(), exampleAppsody, &framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       time.Second * 5,
		RetryInterval: time.Second * 1,
	})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-storage", 1, retryInterval, timeout)
	return err
}

func appsodyBasicScaleTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	helper := int32(3)

	exampleAppsody := util.MakeBasicAppsodyApplication(t, f, "example-appsody", namespace, helper)

	// Create application deployment and wait
	err = f.Client.Create(goctx.TODO(), exampleAppsody, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody", 3, retryInterval, timeout)
	if err != nil {
		return err
	}
	// -- Run all scaling tests below based on the above example deployment of 3 pods ---
	// update the number of replicas and return if failure occurs
	if err = appsodyUpdateScaleTest(t, f, namespace, exampleAppsody); err != nil {
		return err
	}

	return err
}

func appsodyUpdateScaleTest(t *testing.T, f *framework.Framework, namespace string, exampleAppsody *appsodyv1alpha1.AppsodyApplication) error {
	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody", Namespace: namespace}, exampleAppsody)
	if err != nil {
		return err
	}

	helper2 := int32(4)
	exampleAppsody.Spec.Replicas = &helper2
	err = f.Client.Update(goctx.TODO(), exampleAppsody)
	if err != nil {
		return err
	}

	// wait for example-memcached to reach 4 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody", 4, retryInterval, timeout)
	if err != nil {
		return err
	}
	return err
}
