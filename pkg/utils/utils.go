package utils

import (
	"fmt"
	"strings"

	prometheusv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// GetLabels ...
func GetLabels(cr *appsodyv1beta1.AppsodyApplication) map[string]string {
	labels := map[string]string{
		"app.kubernetes.io/name":       cr.Name,
		"app.kubernetes.io/managed-by": "appsody-operator",
	}

	if cr.Spec.Stack != "" {
		labels["app.appsody.dev/stack"] = cr.Spec.Stack
	}

	if cr.Spec.Version != "" {
		labels["app.kubernetes.io/version"] = cr.Spec.Version
	}

	for key, value := range cr.Labels {
		if key != "app.kubernetes.io/name" {
			labels[key] = value
		}
	}

	return labels
}

// CustomizeDeployment ...
func CustomizeDeployment(deploy *appsv1.Deployment, cr *appsodyv1beta1.AppsodyApplication) {
	deploy.Labels = GetLabels(cr)

	deploy.Spec.Replicas = cr.Spec.Replicas

	deploy.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app.kubernetes.io/name": cr.Name,
		},
	}

	if deploy.Annotations == nil {
		deploy.Annotations = make(map[string]string)
	}
	UpdateAppDefinition(deploy.Labels, deploy.Annotations, cr)
}

// CustomizeStatefulSet ...
func CustomizeStatefulSet(statefulSet *appsv1.StatefulSet, cr *appsodyv1beta1.AppsodyApplication) {
	statefulSet.Labels = GetLabels(cr)
	statefulSet.Spec.Replicas = cr.Spec.Replicas
	statefulSet.Spec.ServiceName = cr.Name + "-headless"
	statefulSet.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app.kubernetes.io/name": cr.Name,
		},
	}

	if statefulSet.Annotations == nil {
		statefulSet.Annotations = make(map[string]string)
	}
	UpdateAppDefinition(statefulSet.Labels, statefulSet.Annotations, cr)
}

// UpdateAppDefinition ...
func UpdateAppDefinition(labels map[string]string, annotations map[string]string, cr *appsodyv1beta1.AppsodyApplication) {
	if cr.Spec.CreateAppDefinition != nil && !*cr.Spec.CreateAppDefinition {
		delete(labels, "kappnav.app.auto-create")
		delete(annotations, "kappnav.app.auto-create.name")
		delete(annotations, "kappnav.app.auto-create.kinds")
		delete(annotations, "kappnav.app.auto-create.label")
		delete(annotations, "kappnav.app.auto-create.labels-values")
		delete(annotations, "kappnav.app.auto-create.version")
	} else {
		labels["kappnav.app.auto-create"] = "true"
		annotations["kappnav.app.auto-create.name"] = cr.Name
		annotations["kappnav.app.auto-create.kinds"] = "Deployment, StatefulSet, Service, Route, Ingress, ConfigMap"
		annotations["kappnav.app.auto-create.label"] = "app.kubernetes.io/name"
		annotations["kappnav.app.auto-create.labels-values"] = cr.Name
		if cr.Spec.Version == "" {
			delete(annotations, "kappnav.app.auto-create.version")
		} else {
			annotations["kappnav.app.auto-create.version"] = cr.Spec.Version
		}
	}
}

// CustomizeRoute ...
func CustomizeRoute(route *routev1.Route, cr *appsodyv1beta1.AppsodyApplication) {
	route.Labels = GetLabels(cr)
	route.Spec.To.Kind = "Service"
	route.Spec.To.Name = cr.Name
	weight := int32(100)
	route.Spec.To.Weight = &weight
	if route.Spec.Port == nil {
		route.Spec.Port = &routev1.RoutePort{}
	}
	route.Spec.Port.TargetPort = intstr.FromInt(int(cr.Spec.Service.Port))
}

// ErrorIsNoMatchesForKind ...
func ErrorIsNoMatchesForKind(err error, kind string, version string) bool {
	return strings.HasPrefix(err.Error(), fmt.Sprintf("no matches for kind \"%s\" in version \"%s\"", kind, version))
}

// CustomizeService ...
func CustomizeService(svc *corev1.Service, cr *appsodyv1beta1.AppsodyApplication) {
	svc.Labels = GetLabels(cr)

	if len(svc.Spec.Ports) == 0 {
		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{})
	}
	svc.Spec.Ports[0].Port = cr.Spec.Service.Port
	svc.Spec.Ports[0].TargetPort = intstr.FromInt(int(cr.Spec.Service.Port))
	svc.Spec.Type = *cr.Spec.Service.Type
	svc.Spec.Selector = map[string]string{
		"app.kubernetes.io/name": cr.Name,
	}
}

