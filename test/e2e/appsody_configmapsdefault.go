package e2e

import (
	goctx "context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/intstr"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	"github.com/appsody/appsody-operator/test/util"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// AppsodyConfigMapsDefaultTest : More indepth testing of configmap
func AppsodyConfigMapsDefaultTest(t *testing.T) {

	ctx, err := util.InitializeContext(t, cleanupTimeout, retryInterval)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Cleanup()

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("Couldn't get namespace: %v", err)
	}
	t.Logf("Namespace: %s", namespace)
	f := framework.Global

	// Values to be loaded into the default configmap
	updateData := map[string]string{"jstack": `{"version": 1.0.0,"expose":true, "service":{"port": 3000,"type": NodePort, "annotations":{"prometheus.io/scrape": true}}, "readinessProbe":{"failureThreshold": 12, "httpGet":{"path": /ready, "port": 3000}, "initialDelaySeconds": 5, "periodSeconds": 2, "timeoutSeconds": 1}, "livenessProbe":{"failureThreshold": 12, "httpGet":{"path": /live, "port": 3000}, "initialDelaySeconds": 5, "periodSeconds": 2}}`}
	configMap := &corev1.ConfigMap{}

	// Wait for the operator as the following configmaps won't exist until it has deployed
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, operatorTimeout)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// Gets the configmap that contains the default values that will be applied to unspecified fields in the appsody application
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "appsody-operator-defaults", Namespace: namespace}, configMap)
	if err != nil {
		t.Fatal(err)
	}

	// Sets default values
	configMap.Data = updateData

	err = f.Client.Update(goctx.TODO(), configMap)
	if err != nil {
		t.Fatal(err)
	}

	// Creating a basic appsody application that does not specify fields
	replicas := int32(1)
	apps := &appsodyv1beta1.AppsodyApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AppsodyApplication",
			APIVersion: "appsody.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-appsody-defaultconfigmaps",
			Namespace: namespace,
		},
		Spec: appsodyv1beta1.AppsodyApplicationSpec{
			ApplicationImage: "navidsh/demo-day",
			Replicas:         &replicas,
			Stack:            "jstack",
		},
	}

	err = f.Client.Create(goctx.TODO(), apps, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second, RetryInterval: time.Second})
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// wait for example-appsody-defaultconfigmaps to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-defaultconfigmaps", 1, retryInterval, timeout)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-defaultconfigmaps", Namespace: namespace}, apps)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// update the application to two replicas
	helper := int32(2)
	apps.Spec.Replicas = &helper

	err = f.Client.Update(goctx.TODO(), apps)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-defaultconfigmaps", 2, retryInterval, timeout)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// check that the default values from the default configmap have been applied to the fields that were not specified
	if *apps.Spec.Expose == true {
		t.Log("Expose in configmap defaults is applied")
	} else {
		t.Fatal("Expose in configmap defaults is not applied")
	}

	serviceType := corev1.ServiceTypeNodePort
	if apps.Spec.Service.Port == 3000 && *apps.Spec.Service.Type == serviceType && apps.Spec.Service.Annotations != nil {
		t.Log("Service in configmap defaults is applied")
	} else {
		t.Fatal("Service in configmap defaults is not applied")
	}

	port := intstr.IntOrString{IntVal: 3000}
	if apps.Spec.ReadinessProbe.FailureThreshold == 12 && apps.Spec.ReadinessProbe.HTTPGet.Path == "/ready" && apps.Spec.ReadinessProbe.HTTPGet.Port == port && apps.Spec.ReadinessProbe.InitialDelaySeconds == 5 && apps.Spec.ReadinessProbe.PeriodSeconds == 2 && apps.Spec.ReadinessProbe.TimeoutSeconds == 1 {
		t.Log("ReadinessProbe in configmap defaults is applied")
	} else {
		t.Fatal("ReadinessProbe in configmap defaults is not applied")
	}

	if apps.Spec.LivenessProbe.FailureThreshold == 12 && apps.Spec.LivenessProbe.HTTPGet.Path == "/live" && apps.Spec.LivenessProbe.HTTPGet.Port == port && apps.Spec.LivenessProbe.InitialDelaySeconds == 5 && apps.Spec.LivenessProbe.PeriodSeconds == 2 {
		t.Log("LivenessProbe in configmap defaults is applied")
	} else {
		t.Fatal("LivenessProbe in configmap defaults is not applied")
	}
}
