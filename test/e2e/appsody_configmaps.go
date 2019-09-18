package e2e

import (
	goctx "context"
	"testing"
	"time"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	"github.com/appsody/appsody-operator/test/util"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	k "sigs.k8s.io/controller-runtime/pkg/client"
)

// AppsodyConfigMapsTest : More indepth testing of configmap
func AppsodyConfigMapsTest(t *testing.T) {

	ctx, err := util.InitializeContext(t, cleanupTimeout, retryInterval)
	defer ctx.Cleanup()
	if err != nil {
		t.Fatal(err)
	}

	f := framework.Global
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}
	timestamp := time.Now().UTC()
	t.Logf("%s - Starting appsody configmap test...", timestamp)

	// Make basic appsody application with 1 replica
	replicas := int32(1)
	appsodyApplication := &appsodyv1beta1.AppsodyApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AppsodyApplication",
			APIVersion: "appsody.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-appsody-configmap",
			Namespace: namespace,
		},
		Spec: appsodyv1beta1.AppsodyApplicationSpec{
			ApplicationImage: "navidsh/demo-day",
			Replicas:         &replicas,
			Stack:            "nodejs-express",
		},
	}

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), appsodyApplication, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// wait for example-appsody-configmap to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-configmap", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	m := map[string]string{"metadata.name": "example-appsody-configmap"}
	l := fields.Set(m)
	selec := l.AsSelector()

	apps := &appsodyv1beta1.AppsodyApplicationList{}
	options := k.ListOptions{FieldSelector: selec}

	if err := f.Client.List(goctx.TODO(), &options, apps); err != nil {
		t.Log(err)
	}

	if apps.Items[0].Spec.LivenessProbe.FailureThreshold != 3 && apps.Items[0].Spec.LivenessProbe.InitialDelaySeconds != 60 && apps.Items[0].Spec.LivenessProbe.PeriodSeconds != 5 {
		t.Log("Wrong default values for the LivenessProbe")
		t.Fail()
	} else {
		t.Log("Correct default values for the LivenessProbe")
	}

	if apps.Items[0].Spec.ReadinessProbe.FailureThreshold != 12 && apps.Items[0].Spec.ReadinessProbe.InitialDelaySeconds != 30 && apps.Items[0].Spec.ReadinessProbe.PeriodSeconds != 5 {
		t.Log("Wrong default values for the ReadinessProbe")
		t.Fail()
	} else {
		t.Log("Correct default values for the ReadinessProbe")
	}

	memoryRequest := resource.MustParse("256Mi")
	if apps.Items[0].Spec.ResourceConstraints.Requests.Memory().String() != (&memoryRequest).String() {
		t.Log("Wrong default values for Memory")
		t.Fail()
	} else {
		t.Log("Correct default values for Memory")
	}

	serviceType := corev1.ServiceTypeClusterIP
	if apps.Items[0].Spec.Service.Port != 3000 && apps.Items[0].Spec.Service.Type != &serviceType {
		t.Log("Wrong default values for the Service")
		t.Fail()
	} else {
		t.Log("Correct default values for the Service")
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-configmap", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Log(err)
	}

	serviceType = corev1.ServiceTypeNodePort
	appsodyApplication.Spec.LivenessProbe = setProbe(1, 1, 1, 1, 1)
	appsodyApplication.Spec.ReadinessProbe = setProbe(1, 1, 1, 1, 1)
	appsodyApplication.Spec.ResourceConstraints = setResourceConstraints("512Mi")
	appsodyApplication.Spec.Service = setService(8080, &serviceType)

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Log(err)
	}

	if err := f.Client.List(goctx.TODO(), &options, apps); err != nil {
		t.Log(err)
	}

	if apps.Items[0].Spec.LivenessProbe.FailureThreshold != 1 && apps.Items[0].Spec.LivenessProbe.SuccessThreshold != 1 && apps.Items[0].Spec.LivenessProbe.InitialDelaySeconds != 1 && apps.Items[0].Spec.LivenessProbe.PeriodSeconds != 1 && apps.Items[0].Spec.LivenessProbe.TimeoutSeconds != 1 {
		t.Log("LivenessProbe values not updated")
		t.Fail()
	} else {
		t.Log("LivenessProbe values are successfully updated")
	}

	if apps.Items[0].Spec.ReadinessProbe.FailureThreshold != 1 && apps.Items[0].Spec.ReadinessProbe.SuccessThreshold != 1 && apps.Items[0].Spec.ReadinessProbe.InitialDelaySeconds != 1 && apps.Items[0].Spec.ReadinessProbe.PeriodSeconds != 1 && apps.Items[0].Spec.ReadinessProbe.TimeoutSeconds != 1 {
		t.Log("ReadinessProbe values not updated")
		t.Fail()
	} else {
		t.Log("ReadinessProbe values are successfully updated")
	}

	oldLimit := apps.Items[0].Spec.ResourceConstraints.Requests.Memory()
	newLimit := resource.MustParse("512Mi")
	if *oldLimit != newLimit {
		t.Log("ResourceConstraints not properly updated")
		t.Fail()
	} else {
		t.Log("ResourceConstraints are successfully updated")
	}

	if apps.Items[0].Spec.Service.Port != 8080 && apps.Items[0].Spec.Service.Type != &serviceType {
		t.Log("Services are not properly updated")
		t.Fail()
	} else {
		t.Log("Services are successfully updated")
	}

	//Check the name field that matches // "metadata.name": "appsody-operator-defaults"
	// m = map[string]string{"metadata.name": "appsody-operator-constants"}
	// l = fields.Set(m)
	// selec = l.AsSelector()

	// maps := &corev1.ConfigMapList{}
	// options = k.ListOptions{FieldSelector: selec}

	// if err := f.Client.List(goctx.TODO(), &options, maps); err != nil {
	// 	t.Log(err)
	// }

	// for i := 0; i < len(maps.Items); i++ {
	// 	t.Log(maps.Items[i])
	// 	t.Log("----------------------------------------------------------------------")
	// }

	t.Fail()

}

func setProbe(initialDelay int32, timeoutSeconds int32, periodSeconds int32, successThreshold int32, failureThreshold int32) *corev1.Probe {

	probe := corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: "/",
			Port: intstr.FromInt(3000),
		},
	}

	return &corev1.Probe{
		Handler:             probe,
		InitialDelaySeconds: initialDelay,
		TimeoutSeconds:      timeoutSeconds,
		PeriodSeconds:       periodSeconds,
		SuccessThreshold:    successThreshold,
		FailureThreshold:    failureThreshold,
	}
}

func setResourceConstraints(memory string) *corev1.ResourceRequirements {
	memoryRequest := resource.MustParse(memory)

	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: memoryRequest,
		},
	}
}

func setService(port int32, serviceType *corev1.ServiceType) *appsodyv1beta1.AppsodyApplicationService {
	return &appsodyv1beta1.AppsodyApplicationService{
		Port: port,
		Type: serviceType,
	}
}
