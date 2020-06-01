package e2e

import (
	goctx "context"
	"fmt"
	"errors"
	"testing"
	"time"
	"net/http"
	
	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	routev1 "github.com/openshift/api/route/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	certmngrv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	"github.com/appsody/appsody-operator/test/util"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/types"
)

const (
	tlsCrt = "faketlscrt"
	tlsKey = "faketlskey"
	caCrt = "fakecacrt"
	destCACrt = "fakedestcacrt"
)

// AppsodyCertManagerTest consists of five CertManager-related E2E tests.
func AppsodyCertManagerTest(t *testing.T) {
	// standard initialization
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

	// skip if cert manager not installed
	if !util.IsCertManagerInstalled(t, f, ctx) {
		t.Log("cert manager not installed, skipping...")
		return
	}

	// deplopy the operator first
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "appsody-operator", 1, retryInterval, timeout)
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// required to get route details later
	if err = routev1.AddToScheme(f.Scheme); err != nil {
		t.Logf("Unable to add route scheme: (%v)", err)
		util.FailureCleanup(t, f, namespace, err)
	}

	// start the five tests
	if err = appsodyPodCertTest(t, f, ctx); err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	if err = appsodyRouteCertTest(t, f, ctx); err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	if err = appsodyCustomIssuerTest(t, f, ctx); err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	if err = appsodyExistingCertTest(t, f, ctx); err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	if err = appsodyOpenShiftCATest(t, f, ctx); err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}
}

// Simple scenario test.
func appsodyPodCertTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	const name = "example-appsody-pod-cert"

	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	appsody := util.MakeBasicAppsodyApplication(t, f, name, namespace, 1)
	appsody.Spec.Service.Certificate = &appsodyv1beta1.Certificate{}

	certName := fmt.Sprintf("%s-svc-crt", name)
	err = deployAndWaitForCertificate("Creating cert-manager pod test",
			t, f, ctx, appsody, name, namespace, certName)
	return err	// implicitly return nil if no error occurs
}

// Test behaviour when specifying appsody.Spec.Route and then set it to nil.
func appsodyRouteCertTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	const name = "example-appsody-route-cert"

	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace %v", err)
	}

	appsody := util.MakeBasicAppsodyApplication(t, f, name, namespace, 1)
	terminationPolicy := routev1.TLSTerminationReencrypt
	expose := true
	appsody.Spec.Expose = &expose
	appsody.Spec.Route = &appsodyv1beta1.AppsodyRoute{
		Host:        "myapp.mycompany.com",
		Termination: &terminationPolicy,
		Certificate: &appsodyv1beta1.Certificate{},
	}

	certName := fmt.Sprintf("%s-route-crt", name)
	err = deployAndWaitForCertificate("Creating cert-manager route test",
			t, f, ctx, appsody, name, namespace, certName)
	if err != nil {
		return err
	}

	// set route to nil
	target := types.NamespacedName{Name: name, Namespace: namespace}
	err = util.UpdateApplication(f, target, func(r *appsodyv1beta1.AppsodyApplication) {
		r.Spec.Route = nil
	})
	if err != nil {
		return err
	}

	// wait for the change to take effect and verify the state
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, name, 1, retryInterval, operatorTimeout)
	if err != nil {
		return err
	}

	namespacedName := types.NamespacedName{Name: fmt.Sprintf("%s-route-crt", name), Namespace: namespace}
	certExists, certErr := certificateExists(f, namespacedName) 
	if certErr != nil {
		return certErr
	}
	if certExists {
		return errors.New("certificate persists when appsody.Spec.Route is nil")
	}

	return nil
}

