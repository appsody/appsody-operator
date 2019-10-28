package util

import (
	goctx "context"
	"io/ioutil"
	"testing"
	"time"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
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
				InitialDelaySeconds: 1,
				TimeoutSeconds:      1,
				PeriodSeconds:       5,
				SuccessThreshold:    1,
				FailureThreshold:    16,
			},
			LivenessProbe: &corev1.Probe{
				Handler:             probe,
				InitialDelaySeconds: 4,
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

		if int(statefulset.Status.ReadyReplicas) == replicas {
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

// InitializeContext : Sets up initial context
func InitializeContext(t *testing.T, clean, retryInterval time.Duration) (*framework.TestCtx, error) {
	ctx := framework.NewTestCtx(t)
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       clean,
		RetryInterval: retryInterval,
	})
	if err != nil {
		if ctx != nil {
			ctx.Cleanup()
		}
		return nil, err
	}

	t.Log("Cluster context initialized.")
	return ctx, nil
}

// ResetConfigMap : Resets the configmaps to original empty values, this is required to allow tests to be run after the configmaps test
func ResetConfigMap(t *testing.T, f *framework.Framework, configMap *corev1.ConfigMap, cmName string, fileName string, namespace string) {
	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: cmName, Namespace: namespace}, configMap)
	if err != nil {
		t.Fatal(err)
	}
	fData, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}

	configMap = &corev1.ConfigMap{}
	err = yaml.Unmarshal(fData, configMap)
	if err != nil {
		t.Fatal(err)
	}
	configMap.Namespace = namespace
	err = f.Client.Update(goctx.TODO(), configMap)
	if err != nil {
		t.Fatal(err)
	}
}

// FailureCleanup : Log current state of the namespace and exit with fatal
func FailureCleanup(t *testing.T, f *framework.Framework, ns string, failure error) {
	t.Log("***** FAILURE")
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

	t.Fatal(failure)
}

// WaitForKnativeDeployment : Poll for ksvc creation when createKnativeService is set to true
func WaitForKnativeDeployment(t *testing.T, f *framework.Framework, ns, n string, retryInterval, timeout time.Duration) error {
	// add to scheme to framework can find the resource
	err := servingv1alpha1.AddToScheme(f.Scheme)
	if err != nil {
		return err
	}

	err = wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		ksvc := &servingv1alpha1.ServiceList{}
		lerr := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: n, Namespace: ns}, ksvc)
		if lerr != nil {
			if apierrors.IsNotFound(lerr) {
				t.Logf("Waiting for knative service %s...", n)
				return false, nil
			}
			// issue retrieving ksvc
			return false, lerr
		}

		t.Logf("Found knative service %s", n)
		return true, nil
	})
	return err
}
