package e2e

import (
	goctx "context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	k "sigs.k8s.io/controller-runtime/pkg/client"

	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	"github.com/appsody-operator/test/util"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"

	corev1 "k8s.io/api/core/v1"
)

// AppsodyVolumesTest checks that the configured volume is applied to the pods
func AppsodyVolumesTest(t *testing.T) {

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
	t.Logf("%s - Starting appsody volumes test...", timestamp)

	replicas := int32(2)
	volumes := make([]corev1.Volume, 1)
	volumeMounts := make([]corev1.VolumeMount, 1)
	entry1 := corev1.Volume{Name: "my-volume"}
	entry2 := corev1.VolumeMount{Name: "my-volume", MountPath: "/vol"}

	volumes[0] = entry1
	volumeMounts[0] = entry2

	appsodyApplication := util.MakeBasicAppsodyApplication(t, f, "example-appsody-volumes", namespace, replicas)

	appsodyApplication.Spec.Volumes = volumes
	appsodyApplication.Spec.VolumeMounts = volumeMounts

	m := map[string]string{"app.kubernetes.io/name": "example-appsody-volumes"}
	l := labels.Set(m)

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), appsodyApplication, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// wait for example-appsody-volumes to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-volumes", 2, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	timestamp = time.Now().UTC()
	t.Logf("%s - Deployment created, verifying volumes...", timestamp)

	if err = verifyVolumes(t, f, appsodyApplication, l); err != nil {
		t.Fatal(err)
	}

}

func verifyVolumes(t *testing.T, f *framework.Framework, app *appsodyv1alpha1.AppsodyApplication, l labels.Set) error {
	pods := &corev1.PodList{}
	selec := l.AsSelector()
	options := k.ListOptions{LabelSelector: selec}
	f.Client.List(goctx.TODO(), &options, pods)

	for i := 0; i < len(pods.Items); i++ {
		t.Log(pods.Items[i].Spec.Containers[0].VolumeMounts[0])
		if pods.Items[i].Spec.Containers[0].VolumeMounts[0].Name == "my-volume" && pods.Items[i].Spec.Containers[0].VolumeMounts[0].MountPath == "/vol" && pods.Items[i].Spec.Containers[0].VolumeMounts[0].ReadOnly == false {
		} else {
			t.Log("The volume and volume mount is not set")
			t.Fail()
		}
	}
	return nil
}
