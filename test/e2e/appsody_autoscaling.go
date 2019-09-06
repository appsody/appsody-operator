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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	k "sigs.k8s.io/controller-runtime/pkg/client"
)

// AppsodyAutoScalingTest : More indepth testing of autoscaling
func AppsodyAutoScalingTest(t *testing.T) {

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
	t.Logf("%s - Starting appsody autoscaling test...", timestamp)

	// Make basic appsody application with 1 replica
	replicas := int32(1)
	appsodyApplication := util.MakeBasicAppsodyApplication(t, f, "example-appsody-autoscaling", namespace, replicas)

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), appsodyApplication, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// wait for example-appsody-autoscaling to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-autoscaling", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	// Check the name field that matches
	m := map[string]string{"metadata.name": "example-appsody-autoscaling"}
	l := fields.Set(m)
	selec := l.AsSelector()

	apps := &appsodyv1beta1.AppsodyApplicationList{}
	options := k.ListOptions{FieldSelector: selec}

	apps = getAppsodyApplicationList(apps, t, f, options)

	// Get last time the appsodyApplication resource was updated
	updateTime := apps.Items[0].Status.Conditions[0].LastUpdateTime

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-autoscaling", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Log(err)
	}

	appsodyApplication.Spec.ResourceConstraints = setResources("1")
	appsodyApplication.Spec.Autoscaling = setAutoScale(6, 50)

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Log(err)
	}

	waitForHPA(t, f, options, *apps, updateTime)

	timestamp = time.Now().UTC()
	t.Logf("%s - Deployment created, verifying autoscaling...", timestamp)

	hpa := &autoscalingv1.HorizontalPodAutoscalerList{}
	options2 := k.ListOptions{FieldSelector: selec}
	hpa = getHPA(hpa, t, f, options2)

	updateTest(t, f, appsodyApplication, apps, options, namespace, updateTime, hpa, options2)
	minMaxTest(t, f, appsodyApplication, apps, options, namespace, updateTime, hpa, options2)
	minBoundaryTest(t, f, appsodyApplication, apps, options, namespace, updateTime, hpa, options2)
	incorrectFieldsTest(t, f)
}

func getAppsodyApplicationList(apps *appsodyv1beta1.AppsodyApplicationList, t *testing.T, f *framework.Framework, options k.ListOptions) *appsodyv1beta1.AppsodyApplicationList {
	if err := f.Client.List(goctx.TODO(), &options, apps); err != nil {
		t.Log(err)
	}
	return apps
}

func getHPA(hpa *autoscalingv1.HorizontalPodAutoscalerList, t *testing.T, f *framework.Framework, options2 k.ListOptions) *autoscalingv1.HorizontalPodAutoscalerList {
	if err := f.Client.List(goctx.TODO(), &options2, hpa); err != nil {
		t.Logf("Get HPA: (%v)", err)
	}
	return hpa
}

func waitForHPA(t *testing.T, f *framework.Framework, options k.ListOptions, apps appsodyv1beta1.AppsodyApplicationList, updateTime v1.Time) {
	for {
		if err := f.Client.List(goctx.TODO(), &options, &apps); err != nil {
			t.Log(err)
		}
		if updateTime != apps.Items[0].Status.Conditions[0].LastUpdateTime {
			break
		}
	}
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
	if hpa.Items[0].Spec.MaxReplicas == maxReplicas && *hpa.Items[0].Spec.MinReplicas == minReplicas && *hpa.Items[0].Spec.TargetCPUUtilizationPercentage == utiliz {
		return nil
	}
	return errors.New("Values are notsuccessfully set")
}

// Updates the values and checks they are changed
func updateTest(t *testing.T, f *framework.Framework, appsodyApplication *appsodyv1beta1.AppsodyApplication, apps *appsodyv1beta1.AppsodyApplicationList, options k.ListOptions, namespace string, updateTime v1.Time, hpa *autoscalingv1.HorizontalPodAutoscalerList, options2 k.ListOptions) {

	apps = getAppsodyApplicationList(apps, t, f, options)

	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-autoscaling", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Log(err)
	}

	appsodyApplication.Spec.ResourceConstraints = setResources("0.2")
	appsodyApplication.Spec.Autoscaling = setAutoScale(5, 4, 30)

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Log(err)
	}

	waitForHPA(t, f, options, *apps, updateTime)

	timestamp := time.Now().UTC()
	t.Logf("%s - Deployment created, verifying autoscaling...", timestamp)

	hpa = getHPA(hpa, t, f, options2)

	err = checkValues(hpa, t, 4, 5, 30)
	if err != nil {
		t.Log("Error: There should be an update since the values have been updated.")
		t.Fail()
	}
	t.Log("Values updated to new values successfully")
}

