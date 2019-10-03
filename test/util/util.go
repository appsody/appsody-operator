package util

import (
	goctx "context"
	"testing"
	"time"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// MakeBasicAppsodyApplication : Create a simple Appsody App with provided number of replicas.
func MakeBasicAppsodyApplication(t *testing.T, f *framework.Framework, n string, ns string, replicas int32) *appsodyv1beta1.AppsodyApplication {
	probe := corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: "/",
			Port: intstr.FromInt(3000),
		},
	}
	expose := false
	serviceType := corev1.ServiceTypeClusterIP
	return &appsodyv1beta1.AppsodyApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AppsodyApplication",
			APIVersion: "appsody.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: ns,
		},
		Spec: appsodyv1beta1.AppsodyApplicationSpec{
			ApplicationImage: "navidsh/demo-day",
			Replicas:         &replicas,
			Expose:           &expose,
			Service: &appsodyv1beta1.AppsodyApplicationService{
				Port: 3000,
				Type: &serviceType,
			},
			ReadinessProbe: &corev1.Probe{
				Handler:             probe,
				InitialDelaySeconds: 1, // minor adjustment
				TimeoutSeconds:      1,
				PeriodSeconds:       5,
				SuccessThreshold:    1,
				FailureThreshold:    16,
			},
			LivenessProbe: &corev1.Probe{
				Handler:             probe,
				InitialDelaySeconds: 4, // minor adjustment
				TimeoutSeconds:      1,
				PeriodSeconds:       5,
				SuccessThreshold:    1,
				FailureThreshold:    6,
			},
			Stack: "nodejs-express",
		},
	}
}

// WaitForStatefulSet : Identical to WaitForDeployment but for StatefulSets.
func WaitForStatefulSet(t *testing.T, kc kubernetes.Interface, ns, n string, replicas int, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		statefulset, err := kc.AppsV1().StatefulSets(ns).Get(n, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s statefulset\n", n)
				return false, nil
			}
			return false, err
		}

		if int(statefulset.Status.CurrentReplicas) == replicas {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s statefulset (%d/%d)\n", n, statefulset.Status.CurrentReplicas, replicas)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("StatefulSet available (%d/%d)\n", replicas, replicas)
	return nil
}

func InitializeContext(t *testing.T, clean, retryInterval time.Duration) (*framework.TestCtx, error) {
	ctx := framework.NewTestCtx(t)
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       clean,
		RetryInterval: retryInterval,
	})
	if err != nil {
		return nil, err
	}

	t.Log("Cluster context initialized.")
	return ctx, nil
}

func FailureCleanup(t *testing.T, f *framework.Framework, ns string) {
	options := &dynclient.ListOptions{
		Namespace: ns,
	}
	podlist := &corev1.PodList{}
	err := f.Client.List(goctx.TODO(), options, podlist)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("***** Logging pods in namespace: %s", ns)
	for _, p := range podlist.Items {
		t.Log("--------------------")
		t.Log(p)
	}

	crlist := &appsodyv1beta1.AppsodyApplicationList{}
	err = f.Client.List(goctx.TODO(), options, crlist)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("***** Logging Appsody Applications in namespace: %s", ns)
	for _, application := range crlist.Items {
		t.Log("-------------------")
		t.Log(application)
	}
}
