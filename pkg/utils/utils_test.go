package utils

import (
	"reflect"
	"testing"

	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	name               = "my-app"
	namespace          = "appsody"
	stack              = "java-microprofile"
	appImage           = "my-image"
	replicas     int32 = 2
	expose             = true
	createKNS          = true
	targetCPUPer int32 = 30
	autoscaling        = &appsodyv1alpha1.AppsodyApplicationAutoScaling{
		TargetCPUUtilizationPercentage: &targetCPUPer,
		MinReplicas:                    &replicas,
		MaxReplicas:                    3,
	}
	envFrom            = []corev1.EnvFromSource{{Prefix: namespace}}
	env                = []corev1.EnvVar{{Name: namespace}}
	pullPolicy         = corev1.PullAlways
	pullSecret         = "mysecret"
	serviceAccountName = "service-account"
	serviceType        = corev1.ServiceTypeClusterIP
	service            = &appsodyv1alpha1.AppsodyApplicationService{Type: &serviceType, Port: 8443}
	volumeCT           = &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "pvc", Namespace: namespace},
		TypeMeta:   metav1.TypeMeta{Kind: "StatefulSet"}}
	storage        = appsodyv1alpha1.AppsodyApplicationStorage{Size: "10Mi", MountPath: "/mnt/data", VolumeClaimTemplate: volumeCT}
	arch           = []string{"ppc64le"}
	readinessProbe = &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet:   &corev1.HTTPGetAction{},
			TCPSocket: &corev1.TCPSocketAction{},
		},
	}
	livenessProbe = &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet:   &corev1.HTTPGetAction{},
			TCPSocket: &corev1.TCPSocketAction{},
		},
	}
	volume      = corev1.Volume{Name: "appsody-volume"}
	volumeMount = corev1.VolumeMount{Name: volumeCT.Name, MountPath: storage.MountPath}
	resLimits   = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU: resource.Quantity{},
	}
	resourceContraints = &corev1.ResourceRequirements{Limits: resLimits}
)

type Test struct {
	test     string
	expected interface{}
	actual   interface{}
}

func TestCustomizeRoute(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	spec := appsodyv1alpha1.AppsodyApplicationSpec{Service: service}
	route, appsody := &routev1.Route{}, createAppsodyApp(name, namespace, spec)

	CustomizeRoute(route, appsody)

	//TestGetLabels
	testCR := []Test{
		{"Route labels", name, route.Labels["app.kubernetes.io/name"]},
		{"Route target kind", "Service", route.Spec.To.Kind},
		{"Route target name", name, route.Spec.To.Name},
		{"Route target weight", int32(100), *route.Spec.To.Weight},
		{"Route service target port", intstr.FromInt(int(appsody.Spec.Service.Port)), route.Spec.Port.TargetPort},
	}

	verifyTests(testCR, t)
}

func TestCustomizeService(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	spec := appsodyv1alpha1.AppsodyApplicationSpec{Service: service}
	svc, appsody := &corev1.Service{}, createAppsodyApp(name, namespace, spec)

	CustomizeService(svc, appsody)
	testCS := []Test{
		{"Service number of exposed ports", 1, len(svc.Spec.Ports)},
		{"Sercice first exposed port", appsody.Spec.Service.Port, svc.Spec.Ports[0].Port},
		{"Service first exposed target port", intstr.FromInt(int(appsody.Spec.Service.Port)), svc.Spec.Ports[0].TargetPort},
		{"Service type", *appsody.Spec.Service.Type, svc.Spec.Type},
		{"Service selector", name, svc.Spec.Selector["app.kubernetes.io/name"]},
	}
	verifyTests(testCS, t)
}

