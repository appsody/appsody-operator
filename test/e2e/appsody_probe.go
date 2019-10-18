package e2e

import (
	goctx "context"
	"testing"
	"time"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/appsody/appsody-operator/test/util"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	corev1 "k8s.io/api/core/v1"
)

// AppsodyProbeTest make sure user defined liveness/readiness probes reach ready state.
func AppsodyProbeTest(t *testing.T) {
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
		util.FailureCleanup(t, f, namespace, err)
	}

	libertyProbe := corev1.Handler{
		HTTPGet: &corev1.HTTPGetAction{
			Path: "/",
			Port: intstr.FromInt(3000),
		},
	}

	// run test for readiness probe and then liveness
	if err = probeTest(t, f, ctx, libertyProbe); err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}
}

func probeTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, probe corev1.Handler) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return err
	}
	// default appsody test now has to define probes manually, so we will use those and change in the edit test.
	exampleAppsody := util.MakeBasicAppsodyApplication(t, f, "example-appsody-readiness", namespace, 1)

	err = f.Client.Create(goctx.TODO(), exampleAppsody, &framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       time.Second * 5,
		RetryInterval: time.Second,
	})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-readiness", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	if err = editProbeTest(t, f, ctx, exampleAppsody); err != nil {
		return err
	}
	return nil
}

func editProbeTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, app *appsodyv1beta1.AppsodyApplication) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return err
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-readiness", Namespace: namespace}, app)
	if err != nil {
		return err
	}

	// Adjust tests for update SMALL amounts to keep the test fast.
	app.Spec.LivenessProbe.InitialDelaySeconds = int32(6)
	app.Spec.ReadinessProbe.InitialDelaySeconds = int32(3)
	err = f.Client.Update(goctx.TODO(), app)
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-readiness", 1, retryInterval, timeout)
	if err != nil {
		return err
	}
	return nil
}
