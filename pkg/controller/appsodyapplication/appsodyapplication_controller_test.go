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
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func TestAppsodyControllerServiceAccountHasValue(t *testing.T) {
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
	s.AddKnownTypes(appsodyv1alpha1.SchemeGroupVersion, appsody)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	m := make(map[string]appsodyv1alpha1.AppsodyApplicationSpec)
	m[stack] = appsodyv1alpha1.AppsodyApplicationSpec{
		ServiceAccountName: &serviceAccountName,
		Service:            &service,
	}

	// Create a ReconcileAppsodyApplication object with the scheme and fake client.
	r := &ReconcileAppsodyApplication{
		appsodyutils.NewReconcilerBase(cl, s),
		m,
	}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
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
	t.Log(appsody)

	// Check the result of reconciliation to make sure it has the desired state.
	// if !res.Requeue {
	// 	t.Error("reconcile did not requeue request as expected")
	// }

	// Check if deployment has been created and has the correct size.
	dep := &appsv1.Deployment{}
	err = r.GetClient().Get(context.TODO(), req.NamespacedName, dep)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}

	// Test if service name is assigned in deployment
	sa := dep.Spec.Template.Spec.ServiceAccountName
	if sa != serviceAccountName {
		t.Errorf("Service account name (%v) was not expected service account name (%s)", sa, serviceAccountName)
	}

}

func TestAppsodyControllerServiceAccountIsNil(t *testing.T) {
	// set the logger to development mode for verbose logs
	//logf.SetLogger(logf.ZapLogger(true))

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
	s.AddKnownTypes(appsodyv1alpha1.SchemeGroupVersion, appsody)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	m := make(map[string]appsodyv1alpha1.AppsodyApplicationSpec)
	m[stack] = appsodyv1alpha1.AppsodyApplicationSpec{
		Service: &service,
	}

	// Create a ReconcileAppsodyApplication object with the scheme and fake client.
	r := &ReconcileAppsodyApplication{
		appsodyutils.NewReconcilerBase(cl, s),
		m,
	}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
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

	// Check if deployment has been created and has the correct size.
	dep := &appsv1.Deployment{}
	err = r.GetClient().Get(context.TODO(), req.NamespacedName, dep)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}
	// Test if service name is assigned in deployment
	sa := dep.Spec.Template.Spec.ServiceAccountName
	if sa != name {
		t.Errorf("Service account name (%v) was not expected service account name (%s)", sa, name)
	}
}

func TestAppsodyControllerStorage(t *testing.T) {
	// set the logger to development mode for verbose logs
	logf.SetLogger(logf.ZapLogger(true))

	appsody := &appsodyv1alpha1.AppsodyApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsodyv1alpha1.AppsodyApplicationSpec{
			ApplicationImage: appImage,
			Replicas:         &replicas,
			Storage:          &storage,
			Stack:            stack,
		},
	}

	// objects to track in the fake client
	objs := []runtime.Object{appsody}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appsodyv1alpha1.SchemeGroupVersion, appsody)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	m := make(map[string]appsodyv1alpha1.AppsodyApplicationSpec)

	m[stack] = appsodyv1alpha1.AppsodyApplicationSpec{
		ServiceAccountName: &serviceAccountName,
		Service:            &service,
	}

	// Create a ReconcileAppsodyApplication object with the scheme and fake client.
	r := &ReconcileAppsodyApplication{
		appsodyutils.NewReconcilerBase(cl, s),
		m,
	}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
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

	// Check if StatefulSet has been created and has the correct size.
	statefulSet := &appsv1.StatefulSet{}
	err = r.GetClient().Get(context.TODO(), req.NamespacedName, statefulSet)
	if err != nil {

		t.Fatalf("get StatefulSet: (%v)", err)
	}

	// Check if the quantity of Replicas for this deployment equals the specification
	size := *statefulSet.Spec.Replicas

	if size != replicas {
		t.Errorf("StatefulSet size (%v) is not the expected size (%d)", size, replicas)
	}

	image := statefulSet.Spec.Template.Spec.Containers[0].Image
	if image != appImage {
		t.Errorf("StatefulSet application image (%v) is not the expected application image (%s)", image, appImage)
	}

	sa := statefulSet.Spec.Template.Spec.ServiceAccountName
	if sa != serviceAccountName {
		t.Errorf("Service account name (%v) was not expected service account name (%s)", sa, serviceAccountName)
	}

	serviceName := statefulSet.Spec.ServiceName
	newName := name + "-headless"
	if serviceName != newName {
		t.Errorf("ServiceName (%v) was not the expected ServiceName (%s)", serviceName, newName)
	}
}

