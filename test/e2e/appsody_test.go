package e2e

import (
	"os"
	"testing"

	"github.com/appsody/appsody-operator/pkg/apis"
	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAppsodyApplication(t *testing.T) {
	cluster := os.Getenv("CLUSTER_ENV")
	t.Logf("running e2e tests as '%s'", cluster)

	appsodyApplicationList := &appsodyv1beta1.AppsodyApplicationList{
		TypeMeta: metav1.TypeMeta{
			Kind: "AppsodyApplication",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, appsodyApplicationList)
	if err != nil {
		t.Fatalf("Failed to add CR scheme to framework: %v", err)
	}

	t.Run("AppsodyPullPolicyTest", AppsodyPullPolicyTest)
	t.Run("AppsodyBasicTest", AppsodyBasicTest)
	t.Run("AppsodyStorageTest", AppsodyBasicStorageTest)
	t.Run("AppsodyPersistenceTest", AppsodyPersistenceTest)
	t.Run("AppsodyProbeTest", AppsodyProbeTest)
	t.Run("AppsodyAutoScalingTest", AppsodyAutoScalingTest)
	t.Run("AppsodyConfigMapsDefaultTest", AppsodyConfigMapsDefaultTest)
	t.Run("AppsodyConfigMapsConstTest", AppsodyConfigMapsConstTest)


	if cluster != "local" {
		// only test non-OCP features on minikube
		if cluster == "minikube" {
			testIndependantFeatures(t)
			return
		}

		// test all features that require some configuration
		testAdvancedFeatures(t)

		// test features that require OCP
		if cluster == "ocp" {
			testOCPFeatures(t)
		}
	}
}
func testAdvancedFeatures(t *testing.T) {
	// These features require a bit of configuration
	// which makes them less ideal for quick minikube tests
	t.Run("AppsodyServiceMonitorTest", AppsodyServiceMonitorTest)
	t.Run("AppsodyKnativeTest", AppsodyKnativeTest)
	t.Run("AppsodyServiceBindingTest", AppsodyServiceBindingTest)
	t.Run("AppsodyCertManagerTest", AppsodyCertManagerTest)
}

// Verify functionality that is tied to OCP
func testOCPFeatures(t *testing.T) {
	t.Run("AppsodyImageStreamTest", AppsodyImageStreamTest)
}

// Verify functionality that is not expected to run on OCP
func testIndependantFeatures(t *testing.T) {
	// TODO: implement test for ingress
}
