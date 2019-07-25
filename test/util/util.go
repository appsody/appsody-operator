package util

import (
	"bytes"
	"io"
	"testing"

	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MakeBasicAppsodyApplication : Create a simple Appsody App with provided number of replicas.
func MakeBasicAppsodyApplication(t *testing.T, f *framework.Framework, n string, ns string, replicas int32) *appsodyv1alpha1.AppsodyApplication {
	return &appsodyv1alpha1.AppsodyApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AppsodyApplication",
			APIVersion: "appsody.example.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: ns,
		},
		Spec: appsodyv1alpha1.AppsodyApplicationSpec{
			ApplicationImage: "openliberty/open-liberty:javaee8-ubi-min",
			Replicas:         &replicas,
			Service: appsodyv1alpha1.AppsodyApplicationService{
				Port: 9080,
			},
		},
	}
}

// GetLogs returns the logs from the given pod (in the server's namespace).
func GetLogs(f *framework.Framework, app *appsodyv1alpha1.AppsodyApplication, podName string) (string, error) {
	logsReq := f.KubeClient.CoreV1().Pods(app.ObjectMeta.Namespace).GetLogs(podName, &corev1.PodLogOptions{})
	podLogs, err := logsReq.Stream()
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	logs := buf.String()
	return logs, nil
}
