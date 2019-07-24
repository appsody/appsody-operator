package util

import (
	"testing"

	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MakeBasicAppsodyApplication : Create a simple Appsody App with provided number of replicas.
func MakeBasicAppsodyApplication(t *testing.T, f *framework.Framework, ns string, replicas int32) *appsodyv1alpha1.AppsodyApplication {
	return &appsodyv1alpha1.AppsodyApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AppsodyApplication",
			APIVersion: "appsody.example.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-appsody",
			Namespace: ns,
		},
		Spec: appsodyv1alpha1.AppsodyApplicationSpec{
			ApplicationImage: "appsody:v1",
			Replicas:         &replicas,
			Service: appsodyv1alpha1.AppsodyApplicationService{
				Port: 8000,
			},
		},
	}
}
