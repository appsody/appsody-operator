package e2e

import (
	goctx "context"
	"testing"
	"time"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	"github.com/appsody/appsody-operator/test/util"
	prometheusv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	e2eutil "github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	k "sigs.k8s.io/controller-runtime/pkg/client"
)

// AppsodyServiceMonitorTest ...
func AppsodyServiceMonitorTest(t *testing.T) {
	ctx, err := util.InitializeContext(t, cleanupTimeout, retryInterval)
	defer ctx.Cleanup()
	if err != nil {
		t.Fatal(err)
	}

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("Couldn't get namespace: %v", err)
	}

	t.Logf("Namespace: %s", namespace)
	f := framework.Global

	// Adds the prometheus resources to the scheme
	if err = prometheusv1.AddToScheme(f.Scheme); err != nil {
		t.Fatalf("Unable to add prometheus scheme: (%v)", err)
	}

	// create one replica of the operator deployment in current namespace with provided name
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, operatorTimeout)
	if err != nil {
		t.Fatal(err)
	}

	helper := int32(1)
	appsody := util.MakeBasicAppsodyApplication(t, f, "example-appsody-sm", namespace, helper)

	// Create application deployment and wait
	err = f.Client.Create(goctx.TODO(), appsody, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second, RetryInterval: time.Second})
	if err != nil {
		t.Fatal(err)
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-sm", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	// Returns a list of the service monitor with the specified label
	m := map[string]string{"apps-prometheus": ""}
	l := labels.Set(m)
	selec := l.AsSelector()

	smList := &prometheusv1.ServiceMonitorList{}
	options := k.ListOptions{LabelSelector: selec}

	// If there are no service monitors deployed an error will be thrown below
	err = f.Client.List(goctx.TODO(), &options, smList)
	if err != nil {
		t.Fatal(err)
	}

	if len(smList.Items) != 0 {
		t.Fatal("There is another service monitor running")
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-sm", Namespace: namespace}, appsody)
	if err != nil {
		t.Log(err)
	}

	// Adds the mandatory label to the application so it will be picked up by the prometheus operator
	label := map[string]string{"apps-prometheus": ""}
	monitor := &appsodyv1beta1.AppsodyApplicationMonitoring{Labels: label}
	appsody.Spec.Monitoring = monitor

	// Updates the application so the operator is reconciled
	helper = int32(2)
	appsody.Spec.Replicas = &helper

	err = f.Client.Update(goctx.TODO(), appsody)
	if err != nil {
		t.Fatal(err)
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-sm", 2, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	// If there are no service monitors deployed an error will be thrown below
	err = f.Client.List(goctx.TODO(), &options, smList)
	if err != nil {
		t.Fatal(err)
	}

	// Gets the service monitor
	sm := smList.Items[0]

	smPath := sm.Spec.Endpoints[0].Path
	smPort := sm.Spec.Endpoints[0].Port
	smParams := sm.Spec.Endpoints[0].Params
	smScheme := sm.Spec.Endpoints[0].Scheme
	smScrapeTimeout := sm.Spec.Endpoints[0].ScrapeTimeout
	smInterval := sm.Spec.Endpoints[0].Interval
	smBTK := sm.Spec.Endpoints[0].BearerTokenFile

	if sm.Spec.Selector.MatchLabels["app.kubernetes.io/name"] != "example-appsody-sm" {
		t.Fatal("The service monitor is not connected to the appsody application?")
	}

	if smPath != "" {
		t.Fatal("The service monitor path default is incorrect")
	}

	if smPort != "3000-tcp" {
		t.Fatal("The service monitor port default is incorrect")
	}

	if smParams != nil {
		t.Fatal("The service monitor params default is incorrect")
	}

	if smScheme != "" {
		t.Fatal("The service monitor scheme default is incorrect")
	}

	if smScrapeTimeout != "" {
		t.Fatal("The service monitor scrape timeout default is incorrect")
	}

	if smInterval != "" {
		t.Fatal("The service monitor interval default is incorrect")
	}

	if smBTK != "" {
		t.Fatal("The service monitor bearer token file default is incorrect")
	}

	testSettingAppsodyServiceMonitor(t, f, namespace, appsody)
}

func testSettingAppsodyServiceMonitor(t *testing.T, f *framework.Framework, namespace string, appsody *appsodyv1beta1.AppsodyApplication) {
	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-appsody-sm", Namespace: namespace}, appsody)
	if err != nil {
		t.Log(err)
	}

	params := map[string][]string{
		"params": []string{"param1", "param2"},
	}
	username := v1.SecretKeySelector{Key: "username"}
	password := v1.SecretKeySelector{Key: "password"}

	// Creates the endpoint fields the user can customize
	endpoint := prometheusv1.Endpoint{
		Path:            "/path",
		Scheme:          "myScheme",
		Params:          params,
		Interval:        "30s",
		ScrapeTimeout:   "10s",
		TLSConfig:       &prometheusv1.TLSConfig{InsecureSkipVerify: true},
		BearerTokenFile: "myBTK",
		BasicAuth:       &prometheusv1.BasicAuth{Username: username, Password: password},
	}

	endpoints := []prometheusv1.Endpoint{endpoint}

	// Adds the mandatory label to the application so it will be picked up by the prometheus operator
	label := map[string]string{"apps-prometheus": ""}
	monitor := &appsodyv1beta1.AppsodyApplicationMonitoring{Labels: label, Endpoints: endpoints}
	appsody.Spec.Monitoring = monitor

	// Updates the application so the operator is reconciled
	helper := int32(3)
	appsody.Spec.Replicas = &helper

	err = f.Client.Update(goctx.TODO(), appsody)
	if err != nil {
		t.Fatal(err)
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-appsody-sm", 3, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	// Returns a list of the service monitor with the specified label
	m := map[string]string{"apps-prometheus": ""}
	l := labels.Set(m)
	selec := l.AsSelector()

	smList := &prometheusv1.ServiceMonitorList{}
	options := k.ListOptions{LabelSelector: selec}

	// If there are no service monitors deployed an error will be thrown below
	err = f.Client.List(goctx.TODO(), &options, smList)
	if err != nil {
		t.Fatal(err)
	}

	// Gets the service monitor
	sm := smList.Items[0]

	smPath := sm.Spec.Endpoints[0].Path
	smPort := sm.Spec.Endpoints[0].Port
	smParams := sm.Spec.Endpoints[0].Params
	smScheme := sm.Spec.Endpoints[0].Scheme
	smScrapeTimeout := sm.Spec.Endpoints[0].ScrapeTimeout
	smInterval := sm.Spec.Endpoints[0].Interval
	smBTK := sm.Spec.Endpoints[0].BearerTokenFile
	smTLSConfig := sm.Spec.Endpoints[0].TLSConfig
	smBasicAuth := sm.Spec.Endpoints[0].BasicAuth

	if sm.Spec.Selector.MatchLabels["app.kubernetes.io/name"] != "example-appsody-sm" {
		t.Fatal("The service monitor is not connected to the appsody application?")
	}

	if smPath != "/path" {
		t.Fatal("The service monitor path is incorrect")
	}

	if smPort != "3000-tcp" {
		t.Fatal("The service monitor port is incorrect")
	}

	if smParams == nil {
		t.Fatal("The service monitor params is incorrect")
	}

	if smScheme != "myScheme" {
		t.Fatal("The service monitor scheme is incorrect")
	}

	if smScrapeTimeout != "10s" {
		t.Fatal("The service monitor scrape timeout is incorrect")
	}

	if smInterval != "30s" {
		t.Fatal("The service monitor interval is incorrect")
	}

	if smBTK != "myBTK" {
		t.Fatal("The service monitor bearer token file is incorrect")
	}

	if smTLSConfig == nil {
		t.Fatal("The service monitor TLSConfig is not set")
	}

	if smBasicAuth == nil {
		t.Fatal("The service monitor basic auth is not set")
	}

}