func TestCustomizePodSpec(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	spec := appsodyv1alpha1.AppsodyApplicationSpec{
		ApplicationImage:    appImage,
		Service:             service,
		ResourceConstraints: resourceContraints,
		ReadinessProbe:      readinessProbe,
		LivenessProbe:       livenessProbe,
		VolumeMounts:        []corev1.VolumeMount{volumeMount},
		PullPolicy:          &pullPolicy,
		Env:                 env,
		EnvFrom:             envFrom,
		Volumes:             []corev1.Volume{volume},
	}
	pts, appsody := &corev1.PodTemplateSpec{}, createAppsodyApp(name, namespace, spec)
	// else cond
	CustomizePodSpec(pts, appsody)
	noCont := len(pts.Spec.Containers)
	noPorts := len(pts.Spec.Containers[0].Ports)
	ptsSAN := pts.Spec.ServiceAccountName
	// if cond
	spec = appsodyv1alpha1.AppsodyApplicationSpec{
		ApplicationImage:    appImage,
		Service:             service,
		ResourceConstraints: resourceContraints,
		ReadinessProbe:      readinessProbe,
		LivenessProbe:       livenessProbe,
		VolumeMounts:        []corev1.VolumeMount{volumeMount},
		PullPolicy:          &pullPolicy,
		Env:                 env,
		EnvFrom:             envFrom,
		Volumes:             []corev1.Volume{volume},
		Architecture:        arch,
		ServiceAccountName:  &serviceAccountName,
	}
	appsody = createAppsodyApp(name, namespace, spec)
	CustomizePodSpec(pts, appsody)
	ptsCSAN := pts.Spec.ServiceAccountName

	// affinity tests
	affArchs := pts.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0]
	weight := pts.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Weight
	prefAffArchs := pts.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Preference.MatchExpressions[0].Values[0]

	testCPS := []Test{
		{"No containers", 1, noCont},
		{"No port", 1, noPorts},
		{"No ServiceAccountName", name, ptsSAN},
		{"ServiceAccountName available", serviceAccountName, ptsCSAN},
	}
	verifyTests(testCPS, t)

	testCA := []Test{
		{"Archs", arch[0], affArchs},
		{"Weight", int32(1), int32(weight)},
		{"Archs", arch[0], prefAffArchs},
	}
	verifyTests(testCA, t)
}

func TestCustomizePersistence(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	spec := appsodyv1alpha1.AppsodyApplicationSpec{Storage: &storage}
	statefulSet, appsody := &appsv1.StatefulSet{}, createAppsodyApp(name, namespace, spec)
	statefulSet.Spec.Template.Spec.Containers = []corev1.Container{{}}
	statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{}
	// if vct == 0, appsodyVCT != nil, not found
	CustomizePersistence(statefulSet, appsody)
	ssK := statefulSet.Spec.VolumeClaimTemplates[0].TypeMeta.Kind
	ssMountPath := statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath

	//reset
	storageNilVCT := appsodyv1alpha1.AppsodyApplicationStorage{Size: "10Mi", MountPath: "/mnt/data", VolumeClaimTemplate: nil}
	spec = appsodyv1alpha1.AppsodyApplicationSpec{Storage: &storageNilVCT}
	statefulSet, appsody = &appsv1.StatefulSet{}, createAppsodyApp(name, namespace, spec)

	statefulSet.Spec.Template.Spec.Containers = []corev1.Container{{}}
	statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMount)
	//appsodyVCT == nil, found
	CustomizePersistence(statefulSet, appsody)
	ssVolumeMountName := statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name
	size := statefulSet.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage]
	testCP := []Test{
		{"Persistence kind with VCT", volumeCT.TypeMeta.Kind, ssK},
		{"PVC size", storage.Size, size.String()},
		{"Mount path", storage.MountPath, ssMountPath},
		{"Volume Mount Name", volumeCT.Name, ssVolumeMountName},
	}
	verifyTests(testCP, t)
}

func TestCustomizeServiceAccount(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	spec := appsodyv1alpha1.AppsodyApplicationSpec{PullSecret: &pullSecret}
	sa, appsody := &corev1.ServiceAccount{}, createAppsodyApp(name, namespace, spec)
	CustomizeServiceAccount(sa, appsody)
	emptySAIPS := sa.ImagePullSecrets[0].Name

	newSecret := "my-new-secret"
	spec = appsodyv1alpha1.AppsodyApplicationSpec{PullSecret: &newSecret}
	appsody = createAppsodyApp(name, namespace, spec)
	CustomizeServiceAccount(sa, appsody)

	testCSA := []Test{
		{"ServiceAccount image pull secrets is empty", pullSecret, emptySAIPS},
		{"ServiceAccount image pull secrets", newSecret, sa.ImagePullSecrets[0].Name},
	}
	verifyTests(testCSA, t)
}

