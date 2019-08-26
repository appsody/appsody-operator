package e2e

import (
	goctx "context"
	"testing"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"

	k "sigs.k8s.io/controller-runtime/pkg/client"

	appsodyv1beta1 "github.com/appsody-operator/pkg/apis/appsody/v1beta1"
	"github.com/appsody-operator/test/util"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"

	corev1 "k8s.io/api/core/v1"
)

// AppsodyServicesTest checks that the configured volume is applied to the pods
func AppsodyServicesTest(t *testing.T) {

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

	replicas := int32(2)

	serviceType := corev1.ServiceTypeClusterIP

	service := appsodyv1beta1.AppsodyApplicationService{
		Port: 9080,
		Type: &serviceType,
	}

	expose := true

	appsodyApplication := util.MakeBasicAppsodyApplication(t, f, "example-appsody-services", namespace, replicas)

	appsodyApplication.Spec.Service = &service
	appsodyApplication.Spec.Expose = &expose

	m := map[string]string{"app.kubernetes.io/name": "example-appsody-services"}
	l := labels.Set(m)

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), appsodyApplication, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}

	// wait for example-appsody-volumes to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-services", 2, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = verifyClusterIP(t, f, appsodyApplication, l); err != nil {
		t.Fatal(err)
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-services", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	serviceType = corev1.ServiceTypeNodePort
	appsodyApplication.Spec.Service.Type = &serviceType

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-services", 2, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = verifyNodePort(t, f, appsodyApplication, l); err != nil {
		t.Fatal(err)
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-services", Namespace: namespace}, appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	serviceType = corev1.ServiceTypeLoadBalancer
	appsodyApplication.Spec.Service.Type = &serviceType

	err = f.Client.Update(goctx.TODO(), appsodyApplication)
	if err != nil {
		t.Fatal(err)
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-services", 2, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = verifyLoadBalancer(t, f, appsodyApplication, l); err != nil {
		t.Fatal(err)
	}

}

func verifyClusterIP(t *testing.T, f *framework.Framework, app *appsodyv1beta1.AppsodyApplication, l labels.Set) error {
	services := &corev1.ServiceList{}
	selec := l.AsSelector()
	options := k.ListOptions{LabelSelector: selec}
	f.Client.List(goctx.TODO(), &options, services)

	for i := 0; i < len(services.Items); i++ {
		if services.Items[i].Spec.Ports[0].NodePort == 0 && services.Items[i].Spec.Type == corev1.ServiceTypeClusterIP {
			t.Log(services.Items[i])
			t.Log("Successfully set service as ClusterIP")
		} else {
			t.Log("Failed to set service as ClusterIP")
			t.Fail()
		}
	}
	return nil
}

func verifyNodePort(t *testing.T, f *framework.Framework, app *appsodyv1beta1.AppsodyApplication, l labels.Set) error {
	services := &corev1.ServiceList{}
	selec := l.AsSelector()
	options := k.ListOptions{LabelSelector: selec}
	f.Client.List(goctx.TODO(), &options, services)

	for i := 0; i < len(services.Items); i++ {
		if services.Items[i].Spec.Ports[0].NodePort != 0 && services.Items[i].Spec.Type == corev1.ServiceTypeNodePort && len(services.Items[i].Status.LoadBalancer.Ingress) == 0 {
			t.Log(services.Items[i])
			t.Log("Successfully set service as Node Port")
		} else {
			t.Log("Failed to set service as Node Port")
			t.Fail()
		}
	}
	return nil
}

func verifyLoadBalancer(t *testing.T, f *framework.Framework, app *appsodyv1beta1.AppsodyApplication, l labels.Set) error {
	services := &corev1.ServiceList{}
	selec := l.AsSelector()
	options := k.ListOptions{LabelSelector: selec}
	f.Client.List(goctx.TODO(), &options, services)

	for i := 0; i < len(services.Items); i++ {
		if len(services.Items[i].Status.LoadBalancer.Ingress) > 0 && services.Items[i].Spec.Type == corev1.ServiceTypeLoadBalancer {
			t.Log("Successfully set service as Load Balancer")
			t.Log(services.Items[i].Status.LoadBalancer.Ingress[0].IP)
		} else {
			t.Log("Failed to set service as Load Balancer")
			t.Fail()
		}
	}

	return nil
}
