package appsodyapplication

import (
	"context"
	"testing"

	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	appsodyutils "github.com/appsody-operator/pkg/utils"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
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
	name                       = "app"
	namespace                  = "appsody"
	status                     = appsodyv1alpha1.AppsodyApplicationStatus{Conditions: []appsodyv1alpha1.StatusCondition{{Status: corev1.ConditionTrue}}}
	appImage                   = "my-image"
	ksvcAppImage               = "ksvc-image"
	replicas             int32 = 3
	autoscaling                = &appsodyv1alpha1.AppsodyApplicationAutoScaling{MaxReplicas: 3}
	pullPolicy                 = corev1.PullAlways
	serviceType                = corev1.ServiceTypeClusterIP
	service                    = &appsodyv1alpha1.AppsodyApplicationService{Type: &serviceType, Port: 8443}
	genService                 = &appsodyv1alpha1.AppsodyApplicationService{Type: &serviceType, Port: 9080}
	expose                     = true
	serviceAccountName         = "service-account"
	volumeCT                   = &corev1.PersistentVolumeClaim{TypeMeta: metav1.TypeMeta{Kind: "StatefulSet"}}
	storage                    = appsodyv1alpha1.AppsodyApplicationStorage{Size: "10Mi", MountPath: "/mnt/data", VolumeClaimTemplate: volumeCT}
	createKnativeService       = true
	stack                      = "java-microprofile"
	genStack                   = "generic"
	statefulSetSN              = name + "-headless"
	defaultKSVCName            = "user-container"
)

type Test struct {
	test     string
	expected interface{}
	actual   interface{}
}

