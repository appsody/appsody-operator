package e2e

import (
	goctx "context"
	"testing"
	"time"

	appsodyv1beta1 "github.com/appsody-operator/pkg/apis/appsody/v1beta1"
	"github.com/appsody-operator/test/util"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	k "sigs.k8s.io/controller-runtime/pkg/client"
)

// AppsodyPullSecretTest checks that the configured pull policy is applied to deployment
func AppsodyPullSecretTest(t *testing.T) {

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
	t.Logf("%s - Starting appsody pull policy test...", timestamp)

	replicas := int32(1)
	secret := "pullSecret"

	appsodyApplication := util.MakeBasicAppsodyApplication(t, f, "example-appsody-pullsecret", namespace, replicas)
	appsodyApplication.Spec.PullSecret = &secret

	m := map[string]string{"app.kubernetes.io/name": "example-appsody-pullsecret"}
	l := labels.Set(m)

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), appsodyApplication, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// wait for example-appsody-pullpolicy to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-pullsecret", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	timestamp = time.Now().UTC()
	t.Logf("%s - Deployment created, verifying pull secret...", timestamp)

	if err = verifyPullSecret(t, f, appsodyApplication, l, secret); err != nil {
		t.Fatal(err)
	}
}

func verifyPullSecret(t *testing.T, f *framework.Framework, app *appsodyv1beta1.AppsodyApplication, l labels.Set, secret string) error {
	pods := &corev1.PodList{}
	selec := l.AsSelector()
	options := k.ListOptions{LabelSelector: selec}
	f.Client.List(goctx.TODO(), &options, pods)

	for i := 0; i < len(pods.Items); i++ {
		if pods.Items[i].Spec.ImagePullSecrets[0].Name == secret {
			t.Log("Successfully updated the pull secret")
		} else {
			t.Log("Pull Secret was not updated")
			t.Fail()
		}
	}
	return nil
}