func TestCustomizeKnativeService(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	spec := appsodyv1alpha1.AppsodyApplicationSpec{
		ApplicationImage: appImage,
		Service:          service,
		LivenessProbe:    livenessProbe,
		ReadinessProbe:   readinessProbe,
		PullPolicy:       &pullPolicy,
		Env:              env,
		EnvFrom:          envFrom,
		Volumes:          []corev1.Volume{volume},
	}
	ksvc, appsody := &servingv1alpha1.Service{}, createAppsodyApp(name, namespace, spec)

	CustomizeKnativeService(ksvc, appsody)
	ksvcContainerName := ksvc.Spec.Template.Spec.Containers[0].Name
	ksvcNumPorts := len(ksvc.Spec.Template.Spec.Containers[0].Ports)
	ksvcSAN := ksvc.Spec.Template.Spec.ServiceAccountName

	ksvcLPPort := ksvc.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Port
	ksvcLPTCP := ksvc.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket.Port
	ksvcRPPort := ksvc.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port
	ksvcRPTCP := ksvc.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket.Port

	spec = appsodyv1alpha1.AppsodyApplicationSpec{
		ApplicationImage:   appImage,
		Service:            service,
		PullPolicy:         &pullPolicy,
		Env:                env,
		EnvFrom:            envFrom,
		Volumes:            []corev1.Volume{volume},
		ServiceAccountName: &serviceAccountName,
		LivenessProbe:      livenessProbe,
		ReadinessProbe:     readinessProbe,
	}
	appsody = createAppsodyApp(name, namespace, spec)
	CustomizeKnativeService(ksvc, appsody)
	testCKS := []Test{
		{"ksvc Container Name", "user-container", ksvcContainerName},
		{"ksvc container ports", 1, ksvcNumPorts},
		{"ksvc ServiceAccountName is nil", name, ksvcSAN},
		{"ksvc ServiceAccountName not nil", *appsody.Spec.ServiceAccountName, ksvc.Spec.Template.Spec.ServiceAccountName},
		{"liveness probe port", intstr.IntOrString{}, ksvcLPPort},
		{"liveness probe TCP socket port", intstr.IntOrString{}, ksvcLPTCP},
		{"Readiness probe port", intstr.IntOrString{}, ksvcRPPort},
		{"Readiness probe TCP socket port", intstr.IntOrString{}, ksvcRPTCP},
	}
	verifyTests(testCKS, t)
}

func TestCustomizeHPA(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	spec := appsodyv1alpha1.AppsodyApplicationSpec{Autoscaling: autoscaling}
	hpa, appsody := &autoscalingv1.HorizontalPodAutoscaler{}, createAppsodyApp(name, namespace, spec)
	CustomizeHPA(hpa, appsody)
	nilSTRKind := hpa.Spec.ScaleTargetRef.Kind

	spec = appsodyv1alpha1.AppsodyApplicationSpec{Autoscaling: autoscaling, Storage: &storage}
	appsody = createAppsodyApp(name, namespace, spec)
	CustomizeHPA(hpa, appsody)
	STRKind := hpa.Spec.ScaleTargetRef.Kind

	testCHPA := []Test{
		{"Max replicas", autoscaling.MaxReplicas, hpa.Spec.MaxReplicas},
		{"Min replicas", *autoscaling.MinReplicas, *hpa.Spec.MinReplicas},
		{"Target CPU utilization", *autoscaling.TargetCPUUtilizationPercentage, *hpa.Spec.TargetCPUUtilizationPercentage},
		{"", name, hpa.Spec.ScaleTargetRef.Name},
		{"", "apps/v1", hpa.Spec.ScaleTargetRef.APIVersion},
		{"Storage not enabled", "Deployment", nilSTRKind},
		{"Storage enabled", "StatefulSet", STRKind},
	}
	verifyTests(testCHPA, t)
}

