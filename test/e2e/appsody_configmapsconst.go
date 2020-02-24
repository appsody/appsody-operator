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

// AppsodyConfigMapsConstTest : More indepth testing of configmap
func AppsodyConfigMapsConstTest(t *testing.T) {

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

	// create one replica of the operator deployment in current namespace with provided name
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, operatorTimeout)
	if err != nil {
		t.Fatal(err)
	}

	// Values to be loaded in the constants configmap
	updateData := map[string]string{"jstack": `{"version": 1.0.0,"expose":true, "service":{"port": 3000,"targetPort": 8080,"type": NodePort}, "livenessProbe":{"failureThreshold": 8, "httpGet":{"path": /live, "port": 3000}, "initialDelaySeconds": 8, "periodSeconds": 2}, "readinessProbe":{"failureThreshold": 12, "httpGet":{"path": /ready, "port": 3000}, "initialDelaySeconds": 5, "periodSeconds": 2, "timeoutSeconds": 1}}`}
	configMap := &corev1.ConfigMap{}

	// Wait for the operator as the following configmaps won't exist until it has deployed
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, operatorTimeout)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// Gets the configmap that contains the constant values that will be applied to the appsody application and cannot be changed
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "appsody-operator-constants", Namespace: namespace}, configMap)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// Sets constant values
	configMap.Data = updateData

	err = f.Client.Update(goctx.TODO(), configMap)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// Creating a basic appsody application that specifies new values for fields that are already assigned from the constants configmap
	replicas := int32(1)
	probe := corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: "/health",
			Port: intstr.FromInt(3000),
		},
	}
	expose := false
	serviceType := corev1.ServiceTypeClusterIP
	apps := &appsodyv1beta1.AppsodyApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AppsodyApplication",
			APIVersion: "appsody.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-appsody-constconfigmaps",
			Namespace: namespace,
		},
		Spec: appsodyv1beta1.AppsodyApplicationSpec{
			ApplicationImage: "navidsh/demo-day",
			Replicas:         &replicas,
			Stack:            "jstack",
			Expose:           &expose,
			Service: &appsodyv1beta1.AppsodyApplicationService{
				Type: &serviceType,
			},
			LivenessProbe: &corev1.Probe{
				InitialDelaySeconds: 4,
				PeriodSeconds:       5,
				FailureThreshold:    6,
				Handler:             probe,
			},
			ReadinessProbe: &corev1.Probe{
				InitialDelaySeconds: 4,
				PeriodSeconds:       5,
				FailureThreshold:    6,
				TimeoutSeconds:      2,
				Handler:             probe,
			},
		},
	}

	err = f.Client.Create(goctx.TODO(), apps, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second, RetryInterval: time.Second})
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// wait for example-appsody-constconfigmaps to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-constconfigmaps", 1, retryInterval, timeout)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// creates a new struct for the appsody application otherwise the probe values get overwritten
	apps = &appsodyv1beta1.AppsodyApplication{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-constconfigmaps", Namespace: namespace}, apps)
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

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-constconfigmaps", 2, retryInterval, timeout)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// checks none of the values that were specified in the constants configmap are changed
	if *apps.Spec.Expose == true {
		t.Log("Expose in configmap constants is applied and not changed")
	} else {
		t.Fatal("Expose in configmap constants is not applied")
	}

	serviceType = corev1.ServiceTypeNodePort
	if apps.Spec.Service.Port == 3000 && *apps.Spec.Service.TargetPort == int32(8080) && *apps.Spec.Service.Type == serviceType {
		t.Log("Service from configmap constants is applied and not changed")
	} else {
		t.Fatal("Service in configmap constants is not applied")
	}

	port := intstr.IntOrString{IntVal: 3000}
	if apps.Spec.ReadinessProbe.FailureThreshold == 12 && apps.Spec.ReadinessProbe.HTTPGet.Path == "/ready" && apps.Spec.ReadinessProbe.HTTPGet.Port == port && apps.Spec.ReadinessProbe.InitialDelaySeconds == 5 && apps.Spec.ReadinessProbe.PeriodSeconds == 2 && apps.Spec.ReadinessProbe.TimeoutSeconds == 1 {
		t.Log("ReadinessProbe in configmap constants is applied and not changed")
	} else {
		t.Fatal("ReadinessProbe in configmap constants is not applied")
	}

	if apps.Spec.LivenessProbe.FailureThreshold == 8 && apps.Spec.LivenessProbe.HTTPGet.Path == "/live" && apps.Spec.LivenessProbe.HTTPGet.Port == port && apps.Spec.LivenessProbe.InitialDelaySeconds == 8 && apps.Spec.LivenessProbe.PeriodSeconds == 2 {
		t.Log("LivenessProbe in configmap constants is applied and not changed")
	} else {
		t.Fatal("LivenessProbe in configmap constants is not applied")
	}

	util.ResetConfigMap(t, f, configMap, "appsody-operator-constants", "deploy/stack_constants.yaml", namespace)

}
