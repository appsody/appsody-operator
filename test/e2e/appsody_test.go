package e2e

import (
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	appsody "github.com/appsody/appsody-operator/pkg/apis/appsody/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAppsodyApplication(t *testing.T) {
	appsodyApplicationList := &appsody.AppsodyApplicationList{
		TypeMeta: metav1.TypeMeta {
			Kind: "AppsodyApplication",
			APIVersion: "appsody/v1alpha1",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, appsodyApplicationList)
	if err != nil {
		t.Fatlf("Failed to add CR scheme to framework: %v", err)
	}

	// t.Run("SimpleTest", )
}


// --- Test Functions ----

func AppsodyBasicTest(t *testing.T, applicationTag string) {
}