// CustomizePodSpec ...
func CustomizePodSpec(pts *corev1.PodTemplateSpec, cr *appsodyv1beta1.AppsodyApplication) {
	pts.Labels = GetLabels(cr)
	if len(pts.Spec.Containers) == 0 {
		pts.Spec.Containers = append(pts.Spec.Containers, corev1.Container{})
	}
	pts.Spec.Containers[0].Name = "app"
	if len(pts.Spec.Containers[0].Ports) == 0 {
		pts.Spec.Containers[0].Ports = append(pts.Spec.Containers[0].Ports, corev1.ContainerPort{})
	}
	pts.Spec.Containers[0].Ports[0].ContainerPort = cr.Spec.Service.Port
	pts.Spec.Containers[0].Image = cr.Spec.ApplicationImage
	pts.Spec.Containers[0].Resources = *cr.Spec.ResourceConstraints
	pts.Spec.Containers[0].ReadinessProbe = cr.Spec.ReadinessProbe
	pts.Spec.Containers[0].LivenessProbe = cr.Spec.LivenessProbe
	pts.Spec.Containers[0].VolumeMounts = cr.Spec.VolumeMounts
	pts.Spec.Containers[0].ImagePullPolicy = *cr.Spec.PullPolicy
	pts.Spec.Containers[0].Env = cr.Spec.Env
	pts.Spec.Containers[0].EnvFrom = cr.Spec.EnvFrom
	pts.Spec.Volumes = cr.Spec.Volumes

	if cr.Spec.ServiceAccountName != nil && *cr.Spec.ServiceAccountName != "" {
		pts.Spec.ServiceAccountName = *cr.Spec.ServiceAccountName
	} else {
		pts.Spec.ServiceAccountName = cr.Name
	}
	pts.Spec.RestartPolicy = corev1.RestartPolicyAlways
	pts.Spec.DNSPolicy = corev1.DNSClusterFirst

	if len(cr.Spec.Architecture) > 0 {
		pts.Spec.Affinity = &corev1.Affinity{}
		CustomizeAffinity(pts.Spec.Affinity, cr)
	}
}

// CustomizePersistence ...
func CustomizePersistence(statefulSet *appsv1.StatefulSet, cr *appsodyv1beta1.AppsodyApplication) {
	if len(statefulSet.Spec.VolumeClaimTemplates) == 0 {
		var pvc *corev1.PersistentVolumeClaim
		if cr.Spec.Storage.VolumeClaimTemplate != nil {
			pvc = cr.Spec.Storage.VolumeClaimTemplate
		} else {
			pvc = &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pvc",
					Namespace: cr.Namespace,
					Labels:    GetLabels(cr),
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(cr.Spec.Storage.Size),
						},
					},
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
				},
			}

		}
		statefulSet.Spec.VolumeClaimTemplates = append(statefulSet.Spec.VolumeClaimTemplates, *pvc)
	}

	if cr.Spec.Storage.MountPath != "" {
		found := false
		for _, v := range statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts {
			if v.Name == statefulSet.Spec.VolumeClaimTemplates[0].Name {
				found = true
			}
		}

		if !found {
			vm := corev1.VolumeMount{
				Name:      statefulSet.Spec.VolumeClaimTemplates[0].Name,
				MountPath: cr.Spec.Storage.MountPath,
			}
			statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts, vm)
		}
	}

}

// CustomizeServiceAccount ...
func CustomizeServiceAccount(sa *corev1.ServiceAccount, cr *appsodyv1beta1.AppsodyApplication) {
	sa.Labels = GetLabels(cr)
	if cr.Spec.PullSecret != nil {
		if len(sa.ImagePullSecrets) == 0 {
			sa.ImagePullSecrets = append(sa.ImagePullSecrets, corev1.LocalObjectReference{
				Name: *cr.Spec.PullSecret,
			})
		} else {
			sa.ImagePullSecrets[0].Name = *cr.Spec.PullSecret
		}
	}
}

// CustomizeAffinity ...
func CustomizeAffinity(a *corev1.Affinity, cr *appsodyv1beta1.AppsodyApplication) {

	a.NodeAffinity = &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Operator: corev1.NodeSelectorOpIn,
							Values:   cr.Spec.Architecture,
							Key:      "beta.kubernetes.io/arch",
						},
					},
				},
			},
		},
	}

	archs := len(cr.Spec.Architecture)
	for i := range cr.Spec.Architecture {
		arch := cr.Spec.Architecture[i]
		term := corev1.PreferredSchedulingTerm{
			Weight: int32(archs - i),
			Preference: corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{arch},
						Key:      "beta.kubernetes.io/arch",
					},
				},
			},
		}
		a.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(a.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, term)
	}
}