func TestAppsodyController(t *testing.T) {
	// Set the logger to development mode for verbose logs
	logf.SetLogger(logf.ZapLogger(true))

	spec := appsodyv1alpha1.AppsodyApplicationSpec{Stack: stack}
	appsody := createAppsodyApp(name, namespace, spec)

	// Set objects to track in the fake client and register operator types with the runtime scheme.
	objs, s := []runtime.Object{appsody}, scheme.Scheme

	// Add third party resrouces to scheme
	if err := servingv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add servingv1alpha1 scheme: (%v)", err)
	}

	if err := routev1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add route scheme: (%v)", err)
	}

	s.AddKnownTypes(appsodyv1alpha1.SchemeGroupVersion, appsody)

	// Create a fake client to mock API calls.
	cl := fakeclient.NewFakeClient(objs...)

	rb := appsodyutils.NewReconcilerBase(cl, s, &rest.Config{}, record.NewFakeRecorder(10))
	defaultsMap := map[string]appsodyv1alpha1.AppsodyApplicationSpec{
		stack:    {ServiceAccountName: &serviceAccountName, Service: service},
		genStack: {Service: genService},
	}
	constantsMap := map[string]*appsodyv1alpha1.AppsodyApplicationSpec{}

	// Create a ReconcileAppsodyApplication object
	r := &ReconcileAppsodyApplication{rb, defaultsMap, constantsMap}
	r.SetDiscoveryClient(createFakeDiscoveryClient())

	// Mock request to simulate Reconcile being called on an event for a watched resource
	// then ensure reconcile is successful and does not return an empty result
	req := createReconcileRequest(name, namespace)
	res, err := r.Reconcile(req)
	verifyReconcile(res, err, t)

	// Check if deployment has been created
	dep := &appsv1.Deployment{}
	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, dep); err != nil {
		t.Fatalf("Get Deployment: (%v)", err)
	}

	// Check updated values in deployment
	depTests := []Test{
		{"service account name", serviceAccountName, dep.Spec.Template.Spec.ServiceAccountName},
	}
	verifyTests("dep", depTests, t)

	// Update appsody with values for StatefulSet
	// Update ServiceAccountName for empty case
	*r.StackDefaults[stack].ServiceAccountName = ""
	appsody.Spec = appsodyv1alpha1.AppsodyApplicationSpec{
		Stack:            stack,
		Storage:          &storage,
		Replicas:         &replicas,
		ApplicationImage: appImage,
	}
	updateAppsody(r, appsody, t)

	// Reconcile again to check for the StatefulSet and updated resources
	res, err = r.Reconcile(req)
	verifyReconcile(res, err, t)

	// Check if StatefulSet has been created
	statefulSet := &appsv1.StatefulSet{}
	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, statefulSet); err != nil {
		t.Fatalf("Get StatefulSet: (%v)", err)
	}

	// Storage is enabled so the deployment should be deleted
	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, dep); err == nil {
		t.Fatalf("Deployment was not deleted")
	}

	// Check updated values in StatefulSet
	ssTests := []Test{
		{"replicas", replicas, *statefulSet.Spec.Replicas},
		{"service image name", appImage, statefulSet.Spec.Template.Spec.Containers[0].Image},
		{"pull policy", name, statefulSet.Spec.Template.Spec.ServiceAccountName},
		{"service account name", statefulSetSN, statefulSet.Spec.ServiceName},
	}
	verifyTests("statefulSet", ssTests, t)

	// Enable CreateKnativeService
	appsody.Spec = appsodyv1alpha1.AppsodyApplicationSpec{
		Stack:                stack,
		CreateKnativeService: &createKnativeService,
		PullPolicy:           &pullPolicy,
		ApplicationImage:     ksvcAppImage,
	}
	updateAppsody(r, appsody, t)

	// Reconcile again to check for the KNativeService and updated resources
	res, err = r.Reconcile(req)
	verifyReconcile(res, err, t)

	// Create KnativeService
	ksvc := &servingv1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "serving.knative.dev/v1alpha1",
			Kind:       "Service",
		},
	}
	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, ksvc); err != nil {
		t.Fatalf("Get KnativeService: (%v)", err)
	}

	// KnativeService is enabled so non-Knative resources should be deleted
	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, statefulSet); err == nil {
		t.Fatalf("StatefulSet was not deleted")
	}

	// Check updated values in KnativeService
	ksvcTests := []Test{
		{"service name", defaultKSVCName, ksvc.Spec.Template.Spec.Containers[0].Name},
		{"service image name", ksvcAppImage, ksvc.Spec.Template.Spec.Containers[0].Image},
		{"pull policy", pullPolicy, ksvc.Spec.Template.Spec.Containers[0].ImagePullPolicy},
		{"service account name", name, ksvc.Spec.Template.Spec.ServiceAccountName},
	}
	verifyTests("ksvc", ksvcTests, t)

	// Disable Knative and enable Expose to test route
	appsody.Spec = appsodyv1alpha1.AppsodyApplicationSpec{Stack: stack, Expose: &expose}
	updateAppsody(r, appsody, t)

	// Reconcile again to check for the route and updated resources
	res, err = r.Reconcile(req)
	verifyReconcile(res, err, t)

	// Create Route
	route := &routev1.Route{
		TypeMeta: metav1.TypeMeta{APIVersion: "route.openshift.io/v1", Kind: "Route"},
	}
	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, route); err != nil {
		t.Fatalf("Get Route: (%v)", err)
	}

	// Check updated values in Route
	routeTests := []Test{{"target port", intstr.FromInt(int(service.Port)), route.Spec.Port.TargetPort}}
	verifyTests("route", routeTests, t)

	// Disable Route/Expose and enable Autoscaling
	appsody.Spec = appsodyv1alpha1.AppsodyApplicationSpec{
		Stack:       stack,
		Autoscaling: autoscaling,
	}
	updateAppsody(r, appsody, t)

	// Reconcile again to check for hpa and updated resources
	res, err = r.Reconcile(req)
	verifyReconcile(res, err, t)

	// Create HorizontalPodAutoscaler
	hpa := &autoscalingv1.HorizontalPodAutoscaler{}
	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, hpa); err != nil {
		t.Fatalf("Get HPA: (%v)", err)
	}

	// Expose is disabled so route should be deleted
	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, route); err == nil {
		t.Fatal("Route was not deleted")
	}

	// Check updated values in hpa
	hpaTests := []Test{{"max replicas", autoscaling.MaxReplicas, hpa.Spec.MaxReplicas}}
	verifyTests("hpa", hpaTests, t)

	// Remove autoscaling to ensure hpa is deleted
	// Remove stack: "java-microprofile" from appsody to test "generic" stack
	appsody.Spec.Autoscaling = nil
	appsody.Spec.Stack = ""
	updateAppsody(r, appsody, t)

	res, err = r.Reconcile(req)
	verifyReconcile(res, err, t)

	// Autoscaling is disabled so hpa should be deleted
	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, hpa); err == nil {
		t.Fatal("hpa was not deleted")
	}

	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, appsody); err != nil {
		t.Fatalf("Get appsody: (%v)", err)
	}

	// Check updated values in appsody
	genStackTests := []Test{{"service port", genService.Port, appsody.Spec.Service.Port}}
	verifyTests("generic stack", genStackTests, t)

	// Update appsody to ensure it requeues
	appsody.SetGeneration(1)
	updateAppsody(r, appsody, t)

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
}