// Test the scenario where we create our custom issuer and use it as our certificate issuer.
func appsodyCustomIssuerTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	// standard initialization
	const name = "example-custom-issuer-cert"
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace %v", err)
	}

	// Create a custom issuer, named 'custom-issuer'.
	err = util.CreateCertificateIssuer(t, f, ctx, "custom-issuer")
	if err != nil {
		issuerExists := err.Error() == "clusterissuers.cert-manager.io \"custom-issuer\" already exists"
		if !issuerExists {
			return err
		}
	}

	// configure the appsody application's spec
	appsody := util.MakeBasicAppsodyApplication(t, f, name, namespace, 1)
	terminationPolicy := routev1.TLSTerminationReencrypt
	expose := true
	var durationTime time.Duration = 10 * time.Minute
	duration := metav1.Duration{
		Duration: durationTime,
	}
	appsody.Spec.Expose = &expose
	appsody.Spec.Route = &appsodyv1beta1.AppsodyRoute{
		Host:        "myapp.mycompany.com",
		Termination: &terminationPolicy,
		Certificate: &appsodyv1beta1.Certificate{
			Duration:     &duration,
			Organization: []string{"My Company"},
			IssuerRef: cmmeta.ObjectReference{
				Name: "custom-issuer",
				Kind: "ClusterIssuer",
			},
		},
	}

	// deploy and wait for the certificate to be generated.
	certName := fmt.Sprintf("%s-route-crt", name)
	err = deployAndWaitForCertificate("Creating cert-manager custom issuer test",
			t, f, ctx, appsody, name, namespace, certName)
	return err	// implicitly return nil if no error occurs
}

// Test using an existing certificate for TLS connection.
func appsodyExistingCertTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	const name = "example-existing-cert"
	secretRefName := "myapp-rt-tls"
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace %v", err)
	}

	// Deploy the existing secret, a fake one generated by the helper function in our case.
	secret := makeCertSecret(secretRefName, namespace)
	err = f.Client.Create(goctx.TODO(), secret,
		&framework.CleanupOptions{TestContext: ctx, Timeout: time.Second, RetryInterval: time.Second})
	if err != nil {
		util.FailureCleanup(t, f, namespace, err)
	}

	// configure the appsody
	appsody := util.MakeBasicAppsodyApplication(t, f, name, namespace, 1)
	terminationPolicy := routev1.TLSTerminationReencrypt
	expose := true
	appsody.Spec.Expose = &expose
	appsody.Spec.Route = &appsodyv1beta1.AppsodyRoute{
		Host:        "myapp.mycompany.com",
		Termination: &terminationPolicy,
		CertificateSecretRef: &secretRefName,
	}

	// deploy the appsody
	timestamp := time.Now().UTC()
	t.Logf("%s - Creating cert-manager existing certificate test...", timestamp)

	err = f.Client.Create(goctx.TODO(), appsody,
		&framework.CleanupOptions{TestContext: ctx, Timeout: time.Second, RetryInterval: time.Second})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, name, 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	// test if the deployed appsody's fields are set correctly
	route := &routev1.Route{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, route)
	if err != nil {
		return err
	}
	if route.Spec.TLS.Certificate != tlsCrt ||
		route.Spec.TLS.CACertificate != caCrt ||
		route.Spec.TLS.Key != tlsKey ||
		route.Spec.TLS.DestinationCACertificate != destCACrt {
		return errors.New("route.Spec.TLS fields are not set correctly")
	}

	return nil
}