// CustomizeKnativeService ...
func CustomizeKnativeService(ksvc *servingv1alpha1.Service, cr *appsodyv1beta1.AppsodyApplication) {
	ksvc.Labels = GetLabels(cr)

	// If `expose` is not set to `true`, make Knative route a private route by adding `serving.knative.dev/visibility: cluster-local`
	// to the Knative service. If `serving.knative.dev/visibility: XYZ` is defined in cr.Labels, `expose` always wins.
	if cr.Spec.Expose != nil && *cr.Spec.Expose {
		delete(ksvc.Labels, "serving.knative.dev/visibility")
	} else {
		ksvc.Labels["serving.knative.dev/visibility"] = "cluster-local"
	}

	if ksvc.Spec.Template == nil {
		ksvc.Spec.Template = &servingv1alpha1.RevisionTemplateSpec{}
	}
	if len(ksvc.Spec.Template.Spec.Containers) == 0 {
		ksvc.Spec.Template.Spec.Containers = append(ksvc.Spec.Template.Spec.Containers, corev1.Container{})
	}

	if len(ksvc.Spec.Template.Spec.Containers[0].Ports) == 0 {
		ksvc.Spec.Template.Spec.Containers[0].Ports = append(ksvc.Spec.Template.Spec.Containers[0].Ports, corev1.ContainerPort{})
	}
	ksvc.Spec.Template.ObjectMeta.Labels = GetLabels(cr)
	ksvc.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = cr.Spec.Service.Port
	ksvc.Spec.Template.Spec.Containers[0].Image = cr.Spec.ApplicationImage
	// Knative sets its own resource constraints
	//ksvc.Spec.Template.Spec.Containers[0].Resources = *cr.Spec.ResourceConstraints
	ksvc.Spec.Template.Spec.Containers[0].ReadinessProbe = cr.Spec.ReadinessProbe
	ksvc.Spec.Template.Spec.Containers[0].LivenessProbe = cr.Spec.LivenessProbe
	ksvc.Spec.Template.Spec.Containers[0].VolumeMounts = cr.Spec.VolumeMounts
	ksvc.Spec.Template.Spec.Containers[0].ImagePullPolicy = *cr.Spec.PullPolicy
	ksvc.Spec.Template.Spec.Containers[0].Env = cr.Spec.Env
	ksvc.Spec.Template.Spec.Containers[0].EnvFrom = cr.Spec.EnvFrom

	ksvc.Spec.Template.Spec.Volumes = cr.Spec.Volumes

	if cr.Spec.ServiceAccountName != nil && *cr.Spec.ServiceAccountName != "" {
		ksvc.Spec.Template.Spec.ServiceAccountName = *cr.Spec.ServiceAccountName
	} else {
		ksvc.Spec.Template.Spec.ServiceAccountName = cr.Name
	}

	if ksvc.Spec.Template.Spec.Containers[0].LivenessProbe != nil {
		if ksvc.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet != nil {
			ksvc.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Port = intstr.IntOrString{}
		}
		if ksvc.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket != nil {
			ksvc.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket.Port = intstr.IntOrString{}
		}
	}

	if ksvc.Spec.Template.Spec.Containers[0].ReadinessProbe != nil {
		if ksvc.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet != nil {
			ksvc.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port = intstr.IntOrString{}
		}
		if ksvc.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket != nil {
			ksvc.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket.Port = intstr.IntOrString{}
		}
	}
}

// CustomizeHPA ...
func CustomizeHPA(hpa *autoscalingv1.HorizontalPodAutoscaler, cr *appsodyv1beta1.AppsodyApplication) {
	hpa.Labels = GetLabels(cr)

	hpa.Spec.MaxReplicas = cr.Spec.Autoscaling.MaxReplicas
	hpa.Spec.MinReplicas = cr.Spec.Autoscaling.MinReplicas
	hpa.Spec.TargetCPUUtilizationPercentage = cr.Spec.Autoscaling.TargetCPUUtilizationPercentage

	hpa.Spec.ScaleTargetRef.Name = cr.Name
	hpa.Spec.ScaleTargetRef.APIVersion = "apps/v1"

	if cr.Spec.Storage != nil {
		hpa.Spec.ScaleTargetRef.Kind = "StatefulSet"
	} else {
		hpa.Spec.ScaleTargetRef.Kind = "Deployment"
	}
}

