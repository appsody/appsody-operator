package e2e

import (
	goctx "context"
	"errors"
	"testing"
	"time"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	"github.com/appsody/appsody-operator/test/util"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	k "sigs.k8s.io/controller-runtime/pkg/client"
)

// AppsodyAutoScalingTest : More indepth testing of autoscaling
func AppsodyAutoScalingTest(t *testing.T) {

	ctx, err := util.InitializeContext(t, cleanupTimeout, retryInterval)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Cleanup()

	f := framework.Global
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}

	// Wait for the operator as the following configmaps won't exist until it has deployed
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, operatorTimeout)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	timestamp := time.Now().UTC()
	t.Logf("%s - Starting appsody autoscaling test...", timestamp)

	// create one replica of the operator deployment in current namespace with provided name
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, operatorTimeout)
	if err != nil {
		t.Fatal(err)
	}

	// Make basic appsody application with 1 replica
	replicas := int32(1)
	appsodyApplication := util.MakeBasicAppsodyApplication(t, f, "example-appsody-autoscaling", namespace, replicas)

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), appsodyApplication, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// wait for example-appsody-autoscaling to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-autoscaling", 1, retryInterval, timeout)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// Check the name field that matches
	m := map[string]string{"metadata.name": "example-appsody-autoscaling"}
	l := fields.Set(m)
	selec := l.AsSelector()

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-autoscaling", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	appsodyApplication.Spec.ResourceConstraints = setResources("0.2")
	appsodyApplication.Spec.Autoscaling = setAutoScale(5, 50)

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	hpa := &autoscalingv1.HorizontalPodAutoscalerList{}
	options := k.ListOptions{FieldSelector: selec}
	hpa = getHPA(hpa, t, f, options)

	timestamp = time.Now().UTC()
	t.Logf("%s - Deployment created, verifying autoscaling...", timestamp)

	err = waitForHPA(hpa, t, 1, 5, 50, f, options)
	if err != nil {
		t.Fatal(err)
	}

	updateTest(t, f, appsodyApplication, options, namespace, hpa)
	minMaxTest(t, f, appsodyApplication, options, namespace, hpa)
	minBoundaryTest(t, f, appsodyApplication, options, namespace, hpa)
	incorrectFieldsTest(t, f, ctx)
}

func getHPA(hpa *autoscalingv1.HorizontalPodAutoscalerList, t *testing.T, f *framework.Framework, options k.ListOptions) *autoscalingv1.HorizontalPodAutoscalerList {
	if err := f.Client.List(goctx.TODO(), hpa, &options); err != nil {
		t.Logf("Get HPA: (%v)", err)
	}
	return hpa
}

func waitForHPA(hpa *autoscalingv1.HorizontalPodAutoscalerList, t *testing.T, minReplicas int32, maxReplicas int32, utiliz int32, f *framework.Framework, options k.ListOptions) error {
	for counter := 0; counter < 6; counter++ {
		time.Sleep(4000 * time.Millisecond)
		hpa = getHPA(hpa, t, f, options)
		if checkValues(hpa, t, minReplicas, maxReplicas, utiliz) == nil {
			return nil
		}
	}
	return checkValues(hpa, t, minReplicas, maxReplicas, utiliz)

}

func setResources(cpu string) *corev1.ResourceRequirements {
	cpuRequest := resource.MustParse(cpu)

	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU: cpuRequest,
		},
	}
}

func setAutoScale(values ...int32) *appsodyv1beta1.AppsodyApplicationAutoScaling {
	if len(values) == 3 {
		return &appsodyv1beta1.AppsodyApplicationAutoScaling{
			TargetCPUUtilizationPercentage: &values[2],
			MaxReplicas:                    values[0],
			MinReplicas:                    &values[1],
		}
	} else if len(values) == 2 {
		return &appsodyv1beta1.AppsodyApplicationAutoScaling{
			TargetCPUUtilizationPercentage: &values[1],
			MaxReplicas:                    values[0],
		}
	}

	return &appsodyv1beta1.AppsodyApplicationAutoScaling{}

}