func TestInitAndValidate(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	emptyService := &appsodyv1alpha1.AppsodyApplicationService{Port: 0}
	appsody := createAppsodyApp(name, namespace, appsodyv1alpha1.AppsodyApplicationSpec{})
	defaults := appsodyv1alpha1.AppsodyApplicationSpec{
		PullSecret: &pullSecret,
		Service:    emptyService,
	}
	constants := &appsodyv1alpha1.AppsodyApplicationSpec{}

	InitAndValidate(appsody, defaults, constants)
	defNilPP := *appsody.Spec.PullPolicy
	defResConNil := *appsody.Spec.ResourceConstraints
	servType := *appsody.Spec.Service.Type
	servPort := appsody.Spec.Service.Port

	emptyService.Port = 0
	emptyService.Type = nil
	appsody = createAppsodyApp(name, namespace, appsodyv1alpha1.AppsodyApplicationSpec{Service: emptyService})
	defaults = appsodyv1alpha1.AppsodyApplicationSpec{
		PullPolicy:           &pullPolicy,
		PullSecret:           &pullSecret,
		ServiceAccountName:   &serviceAccountName,
		ReadinessProbe:       readinessProbe,
		LivenessProbe:        livenessProbe,
		Env:                  env,
		EnvFrom:              envFrom,
		Volumes:              []corev1.Volume{volume},
		VolumeMounts:         []corev1.VolumeMount{volumeMount},
		ResourceConstraints:  resourceContraints,
		Autoscaling:          autoscaling,
		Expose:               &expose,
		CreateKnativeService: &createKNS,
		Service:              service,
	}
	InitAndValidate(appsody, defaults, constants)

	testIAV := []Test{
		{"Appsody PullPolicy is nil", pullPolicy, *appsody.Spec.PullPolicy},
		{"Appsody and Defaults PullPolicy is nil", corev1.PullIfNotPresent, defNilPP},
		{"Appsody PullSecret is nil", pullSecret, *appsody.Spec.PullSecret},
		{"Appsody ServiceAccountName is nil", serviceAccountName, *appsody.Spec.ServiceAccountName},
		{"Appsody ReadinessProbe is nil", readinessProbe, appsody.Spec.ReadinessProbe},
		{"Appsody LivenessProbe is nil", livenessProbe, appsody.Spec.LivenessProbe},
		{"Appsody Env is nil", namespace, appsody.Spec.Env[0].Name},
		{"Appsody EnvFrom is nil", namespace, appsody.Spec.EnvFrom[0].Prefix},
		{"Appsody Volumes is nil", volume.Name, appsody.Spec.Volumes[0].Name},
		{"Appsody VolumeMounts is nil", volumeMount.MountPath, appsody.Spec.VolumeMounts[0].MountPath},
		{"Appsody ResourceConstraints is nil", resourceContraints, appsody.Spec.ResourceConstraints},
		{"Appsody and Defaults ResourceConstraints is nil", reflect.TypeOf(corev1.ResourceRequirements{}), reflect.TypeOf(defResConNil)},
		{"Appsody Autoscaling is nil", autoscaling.MaxReplicas, appsody.Spec.Autoscaling.MaxReplicas},
		{"Appsody Expose is nil", expose, *appsody.Spec.Expose},
		{"Appsody CreateKnativeService is nil", createKNS, *appsody.Spec.CreateKnativeService},
		{"Appsody Service Type is nil", *service.Type, *appsody.Spec.Service.Type},
		{"Appsody Service Port is nil", int32(service.Port), int32(appsody.Spec.Service.Port)},
		{"Appsody and Defaults Service Type is nil", corev1.ServiceTypeClusterIP, servType},
		{"Appsody and Defaults Service Port is nil", int32(8080), int32(servPort)},
	}
	verifyTests(testIAV, t)
}

