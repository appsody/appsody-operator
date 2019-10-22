package e2e

import (
	"testing"

	"github.com/appsody/appsody-operator/pkg/apis"
	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAppsodyApplication(t *testing.T) {
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
	t.Run("AppsodyServiceMonitorTest", AppsodyServiceMonitorTest)
	t.Run("AppsodyConfigMapsDefaultTest", AppsodyConfigMapsDefaultTest)
	t.Run("AppsodyConfigMapsConstTest", AppsodyConfigMapsConstTest)
	t.Run("AppsodyKnativeTest", AppsodyKNativeTest)
}
