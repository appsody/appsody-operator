package tests

import (
	goctx "context"
	"errors"
	"testing"
	"time"

	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	"github.com/appsody-operator/test/util"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppsodyServiceAccountTest checks that the configured pull policy is applied to deployment
func AppsodyServiceAccountTest(t *testing.T) {

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
	t.Logf("%s - Starting appsody namespace test...", timestamp)

	replicas := int32(1)

	appsodyApplication := util.MakeBasicAppsodyApplication(t, f, "example-appsody-serviceaccount", namespace, replicas)

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), appsodyApplication, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// wait for example-appsody-namespace to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-serviceaccount", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	timestamp = time.Now().UTC()
	t.Logf("%s - Deployment created, verifying service account name...", timestamp)

	if err = verifyServiceAccount(t, f, appsodyApplication); err != nil {
		t.Fatal(err)
	}
}

func verifyServiceAccount(t *testing.T, f *framework.Framework, app *appsodyv1alpha1.AppsodyApplication) error {
	name := app.ObjectMeta.Name
	ns := app.ObjectMeta.Namespace

	deploy, err := f.KubeClient.AppsV1().Deployments(ns).Get(name, metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		t.Logf("Got error when getting NameSpace %s: %s", name, err)
		return err
	}

	if deploy.Spec.Template.Spec.ServiceAccountName != "example-appsody-serviceaccount" {
		return errors.New("ServiceAccountName is incorrect")
	}
	return nil
}