// Test using the OpenShift CA by adding annotations to the rutnime.
func appsodyOpenShiftCATest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	const name = "example-oc-cert"
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace %v", err)
	}
	namespacedName := types.NamespacedName{Name: name, Namespace: namespace}
	secretRefName := "openshift-generated-secret-"+namespace	// use a non-existent name

	// configure the appsody
	appsody := util.MakeBasicAppsodyApplication(t, f, name, namespace, 1)
	terminationPolicy := routev1.TLSTerminationReencrypt
	expose := true
	appsody.Spec.Expose = &expose
	annotations := map[string]string {
		"service.alpha.openshift.io/serving-cert-secret-name": secretRefName,
	}	// important step: add the annotation
	appsody.Spec.ApplicationImage = "navidsh/e2e-app-ssl"	// simple nodejs app with https enabled
	appsody.Spec.Service = &appsodyv1beta1.AppsodyApplicationService {
		Annotations: annotations,
		CertificateSecretRef: &secretRefName,
		Port: 3443,	// match the nodejs app's source code
	}

	insecureEdgeTerminationPolicy := routev1.InsecureEdgeTerminationPolicyRedirect
	appsody.Spec.Route = &appsodyv1beta1.AppsodyRoute{
		Termination: &terminationPolicy,
		InsecureEdgeTerminationPolicy: &insecureEdgeTerminationPolicy,
	}

	// deploy the appsody
	timestamp := time.Now().UTC()
	t.Logf("%s - Creating cert-manager OpenShift CA test...", timestamp)

	err = f.Client.Create(goctx.TODO(), appsody,
		&framework.CleanupOptions{TestContext: ctx, Timeout: time.Second, RetryInterval: time.Second})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, name, 1, retryInterval, timeout)
	if err != nil {
		return err
	}
	
	// check if the secret is automatically generated
	secret := &corev1.Secret{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: secretRefName, Namespace: namespace}, secret)
	if err != nil {
		return err
	}

	// try to initialize https connection
	err = makeHTTPSRequest(t, f, ctx, namespacedName)
	return err	// implicitly return nil if no error occurs
}

/* Helper Functions Below */
// deployAndWaitForCertificate is a helper function that deploy a appsody, wait for its deployment,
// and wait for the certificate to be gererated. (reduce code duplication)
func deployAndWaitForCertificate (msg string, t *testing.T, f *framework.Framework, ctx *framework.TestCtx, 
		appsody *appsodyv1beta1.AppsodyApplication, n string, ns string, certName string) error {
	timestamp := time.Now().UTC()
	t.Logf("%s - %s...", timestamp, msg)
	err := f.Client.Create(goctx.TODO(), appsody,
		&framework.CleanupOptions{TestContext: ctx, Timeout: time.Second, RetryInterval: time.Second})
	if err != nil {
		return err
	}

	err = e2eutil.WaitForDeployment(t, f.KubeClient, ns, n, 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = util.WaitForCertificate(t, f, ns, certName, retryInterval, timeout)
	return err	// implicitly return nil if no error occurs
}

// makeCertSecret returns a pointer to a simple Secret object with fake values inside.
func makeCertSecret(n string, ns string) *corev1.Secret {
	data := map[string][]byte{
		"ca.crt": []byte(caCrt),
		"tls.crt": []byte(tlsCrt),
		"tls.key": []byte(tlsKey),
		"destCA.crt": []byte(destCACrt),
	}
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: n,
			Namespace: ns,
		},
		Type: "kubernetes.io/tls",
		Data: data,
	}
	return &secret
}

// certificateExists checks if the certificate, named `n`, exists in the namespace `ns`.
func certificateExists(f *framework.Framework, namespacedName types.NamespacedName) (bool, error) {
	cert := &certmngrv1alpha2.Certificate{}
	certErr := f.Client.Get(goctx.TODO(), namespacedName, cert)
	if certErr != nil {
		if apierrors.IsNotFound(certErr) {
			return false, nil
		}
		return false, certErr
	}
	return true, nil
}

// makeHttpsRequest tries to poll a GET call to the deployment's route via https protocol.
// The expected result is a response with 200 status code.
// Return error if the status code is outside of the 200 range.
func makeHTTPSRequest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx,
		namespacedName types.NamespacedName) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		route := &routev1.Route{}
		err = f.Client.Get(goctx.TODO(), namespacedName, route)
		if err != nil {
			return true, err
		}

		resp, err := http.Get("https://" + route.Spec.Host)
		if err != nil {
			return true, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			t.Log("Retrying to make https connection ...")
			return false, nil
		}
		return true, nil
	})

	if errors.Is(err, wait.ErrWaitTimeout) {
		return errors.New("status code outside of 200 range upon initiating https request")
	}

	return err	// implicitly return nil if no errors
}