// InitAndValidate ...
func InitAndValidate(cr *appsodyv1beta1.AppsodyApplication, defaults appsodyv1beta1.AppsodyApplicationSpec, constants *appsodyv1beta1.AppsodyApplicationSpec) {

	if cr.Spec.PullPolicy == nil {
		cr.Spec.PullPolicy = defaults.PullPolicy
		if cr.Spec.PullPolicy == nil {
			pp := corev1.PullIfNotPresent
			cr.Spec.PullPolicy = &pp
		}
	}

	if cr.Spec.PullSecret == nil {
		cr.Spec.PullSecret = defaults.PullSecret
	}

	if cr.Spec.ServiceAccountName == nil {
		cr.Spec.ServiceAccountName = defaults.ServiceAccountName
	}

	if cr.Spec.ReadinessProbe == nil {
		cr.Spec.ReadinessProbe = defaults.ReadinessProbe
	}
	if cr.Spec.LivenessProbe == nil {
		cr.Spec.LivenessProbe = defaults.LivenessProbe
	}
	if cr.Spec.Env == nil {
		cr.Spec.Env = defaults.Env
	}
	if cr.Spec.EnvFrom == nil {
		cr.Spec.EnvFrom = defaults.EnvFrom
	}

	if cr.Spec.Volumes == nil {
		cr.Spec.Volumes = defaults.Volumes
	}

	if cr.Spec.VolumeMounts == nil {
		cr.Spec.VolumeMounts = defaults.VolumeMounts
	}

	if cr.Spec.ResourceConstraints == nil {
		if defaults.ResourceConstraints != nil {
			cr.Spec.ResourceConstraints = defaults.ResourceConstraints
		} else {
			cr.Spec.ResourceConstraints = &corev1.ResourceRequirements{}
		}
	}

	if cr.Spec.Autoscaling == nil {
		cr.Spec.Autoscaling = defaults.Autoscaling
	}

	if cr.Spec.Expose == nil {
		cr.Spec.Expose = defaults.Expose
	}

	if cr.Spec.CreateKnativeService == nil {
		cr.Spec.CreateKnativeService = defaults.CreateKnativeService
	}

	if cr.Spec.Service == nil {
		cr.Spec.Service = defaults.Service
	}

	// This is to handle when there is no service in the CR nor defaults
	if cr.Spec.Service == nil {
		cr.Spec.Service = &appsodyv1beta1.AppsodyApplicationService{}
	}

	if cr.Spec.Service.Type == nil {
		st := corev1.ServiceTypeClusterIP
		cr.Spec.Service.Type = &st
	}
	if cr.Spec.Service.Port == 0 {
		if defaults.Service != nil && defaults.Service.Port != 0 {
			cr.Spec.Service.Port = defaults.Service.Port
		} else {
			cr.Spec.Service.Port = 8080
		}
	}

	if constants != nil {
		applyConstants(cr, defaults, constants)
	}
}

func applyConstants(cr *appsodyv1beta1.AppsodyApplication, defaults appsodyv1beta1.AppsodyApplicationSpec, constants *appsodyv1beta1.AppsodyApplicationSpec) {

	if constants.Replicas != nil {
		cr.Spec.Replicas = constants.Replicas
	}

	if constants.Stack != "" {
		cr.Spec.Stack = constants.Stack
	}

	if constants.ApplicationImage != "" {
		cr.Spec.ApplicationImage = constants.ApplicationImage
	}

	if constants.PullPolicy != nil {
		cr.Spec.PullPolicy = constants.PullPolicy
	}

	if constants.PullSecret != nil {
		cr.Spec.PullSecret = constants.PullSecret
	}

	if constants.Expose != nil {
		cr.Spec.Expose = constants.Expose
	}

	if constants.CreateKnativeService != nil {
		cr.Spec.CreateKnativeService = constants.CreateKnativeService
	}

	if constants.ServiceAccountName != nil {
		cr.Spec.ServiceAccountName = constants.ServiceAccountName
	}

	if constants.Architecture != nil {
		cr.Spec.Architecture = constants.Architecture
	}

	if constants.ReadinessProbe != nil {
		cr.Spec.ReadinessProbe = constants.ReadinessProbe
	}

	if constants.LivenessProbe != nil {
		cr.Spec.LivenessProbe = constants.LivenessProbe
	}

	if constants.EnvFrom != nil {
		for _, v := range constants.EnvFrom {

			found := false
			for _, v2 := range cr.Spec.EnvFrom {
				if v2 == v {
					found = true
				}
			}
			if !found {
				cr.Spec.EnvFrom = append(cr.Spec.EnvFrom, v)
			}
		}
	}

	if constants.Env != nil {
		for _, v := range constants.Env {
			found := false
			for _, v2 := range cr.Spec.Env {
				if v2.Name == v.Name {
					found = true
				}
			}
			if !found {
				cr.Spec.Env = append(cr.Spec.Env, v)
			}
		}
	}

	if constants.Volumes != nil {
		for _, v := range constants.Volumes {
			found := false
			for _, v2 := range cr.Spec.Volumes {
				if v2.Name == v.Name {
					found = true
				}
			}
			if !found {
				cr.Spec.Volumes = append(cr.Spec.Volumes, v)
			}
		}
	}

	if constants.VolumeMounts != nil {
		for _, v := range constants.VolumeMounts {
			found := false
			for _, v2 := range cr.Spec.VolumeMounts {
				if v2.Name == v.Name {
					found = true
				}
			}
			if !found {
				cr.Spec.VolumeMounts = append(cr.Spec.VolumeMounts, v)
			}
		}
	}

	if constants.ResourceConstraints != nil {
		cr.Spec.ResourceConstraints = constants.ResourceConstraints
	}

	if constants.Service != nil {
		if constants.Service.Type != nil {
			cr.Spec.Service.Type = constants.Service.Type
		}
		if constants.Service.Port != 0 {
			cr.Spec.Service.Port = constants.Service.Port
		}
	}

	if constants.Autoscaling != nil {
		cr.Spec.Autoscaling = constants.Autoscaling
	}
}

