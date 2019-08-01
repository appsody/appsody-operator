package appsodyapplication

import (
	"context"
	"testing"

	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	appsodyutils "github.com/appsody-operator/pkg/utils"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	coretesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	serviceType                = corev1.ServiceTypeClusterIP
	pullPolicy                 = corev1.PullAlways
	name                       = "app"
	namespace                  = "appsody"
	appImage                   = "my-image"
	replicas             int32 = 3
	serviceAccountName         = "service-account"
	createKnativeService       = true
	pullSecret                 = "pass"
	service                    = appsodyv1alpha1.AppsodyApplicationService{
		Type: &serviceType,
		Port: 8443,
	}
	storage appsodyv1alpha1.AppsodyApplicationStorage = appsodyv1alpha1.AppsodyApplicationStorage{
		Size:      "10Mi",
		MountPath: "/mnt/data",
		VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				Kind: "StatefulSet",
			},
		},
	}
	expose = true
	stack  = "microprofile"
)

func TestAppsodyController(t *testing.T) {
	// set the logger to development mode for verbose logs
	logf.SetLogger(logf.ZapLogger(true))

	appsody := &appsodyv1alpha1.AppsodyApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsodyv1alpha1.AppsodyApplicationSpec{
			Stack: stack,
		},
	}

	// objects to track in the fake client
	objs := []runtime.Object{appsody}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme

	if err := servingv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add servingv1alpha1 scheme: (%v)", err)
	}

	if err := routev1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add route scheme: (%v)", err)
	}

	s.AddKnownTypes(appsodyv1alpha1.SchemeGroupVersion, appsody)

	// Create a fake client to mock API calls.
	cl := fakeclient.NewFakeClient(objs...)

	m := make(map[string]appsodyv1alpha1.AppsodyApplicationSpec)
	m[stack] = appsodyv1alpha1.AppsodyApplicationSpec{
		ServiceAccountName: &serviceAccountName,
		Service:            &service,
	}

	// Create a ReconcileAppsodyApplication object with the scheme and fake client.
	r := &ReconcileAppsodyApplication{
		appsodyutils.NewReconcilerBase(cl, s, &rest.Config{}, record.NewFakeRecorder(10)),
		m,
	}

	r.SetDiscoveryClient(createFakeDiscoveryClient())

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	// if !res.Requeue {
	// 	t.Error("reconcile did not requeue request as expected")
	// }

	// Check if deployment has been created
	dep := &appsv1.Deployment{}
	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, dep); err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}

	// Check if service name is assigned in deployment
	sa := dep.Spec.Template.Spec.ServiceAccountName
	if sa != serviceAccountName {
		t.Errorf("Service account name (%v) was not expected service account name (%s)", sa, serviceAccountName)
	}

	// Update appsody with values for statefulset
	appsody.Spec.Storage = &storage
	appsody.Spec.Replicas = &replicas
	appsody.Spec.ApplicationImage = appImage
	appsody.Spec.PullSecret = &pullSecret
	if err = r.GetClient().Update(context.TODO(), appsody); err != nil {
		t.Fatalf("Update appsody: (%v)", err)
	}

	// Update ServiceAccountName for empty case
	*r.StackDefaults[stack].ServiceAccountName = ""

	statefulSet := &appsv1.StatefulSet{}
	if err = r.GetClient().Create(context.TODO(), statefulSet); err != nil {
		t.Fatalf("create StatefulSet: (%v)", err)
	}

	// reconcile again to check for the StatefulSet and updates resources
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	if res != (reconcile.Result{}) {
		t.Errorf("reconcile did not return an empty result")
	}

	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, statefulSet); err != nil {
		t.Fatalf("get StatefulSet: (%v)", err)
	}

	// make sure deployment gets deleted since storage is now enabled
	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, dep); err == nil {
		t.Fatalf("deployment was not deleted")
	}

	// Updated values in StatefulSet
	size := *statefulSet.Spec.Replicas
	image := statefulSet.Spec.Template.Spec.Containers[0].Image
	serviceName := statefulSet.Spec.ServiceName
	sa = statefulSet.Spec.Template.Spec.ServiceAccountName // should be equal to name
	// use updated name to agree with how reconcile will update
	newName := name + "-headless"

	if size != replicas {
		t.Errorf("Service account name (%v) was not expected service account name (%d)", size, replicas)
	}

	if image != appImage {
		t.Errorf("StatefulSet application image (%v) is not the expected application image (%s)", image, appImage)
	}

	// check the value is correct when ServiceAccountName is the empty string
	if sa != name {
		t.Errorf("Service account name (%v) was not expected service account name (%s)", sa, name)
	}

	if serviceName != newName {
		t.Errorf("ServiceName (%v) was not the expected ServiceName (%s)", serviceName, newName)
	}

	// enable CreateKnativeService
	appsody.Spec.CreateKnativeService = &createKnativeService
	appsody.Spec.PullPolicy = &pullPolicy
	if err = r.GetClient().Update(context.TODO(), appsody); err != nil {
		t.Fatalf("Update appsody: (%v)", err)
	}

	ksvc := &servingv1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "serving.knative.dev/v1alpha1",
			Kind:       "Service",
		},
	}
	if err = r.GetClient().Create(context.TODO(), ksvc); err != nil {
		t.Fatalf("create ksvc: (%v)", err)
	}

	// reconcile again to check for the kNativeService and updates resources
	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	if res != (reconcile.Result{}) {
		t.Errorf("reconcile did not return an empty result")
	}

	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, ksvc); err != nil {
		t.Fatalf("get StatefulSet: (%v)", err)
	}

	// make sure StatefulSet gets deleted since kNativeService is now enabled
	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, statefulSet); err == nil {
		t.Fatalf("StatefulSet was not deleted")
	}

	dName := "user-container"
	ksvcName := ksvc.Spec.Template.Spec.Containers[0].Name
	ksvcImage := ksvc.Spec.Template.Spec.Containers[0].Image
	ksvcPP := ksvc.Spec.Template.Spec.Containers[0].ImagePullPolicy
	ksvcServiceAccountName := ksvc.Spec.Template.Spec.ServiceAccountName

	if ksvcName != dName {
		t.Errorf("knative service name (%v) was not expected (%s)", ksvcName, dName)
	}
	if ksvcImage != appImage {
		t.Errorf("knative service image name (%v) was not expected (%s)", ksvcImage, appImage)
	}
	if ksvcPP != pullPolicy {
		t.Errorf("knative service pull policy (%v) was not expected (%s)", ksvcPP, pullPolicy)
	}
	if ksvcServiceAccountName != name {
		t.Errorf("knative service account name (%v) was not expected (%s)", ksvcServiceAccountName, name)
	}

	// disable knative to test route
	disable := false
	appsody.Spec.CreateKnativeService = &disable
	// enable expose
	appsody.Spec.Expose = &expose
	if err = r.GetClient().Update(context.TODO(), appsody); err != nil {
		t.Fatalf("Update appsody: (%v)", err)
	}

	route := &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "route.openshift.io/v1",
			Kind:       "Route",
		},
	}
	if err = r.GetClient().Create(context.TODO(), route); err != nil {
		t.Fatalf("create route: (%v)", err)
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	if res != (reconcile.Result{}) {
		t.Errorf("reconcile did not return an empty result")
	}

	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, route); err != nil {
		t.Fatalf("get route: (%v)", err)
	}

	routePort := route.Spec.Port.TargetPort
	if routePort != intstr.FromInt(int(service.Port)) {
		t.Errorf("RoutePort (%v) is not the port (%d)", routePort, service.Port)
	}

	// disable expose to ensure route becomes deleted
	appsody.Spec.Expose = &disable
	if err = r.GetClient().Update(context.TODO(), appsody); err != nil {
		t.Fatalf("Update appsody: (%v)", err)
	}

	// reconcile again to check for the kNativeService and updates resources
	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	if res != (reconcile.Result{}) {
		t.Errorf("reconcile did not return an empty result")
	}

	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, route); err == nil {
		t.Fatalf("route was not deleted")
	}
}

func createFakeDiscoveryClient() discovery.DiscoveryInterface {
	fakeDiscoveryClient := &fakediscovery.FakeDiscovery{Fake: &coretesting.Fake{}}
	fakeDiscoveryClient.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: routev1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "routes", Namespaced: true, Kind: "Route"},
			},
		},
		{
			GroupVersion: servingv1alpha1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "services", Namespaced: true, Kind: "Service", SingularName: "service"},
			},
		},
	}

	return fakeDiscoveryClient
}