func TestApplyConstants(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	emptyService := &appsodyv1alpha1.AppsodyApplicationService{Port: 0}
	appsody := createAppsodyApp(name, namespace, appsodyv1alpha1.AppsodyApplicationSpec{Service: emptyService})
	defaults := appsodyv1alpha1.AppsodyApplicationSpec{}
	constants := &appsodyv1alpha1.AppsodyApplicationSpec{
		Replicas:             &replicas,
		Stack:                stack,
		ApplicationImage:     appImage,
		PullPolicy:           &pullPolicy,
		PullSecret:           &pullSecret,
		Expose:               &expose,
		CreateKnativeService: &createKNS,
		ServiceAccountName:   &serviceAccountName,
		Architecture:         arch,
		ReadinessProbe:       readinessProbe,
		LivenessProbe:        livenessProbe,
		EnvFrom:              envFrom,
		Env:                  env,
		Volumes:              []corev1.Volume{volume},
		VolumeMounts:         []corev1.VolumeMount{volumeMount},
		ResourceConstraints:  resourceContraints,
		Service:              service,
		Autoscaling:          autoscaling,
	}
	applyConstants(appsody, defaults, constants)
	// if cond in for len of envFrom, Env, Volumes, and VolumeMounts should stay the same
	envFromBefore := len(appsody.Spec.EnvFrom)
	envBefore := len(appsody.Spec.Env)
	volumesBefore := len(appsody.Spec.Volumes)
	volumeMountBefore := len(appsody.Spec.VolumeMounts)
	applyConstants(appsody, defaults, constants)
	envFromAfter := len(appsody.Spec.EnvFrom)
	envAfter := len(appsody.Spec.Env)
	volumesAfter := len(appsody.Spec.Volumes)
	volumeMountAfter := len(appsody.Spec.VolumeMounts)

	testAC := []Test{
		{"Constants Replicas", replicas, *appsody.Spec.Replicas},
		{"Constants Stack", stack, appsody.Spec.Stack},
		{"Constants ApplicationImage", appImage, appsody.Spec.ApplicationImage},
		{"Constants PullPolicy", pullPolicy, *appsody.Spec.PullPolicy},
		{"Constants PullSecret", pullSecret, *appsody.Spec.PullSecret},
		{"Constants Expose", expose, *appsody.Spec.Expose},
		{"Constants CreateKnativeService", createKNS, *appsody.Spec.CreateKnativeService},
		{"Constants ServiceAccountName", serviceAccountName, *appsody.Spec.ServiceAccountName},
		{"Constants ReadinessProbe", readinessProbe, appsody.Spec.ReadinessProbe},
		{"Constants LivenessProbe", livenessProbe, appsody.Spec.LivenessProbe},
		{"Constants EnvFrom", namespace, appsody.Spec.EnvFrom[0].Prefix},
		{"Constants Env", namespace, appsody.Spec.Env[0].Name},
		{"Constants Volumes", volume.Name, appsody.Spec.Volumes[0].Name},
		{"Constants VolumeMounts", volumeMount.MountPath, appsody.Spec.VolumeMounts[0].MountPath},
		{"Constants ResourceConstraints", resourceContraints, appsody.Spec.ResourceConstraints},
		{"Constants ServiceType", *service.Type, *appsody.Spec.Service.Type},
		{"Constants ServicePort", int32(service.Port), int32(appsody.Spec.Service.Port)},
		{"Constants Autoscaling", autoscaling.MaxReplicas, appsody.Spec.Autoscaling.MaxReplicas},
		{"Constants EnvFrom Found", envFromBefore, envFromAfter},
		{"Constants Env Found", envBefore, envAfter},
		{"Constants Volumes Found", volumesBefore, volumesAfter},
		{"Constants VolumeMount Found", volumeMountBefore, volumeMountAfter},
	}
	verifyTests(testAC, t)
}

func TestGetCondition(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	status := &appsodyv1alpha1.AppsodyApplicationStatus{
		Conditions: []appsodyv1alpha1.StatusCondition{
			{
				Status: corev1.ConditionTrue,
				Type:   appsodyv1alpha1.StatusConditionTypeReconciled,
			},
		},
	}
	conditionType := appsodyv1alpha1.StatusConditionTypeReconciled
	cond := GetCondition(conditionType, status)
	testGC := []Test{{"Set status condition", status.Conditions[0].Status, cond.Status}}
	verifyTests(testGC, t)
}

func TestSetCondition(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	status := &appsodyv1alpha1.AppsodyApplicationStatus{
		Conditions: []appsodyv1alpha1.StatusCondition{
			{Type: appsodyv1alpha1.StatusConditionTypeReconciled},
		},
	}
	condition := appsodyv1alpha1.StatusCondition{
		Status: corev1.ConditionTrue,
		Type:   appsodyv1alpha1.StatusConditionTypeReconciled,
	}
	SetCondition(condition, status)
	testSC := []Test{{"Set status condition", condition.Status, status.Conditions[0].Status}}
	verifyTests(testSC, t)
}

// Helper Functions
func createAppsodyApp(n, ns string, spec appsodyv1alpha1.AppsodyApplicationSpec) *appsodyv1alpha1.AppsodyApplication {
	app := &appsodyv1alpha1.AppsodyApplication{
		ObjectMeta: metav1.ObjectMeta{Name: n, Namespace: ns},
		Spec:       spec,
	}
	return app
}

func verifyTests(tests []Test, t *testing.T) {
	for _, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("%s test expected: (%v) actual: (%v)", tt.test, tt.expected, tt.actual)
		}
	}
}
