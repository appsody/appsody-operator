package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	"github.com/appsody-operator/pkg/apis"
	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Minute * 3
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
		t.Fatal(err)
	}

	t.Logf("Namespace: %s", namespace)

	f := framework.Global

	// create one replica of the operator deployment in current namespace with provided name
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, timeout)
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

	helper := int32(3)
	exampleAppsody := &appsodyv1alpha1.AppsodyApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AppsodyApplication",
			APIVersion: "appsody.example.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-appsody",
			Namespace: namespace,
		},
		Spec: appsodyv1alpha1.AppsodyApplicationSpec{
			ApplicationImage: "appsody:v1",
			Replicas:         &helper,
			Service: appsodyv1alpha1.AppsodyApplicationService{
				Port: 8000,
			},
		},
	}
	err = f.Client.Create(goctx.TODO(), exampleAppsody, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	if err != nil {
		return err
	}

	return err
}