// CustomizeServiceMonitor ...
func CustomizeServiceMonitor(sm *prometheusv1.ServiceMonitor, cr *appsodyv1beta1.AppsodyApplication) {
	sm.Labels = GetLabels(cr)
	sm.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app.kubernetes.io/name":  cr.Name,
			"app.appsody.dev/monitor": "true",
		},
	}
	if len(sm.Spec.Endpoints) == 0 {
		sm.Spec.Endpoints = append(sm.Spec.Endpoints, prometheusv1.Endpoint{})
	}
	targetPort := intstr.FromInt(int(cr.Spec.Service.Port))
	sm.Spec.Endpoints[0].TargetPort = &targetPort
	if len(cr.Spec.Monitoring.Labels) > 0 {
		for k, v := range cr.Spec.Monitoring.Labels {
			sm.Labels[k] = v
		}
	}

	if len(cr.Spec.Monitoring.Endpoints) > 0 {

		if cr.Spec.Monitoring.Endpoints[0].Scheme != "" {
			sm.Spec.Endpoints[0].Scheme = cr.Spec.Monitoring.Endpoints[0].Scheme
		}
		if cr.Spec.Monitoring.Endpoints[0].Interval != "" {
			sm.Spec.Endpoints[0].Interval = cr.Spec.Monitoring.Endpoints[0].Interval
		}
		if cr.Spec.Monitoring.Endpoints[0].Path != "" {
			sm.Spec.Endpoints[0].Path = cr.Spec.Monitoring.Endpoints[0].Path
		}

		if cr.Spec.Monitoring.Endpoints[0].TLSConfig != nil {
			sm.Spec.Endpoints[0].TLSConfig = cr.Spec.Monitoring.Endpoints[0].TLSConfig
		}

		if cr.Spec.Monitoring.Endpoints[0].BasicAuth != nil {
			sm.Spec.Endpoints[0].BasicAuth = cr.Spec.Monitoring.Endpoints[0].BasicAuth
		}

		if cr.Spec.Monitoring.Endpoints[0].Params != nil {
			sm.Spec.Endpoints[0].Params = cr.Spec.Monitoring.Endpoints[0].Params
		}
	}

}

// GetCondition ...
func GetCondition(conditionType appsodyv1beta1.StatusConditionType, status *appsodyv1beta1.AppsodyApplicationStatus) *appsodyv1beta1.StatusCondition {
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return &status.Conditions[i]
		}
	}

	return nil
}

// SetCondition ...
func SetCondition(condition appsodyv1beta1.StatusCondition, status *appsodyv1beta1.AppsodyApplicationStatus) {
	for i := range status.Conditions {
		if status.Conditions[i].Type == condition.Type {
			status.Conditions[i] = condition
			return
		}
	}

	status.Conditions = append(status.Conditions, condition)
}

// GetWatchNamespaces returns a slice of namespaces the operator should watch based on WATCH_NAMESPSCE value
// WATCH_NAMESPSCE value could be empty for watching the whole cluster or a comma-separated list of namespaces
func GetWatchNamespaces() ([]string, error) {
	watchNamespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		return nil, err
	}

	var watchNamespaces []string
	for _, ns := range strings.Split(watchNamespace, ",") {
		watchNamespaces = append(watchNamespaces, strings.TrimSpace(ns))
	}

	return watchNamespaces, nil
}