func TestAppsodyControllerCreatingKNative(t *testing.T) {
	// set the logger to development mode for verbose logs
	logf.SetLogger(logf.ZapLogger(true))

	appsody := &appsodyv1alpha1.AppsodyApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsodyv1alpha1.AppsodyApplicationSpec{
			ApplicationImage:   appImage,
			ServiceAccountName: &serviceAccountName,
			Service:            &service,
			Stack:              stack,
		},
	}

	// objects to track in the fake client
	objs := []runtime.Object{appsody}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme

	if err := servingv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add servingv1alpha1 scheme: (%v)", err)
	}

	s.AddKnownTypes(appsodyv1alpha1.SchemeGroupVersion, appsody)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	m := make(map[string]appsodyv1alpha1.AppsodyApplicationSpec)
	m[stack] = appsodyv1alpha1.AppsodyApplicationSpec{
		PullPolicy:           &pullPolicy,
		CreateKnativeService: &createKnativeService,
	}

	// Create a ReconcileAppsodyApplication object with the scheme and fake client.
	r := &ReconcileAppsodyApplication{
		appsodyutils.NewReconcilerBase(cl, s),
		m,
	}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
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

	ksvc := &servingv1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "serving.knative.dev/v1alpha1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	t.Log(appsody)

	err = r.GetClient().Get(context.TODO(), req.NamespacedName, ksvc)
	if err != nil {
		t.Fatalf("get ksvc: (%v)", err)
	}

	dName := "user-container"
	ksvcName := ksvc.Spec.Template.Spec.Containers[0].Name
	ksvcImage := ksvc.Spec.Template.Spec.Containers[0].Image
	ksvcPP := ksvc.Spec.Template.Spec.Containers[0].ImagePullPolicy
	ksvcServiceAccountName := ksvc.Spec.Template.Spec.ServiceAccountName

	if ksvcName != dName {
		t.Errorf("knative service name (%v) was not expected (%s)", ksvcServiceAccountName, dName)
	}
	if ksvcImage != appImage {
		t.Errorf("knative service image name (%v) was not expected (%s)", ksvcImage, appImage)
	}
	if ksvcPP != pullPolicy {
		t.Errorf("knative service pull policy (%v) was not expected (%s)", ksvcPP, pullPolicy)
	}
	if ksvcServiceAccountName != serviceAccountName {
		t.Errorf("knative service account name (%v) was not expected (%s)", ksvcServiceAccountName, serviceAccountName)
	}
}

func TestAppsodyControllerExpose(t *testing.T) {
	// set the logger to development mode for verbose logs
	logf.SetLogger(logf.ZapLogger(true))

	appsody := &appsodyv1alpha1.AppsodyApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsodyv1alpha1.AppsodyApplicationSpec{
			ServiceAccountName: &serviceAccountName,
			Service:            &service,
			Stack:              stack,
		},
	}

	// objects to track in the fake client
	objs := []runtime.Object{appsody}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme

	if err := routev1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add route scheme: (%v)", err)
	}

	s.AddKnownTypes(appsodyv1alpha1.SchemeGroupVersion, appsody)
	m := make(map[string]appsodyv1alpha1.AppsodyApplicationSpec)
	m[stack] = appsodyv1alpha1.AppsodyApplicationSpec{
		Expose: &expose,
	}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileAppsodyApplication object with the scheme and fake client.
	r := &ReconcileAppsodyApplication{
		appsodyutils.NewReconcilerBase(cl, s),
		m,
	}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
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

	route := &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "route.openshift.io/v1",
			Kind:       "Route",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err = r.GetClient().Get(context.TODO(), req.NamespacedName, route)
	if err != nil {
		t.Fatalf("get route: (%v)", err)
	}

	routePort := route.Spec.Port.TargetPort
	if routePort != intstr.FromInt(int(service.Port)) {
		t.Errorf("RoutePort (%v) is not the port (%d)", routePort, service.Port)
	}
}
