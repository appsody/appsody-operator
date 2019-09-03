package e2e

import (
	goctx "context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	k "sigs.k8s.io/controller-runtime/pkg/client"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	"github.com/appsody/appsody-operator/test/util"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"

	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

// AppsodyResourcesTest checks that the configured volume is applied to the pods
func AppsodyResourcesTest(t *testing.T) {

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
	t.Logf("%s - Starting appsody resource constraints test...", timestamp)

	replicas := int32(1)

	cpuLimit := resource.MustParse("2")
	memoryLimit := resource.MustParse("600Mi")
	cpuRequest := resource.MustParse("1")
	memoryRequest := resource.MustParse("300Mi")

	resources := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    cpuLimit,
			corev1.ResourceMemory: memoryLimit,
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    cpuRequest,
			corev1.ResourceMemory: memoryRequest,
		},
	}

	appsodyApplication := util.MakeBasicAppsodyApplication(t, f, "example-appsody-resources", namespace, replicas)

	appsodyApplication.Spec.ResourceConstraints = &resources

	m := map[string]string{"app.kubernetes.io/name": "example-appsody-resources"}
	l := labels.Set(m)

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), appsodyApplication, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// wait for example-appsody-volumes to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-resources", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	timestamp = time.Now().UTC()
	t.Logf("%s - Deployment created, verifying resources...", timestamp)

	if err = verifyResources(t, f, appsodyApplication, l, cpuLimit, memoryLimit, cpuRequest, memoryRequest); err != nil {
		t.Fatal(err)
	}

}

func verifyResources(t *testing.T, f *framework.Framework, app *appsodyv1beta1.AppsodyApplication, l labels.Set, cpuLimit resource.Quantity, memoryLimit resource.Quantity, cpuRequest resource.Quantity, memoryRequest resource.Quantity) error {
	pods := &corev1.PodList{}
	selec := l.AsSelector()
	options := k.ListOptions{LabelSelector: selec}
	f.Client.List(goctx.TODO(), &options, pods)

	for i := 0; i < len(pods.Items); i++ {
		if *pods.Items[i].Spec.Containers[0].Resources.Limits.Cpu() == cpuLimit && *pods.Items[i].Spec.Containers[0].Resources.Limits.Memory() == memoryLimit && *pods.Items[i].Spec.Containers[0].Resources.Requests.Cpu() == cpuRequest && *pods.Items[i].Spec.Containers[0].Resources.Requests.Memory() == memoryRequest {
			t.Log("Successfully updated resources")
		} else {
			t.Log("Resources were not updated")
			t.Fail()
		}
	}
	return nil
}