func TestConfigMapDefaults(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	spec := appsodyv1alpha1.AppsodyApplicationSpec{Stack: stack, Service: service}
	appsody := createAppsodyApp(name, namespace, spec)

	objs, s := []runtime.Object{appsody}, scheme.Scheme
	s.AddKnownTypes(appsodyv1alpha1.SchemeGroupVersion, appsody)
	cl := fakeclient.NewFakeClient(objs...)

	rb := appsodyutils.NewReconcilerBase(cl, s, &rest.Config{}, record.NewFakeRecorder(10))
	defaultsMap := map[string]appsodyv1alpha1.AppsodyApplicationSpec{stack: {Service: service}}
	constantsMap := map[string]*appsodyv1alpha1.AppsodyApplicationSpec{}

	r := &ReconcileAppsodyApplication{rb, defaultsMap, constantsMap}
	r.SetDiscoveryClient(createFakeDiscoveryClient())

	// Create request for defaults case
	req := createReconcileRequest("appsody-operator", namespace)

	// Create configMap for defaults case
	data := map[string]string{stack: `{"expose":true}`}
	configMap := createConfigMap("appsody-operator", namespace, data)
	if err := r.GetClient().Create(context.TODO(), configMap); err != nil {
		t.Fatalf("Create configMap: (%v)", err)
	}

	res, err := r.Reconcile(req)
	verifyReconcile(res, err, t)

	// Update request name
	req = createReconcileRequest(name, namespace)
	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, appsody); err != nil {
		t.Fatalf("Get appsody: (%v)", err)
	}

	// Check updated values in appsody
	configMapDefTests := []Test{{"expose", true, *appsody.Spec.Expose}}
	verifyTests("configMapDefaults", configMapDefTests, t)
}

func TestConfigMapConstants(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	spec := appsodyv1alpha1.AppsodyApplicationSpec{Stack: stack}
	appsody := createAppsodyApp(name, namespace, spec)

	objs, s := []runtime.Object{appsody}, scheme.Scheme
	s.AddKnownTypes(appsodyv1alpha1.SchemeGroupVersion, appsody)
	cl := fakeclient.NewFakeClient(objs...)

	rb := appsodyutils.NewReconcilerBase(cl, s, &rest.Config{}, record.NewFakeRecorder(10))
	defaultsMap := map[string]appsodyv1alpha1.AppsodyApplicationSpec{stack: {Service: service}}
	constantsMap := map[string]*appsodyv1alpha1.AppsodyApplicationSpec{stack: {Service: service}}

	r := &ReconcileAppsodyApplication{rb, defaultsMap, constantsMap}
	r.SetDiscoveryClient(createFakeDiscoveryClient())

	// Create request for constants case
	req := createReconcileRequest("appsody-operator-constants", namespace)

	// Expose enabled and port updated to 3000
	data := map[string]string{stack: `{"expose":true, "service":{"port": 3000,"type": "ClusterIP"}}`}
	configMap := createConfigMap("appsody-operator-constants", namespace, data)

	if err := r.GetClient().Create(context.TODO(), configMap); err != nil {
		t.Fatalf("Create configMap: (%v)", err)
	}

	res, err := r.Reconcile(req)
	verifyReconcile(res, err, t)

	// Update reconcile request name
	req = createReconcileRequest(name, namespace)
	_, err = r.Reconcile(req)

	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	if err = r.GetClient().Get(context.TODO(), req.NamespacedName, appsody); err != nil {
		t.Fatalf("Get appsody: (%v)", err)
	}

	configMapConstTests := []Test{
		{"expose", true, *appsody.Spec.Expose},
		{"service port", int32(3000), appsody.Spec.Service.Port},
	}
	verifyTests("configMapConstants", configMapConstTests, t)
}

// Helper Functions
func createAppsodyApp(n, ns string, spec appsodyv1alpha1.AppsodyApplicationSpec) *appsodyv1alpha1.AppsodyApplication {
	app := &appsodyv1alpha1.AppsodyApplication{
		ObjectMeta: metav1.ObjectMeta{Name: n, Namespace: ns},
		Spec:       spec,
		Status:     status,
	}
	return app
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

func createReconcileRequest(n, ns string) reconcile.Request {
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{Name: n, Namespace: ns},
	}
	return req
}

func createConfigMap(n, ns string, data map[string]string) *corev1.ConfigMap {
	app := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: n, Namespace: ns},
		Data:       data,
	}
	return app
}

func verifyReconcile(res reconcile.Result, err error, t *testing.T) {
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	if res != (reconcile.Result{}) {
		t.Errorf("reconcile did not return an empty result (%v)", res)
	}
}

func verifyTests(n string, tests []Test, t *testing.T) {
	for _, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("%s %s test expected: (%v) actual: (%v)", n, tt.test, tt.expected, tt.actual)
		}
	}
}

func updateAppsody(r *ReconcileAppsodyApplication, appsody *appsodyv1alpha1.AppsodyApplication, t *testing.T) {
	if err := r.GetClient().Update(context.TODO(), appsody); err != nil {
		t.Fatalf("Update appsody: (%v)", err)
	}
}
