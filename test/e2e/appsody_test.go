package e2e

import (
	"testing"
	"time"

	"github.com/appsody-operator/pkg/apis"
	appsody "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Minute * 3
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestAppsodyApplication(t *testing.T) {
	appsodyApplicationList := &appsody.AppsodyApplicationList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AppsodyApplication",
			APIVersion: "appsody/v1alpha1",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, appsodyApplicationList)
	if err != nil {
		t.Fatalf("Failed to add CR scheme to framework: %v", err)
	}

	// t.Run("SimpleTest", )
}

// --- Test Functions ----

func AppsodyBasicTest(t *testing.T, applicationTag string) {
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

	t.Log("Cluster Resource Initialized")

	namespace, err := ctx.GetNamespace()

	if err != nil {
		t.Fatal(err)
	}
}
