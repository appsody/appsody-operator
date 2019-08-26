package e2e

import (
	"testing"

	"github.com/appsody-operator/pkg/apis"
	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	t.Run("AppsodyPullPolicyTest", AppsodyPullPolicyTest)
	t.Run("AppsodyBasicTest", AppsodyBasicTest)
	t.Run("AppsodyStorageTest", AppsodyBasicStorageTest)
	t.Run("AppsodyPullSecretTest", AppsodyPullSecretTest)
	t.Run("AppsodyResourcesTest", AppsodyResourcesTest)
	t.Run("AppsodyServiceAccountTest", AppsodyServiceAccountTest)
	t.Run("AppsodyServicesTest", AppsodyServicesTest)
	t.Run("AppsodyVolumesTest", AppsodyVolumesTest)

}
