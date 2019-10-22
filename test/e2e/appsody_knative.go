package e2e

import (
	goctx "context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/appsody/appsody-operator/test/util"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// AppsodyKNativeTest verify functionality of kNative option in appsody
func AppsodyKNativeTest(t *testing.T) {
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

	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, operatorTimeout)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}
	knativeBool := true

	exampleAppsody := util.MakeBasicAppsodyApplication(t, f, "example-appsody-knative", namespace, 1)
	exampleAppsody.Spec.CreateKnativeService = &knativeBool

	// Create application deployment and wait
	err = f.Client.Create(goctx.TODO(), exampleAppsody, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second, RetryInterval: time.Second})
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	err = verifyKnativeDeployment(t, f, namespace, "example-appsody-knative", retryInterval, timeout)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}
}

func verifyKnativeDeployment(t *testing.T, f *framework.Framework, ns, n string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		options := &dynclient.ListOptions{
			Namespace: ns,
		}

		serviceList := &corev1.ServiceList{}
		listError := f.Client.List(goctx.TODO(), options, serviceList)
		if listError != nil {
			return true, err
		}
		// verify that the three extra services were created by knative
		services := 0
		for _, svc := range serviceList.Items {
			matched, failure := regexp.MatchString(n+"*", svc.GetName())
			if failure != nil {
				return true, failure
			}
			if matched {
				services++
			}
		}
		if services <= 1 {
			return true, errors.New("Could not find knative services")
		}
		return true, nil
	})
	return err
}