// Check default value of MinReplicas
func minMaxTest(t *testing.T, f *framework.Framework, appsodyApplication *appsodyv1beta1.AppsodyApplication, apps *appsodyv1beta1.AppsodyApplicationList, options k.ListOptions, namespace string, updateTime v1.Time, hpa *autoscalingv1.HorizontalPodAutoscalerList, options2 k.ListOptions) {

	apps = getAppsodyApplicationList(apps, t, f, options)

	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-autoscaling", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Log(err)
	}

	appsodyApplication.Spec.ResourceConstraints = setResources("0.2")
	appsodyApplication.Spec.Autoscaling = setAutoScale(1, 6, 10)

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Log(err)
	}

	waitForHPA(t, f, options, *apps, updateTime)

	timestamp := time.Now().UTC()
	t.Logf("%s - Deployment created, verifying autoscaling...", timestamp)

	hpa = getHPA(hpa, t, f, options2)

	err = checkValues(hpa, t, 4, 5, 30)
	if err != nil {
		t.Log("Error: There should be no update since the minReplicas are greater than the maxReplicas")
		t.Fail()
	}
	t.Log("There is no update, due to the minReplicas being greater than the maxReplicas. The values remain the same")

}

// When min is set to less than 1
func minBoundaryTest(t *testing.T, f *framework.Framework, appsodyApplication *appsodyv1beta1.AppsodyApplication, apps *appsodyv1beta1.AppsodyApplicationList, options k.ListOptions, namespace string, updateTime v1.Time, hpa *autoscalingv1.HorizontalPodAutoscalerList, options2 k.ListOptions) {

	apps = getAppsodyApplicationList(apps, t, f, options)

	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-autoscaling", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Log(err)
	}

	appsodyApplication.Spec.ResourceConstraints = setResources("0.5")
	appsodyApplication.Spec.Autoscaling = setAutoScale(4, 0, 20)

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Log(err)
	}

	waitForHPA(t, f, options, *apps, updateTime)

	timestamp := time.Now().UTC()
	t.Logf("%s - Deployment created, verifying autoscaling...", timestamp)

	hpa = getHPA(hpa, t, f, options2)

	err = checkValues(hpa, t, 4, 5, 30)
	if err != nil {
		t.Log("Error: There should be no update since the minReplicas are updated to a value less than 1")
		t.Fail()
	}
	t.Log("There is no update, due to the minReplicas being less than 1. The values remain the same")
}

// When the mandatory fields for autoscaling are not set
func incorrectFieldsTest(t *testing.T, f *framework.Framework) {

	ctx, err := util.InitializeContext(t, cleanupTimeout, retryInterval)
	defer ctx.Cleanup()
	if err != nil {
		t.Fatal(err)
	}

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

	apps := &appsodyv1beta1.AppsodyApplicationList{}
	options := k.ListOptions{FieldSelector: selec}

	apps = getAppsodyApplicationList(apps, t, f, options)

	// Get last time the appsodyApplication resource was updated
	updateTime := apps.Items[0].Status.Conditions[0].LastUpdateTime

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-autoscaling2", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Log(err)
	}

	appsodyApplication.Spec.ResourceConstraints = setResources("1")
	appsodyApplication.Spec.Autoscaling = setAutoScale(6)

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Log(err)
	}

	waitForHPA(t, f, options, *apps, updateTime)

	timestamp = time.Now().UTC()
	t.Logf("%s - Deployment created, verifying autoscaling...", timestamp)

	hpa := &autoscalingv1.HorizontalPodAutoscalerList{}
	options2 := k.ListOptions{FieldSelector: selec}
	hpa = getHPA(hpa, t, f, options2)

	if len(hpa.Items) == 0 {
		t.Log("The mandatory fields were not set so autoscaling is not enabled")
	} else {
		t.Log("Error: The mandatory fields were not set so autoscaling should not be enabled")
		t.Fail()
	}
}
