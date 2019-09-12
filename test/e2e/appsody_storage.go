package e2e

import (
	goctx "context"
	"testing"
	"time"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	"github.com/appsody/appsody-operator/test/util"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	retryInterval        = time.Second * 5
	operatorTimeout      = time.Minute * 3
	timeout              = time.Minute * 20
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

// AppsodyBasicStorageTest check that when persistence is configured that a statefulset is deployed
func AppsodyBasicStorageTest(t *testing.T) {
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

	// create one replica of the operator deployment in current namespace with provided name
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, operatorTimeout)
	if err != nil {
		t.Fatal(err)
	}

	exampleAppsody := util.MakeBasicAppsodyApplication(t, f, "example-appsody-storage", namespace, 1)
	exampleAppsody.Spec.Storage = &appsodyv1beta1.AppsodyApplicationStorage{
		Size:      "10Mi",
		MountPath: "/mnt/data",
	}

	err = f.Client.Create(goctx.TODO(), exampleAppsody, &framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       time.Second * 5,
		RetryInterval: time.Second * 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = util.WaitForStatefulSet(t, f.KubeClient, namespace, "example-appsody-storage", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	// verify that removing the storage config returns it to a deployment not a stateful set
	if err = updateStorageConfig(t, f, ctx, exampleAppsody); err != nil {
		t.Fatal(err)
	}
}

func updateStorageConfig(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, app *appsodyv1beta1.AppsodyApplication) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return err
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: app.Name, Namespace: namespace}, app)
	if err != nil {
		return err
	}
	// remove storage definition to return it to a deployment
	app.Spec.Storage = nil
	app.Spec.VolumeMounts = nil

	err = f.Client.Update(goctx.TODO(), app)
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, app.Name, 1, retryInterval, timeout)
	if err != nil {
		return err
	}
	return nil
}

// AppsodyPersistenceTest Verify the volume persistence claims.
func AppsodyPersistenceTest(t *testing.T) {
	ctx, err := util.InitializeContext(t, cleanupTimeout, retryInterval)
	defer ctx.Cleanup()
	if err != nil {
		t.Fatal(err)
	}

	f := framework.Global

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	RequestLimits := map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceStorage: resource.MustParse("1Gi"),
	}

	// Create PVC and mount for our statefulset.
	exampleAppsody := util.MakeBasicAppsodyApplication(t, f, "example-appsody-persistence", namespace, 1)
	exampleAppsody.Spec.Storage = &appsodyv1beta1.AppsodyApplicationStorage{
		VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
			metav1.TypeMeta{},
			metav1.ObjectMeta{
				Name: "pvc",
			},
			corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.ResourceRequirements{
					Requests: RequestLimits,
				},
			},
			corev1.PersistentVolumeClaimStatus{},
		},
	}
	exampleAppsody.Spec.VolumeMounts = []corev1.VolumeMount{corev1.VolumeMount{
		Name:      "pvc",
		MountPath: "/data",
	}}

	err = f.Client.Create(goctx.TODO(), exampleAppsody, &framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       cleanupTimeout,
		RetryInterval: cleanupRetryInterval,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = util.WaitForStatefulSet(t, f.KubeClient, namespace, "example-appsody-persistence", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	// again remove the storage configuration and see that it deploys correctly.
	if err = updateStorageConfig(t, f, ctx, exampleAppsody); err != nil {
		t.Fatal(err)
	}
}