func checkValues(hpa *autoscalingv1.HorizontalPodAutoscalerList, t *testing.T, minReplicas int32, maxReplicas int32, utiliz int32) error {

	if hpa.Items[0].Spec.MaxReplicas != maxReplicas {
		t.Logf("Max replicas is set to: %d", hpa.Items[0].Spec.MaxReplicas)
		return errors.New("Error: Max replicas is not correctly set")
	}

	if *hpa.Items[0].Spec.MinReplicas != minReplicas {
		t.Logf("Min replicas is set to: %d", *hpa.Items[0].Spec.MinReplicas)
		return errors.New("Error: Min replicas is not correctly set")
	}

	if *hpa.Items[0].Spec.TargetCPUUtilizationPercentage != utiliz {
		t.Logf("TargetCPUUtilization is set to: %d", *hpa.Items[0].Spec.TargetCPUUtilizationPercentage)
		return errors.New("Error: TargetCPUUtilizationis is not correctly set")
	}

	return nil
}

// Updates the values and checks they are changed
func updateTest(t *testing.T, f *framework.Framework, appsodyApplication *appsodyv1beta1.AppsodyApplication, options k.ListOptions, namespace string, hpa *autoscalingv1.HorizontalPodAutoscalerList) {

	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-autoscaling", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	appsodyApplication.Spec.ResourceConstraints = setResources("0.2")
	appsodyApplication.Spec.Autoscaling = setAutoScale(3, 2, 30)

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	timestamp := time.Now().UTC()
	t.Logf("%s - Deployment created, verifying autoscaling...", timestamp)

	hpa = getHPA(hpa, t, f, options)

	err = waitForHPA(hpa, t, 2, 3, 30, f, options)
	if err != nil {
		t.Fatal(err)
	}
}

// Checks when max is less than min, there should be no update
func minMaxTest(t *testing.T, f *framework.Framework, appsodyApplication *appsodyv1beta1.AppsodyApplication, options k.ListOptions, namespace string, hpa *autoscalingv1.HorizontalPodAutoscalerList) {

	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-autoscaling", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	appsodyApplication.Spec.ResourceConstraints = setResources("0.2")
	appsodyApplication.Spec.Autoscaling = setAutoScale(1, 6, 10)

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	timestamp := time.Now().UTC()
	t.Logf("%s - Deployment created, verifying autoscaling...", timestamp)

	hpa = getHPA(hpa, t, f, options)

	err = waitForHPA(hpa, t, 2, 3, 30, f, options)
	if err != nil {
		t.Fatal(err)
	}
}

// When min is set to less than 1, there should be no update since the minReplicas are updated to a value less than 1
func minBoundaryTest(t *testing.T, f *framework.Framework, appsodyApplication *appsodyv1beta1.AppsodyApplication, options k.ListOptions, namespace string, hpa *autoscalingv1.HorizontalPodAutoscalerList) {

	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-autoscaling", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	appsodyApplication.Spec.ResourceConstraints = setResources("0.5")
	appsodyApplication.Spec.Autoscaling = setAutoScale(4, 0, 20)

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	timestamp := time.Now().UTC()
	t.Logf("%s - Deployment created, verifying autoscaling...", timestamp)

	hpa = getHPA(hpa, t, f, options)

	err = waitForHPA(hpa, t, 2, 3, 30, f, options)
	if err != nil {
		t.Fatal(err)
	}
}

// When the mandatory fields for autoscaling are not set
func incorrectFieldsTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}

	timestamp := time.Now().UTC()
	t.Logf("%s - Starting appsody autoscaling test...", timestamp)

	// Make basic appsody application with 1 replica
	replicas := int32(1)
	appsodyApplication := util.MakeBasicAppsodyApplication(t, f, "example-appsody-autoscaling2", namespace, replicas)

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), appsodyApplication, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// wait for example-appsody-autoscaling to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-autoscaling2", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	// Check the name field that matches
	m := map[string]string{"metadata.name": "example-appsody-autoscaling2"}
	l := fields.Set(m)
	selec := l.AsSelector()

	options := k.ListOptions{FieldSelector: selec}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-autoscaling2", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	appsodyApplication.Spec.ResourceConstraints = setResources("0.3")
	appsodyApplication.Spec.Autoscaling = setAutoScale(4)

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	timestamp = time.Now().UTC()
	t.Logf("%s - Deployment created, verifying autoscaling...", timestamp)

	hpa := &autoscalingv1.HorizontalPodAutoscalerList{}
	hpa = getHPA(hpa, t, f, options)

	if len(hpa.Items) == 0 {
		t.Log("The mandatory fields were not set so autoscaling is not enabled")
	} else {
		t.Fatal("Error: The mandatory fields were not set so autoscaling should not be enabled")
	}
}
