package utils

import (
	"fmt"
	"strings"

	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// GetLabels ...
func GetLabels(cr *appsodyv1alpha1.AppsodyApplication) map[string]string {
	labels := map[string]string{
		"app.kubernetes.io/name":       cr.Name,
		"app.kubernetes.io/managed-by": "appsody-operator",
	}
	return labels
}

// CustomizeRoute ...
func CustomizeRoute(route *routev1.Route, cr *appsodyv1alpha1.AppsodyApplication) {
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
func CustomizeService(svc *corev1.Service, cr *appsodyv1alpha1.AppsodyApplication) {
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
func CustomizePodSpec(pts *corev1.PodTemplateSpec, cr *appsodyv1alpha1.AppsodyApplication) {
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
func CustomizePersistence(statefulSet *appsv1.StatefulSet, cr *appsodyv1alpha1.AppsodyApplication) {
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
}

// CustomizeServiceAccount ...
func CustomizeServiceAccount(sa *corev1.ServiceAccount, cr *appsodyv1alpha1.AppsodyApplication) {
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
func CustomizeAffinity(a *corev1.Affinity, cr *appsodyv1alpha1.AppsodyApplication) {

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
func CustomizeKnativeService(ksvc *servingv1alpha1.Service, cr *appsodyv1alpha1.AppsodyApplication) {
	ksvc.Labels = GetLabels(cr)

	if ksvc.Spec.Template == nil {
		ksvc.Spec.Template = &servingv1alpha1.RevisionTemplateSpec{}
	}
	if len(ksvc.Spec.Template.Spec.Containers) == 0 {
		ksvc.Spec.Template.Spec.Containers = append(ksvc.Spec.Template.Spec.Containers, corev1.Container{Name: "user-container"})
	}

	ksvc.Spec.Template.Spec.Containers[0].Name = "user-container"
	ksvc.Spec.Template.Spec.Containers[0].Image = cr.Spec.ApplicationImage
	ksvc.Spec.Template.Spec.Containers[0].Resources = *cr.Spec.ResourceConstraints
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
}

// InitAndValidate ...
func InitAndValidate(cr *appsodyv1alpha1.AppsodyApplication, defaults appsodyv1alpha1.AppsodyApplicationSpec) {

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

	if cr.Spec.Service.Type == nil {
		if defaults.Service.Type != nil {
			cr.Spec.Service.Type = defaults.Service.Type
		} else {
			st := corev1.ServiceTypeClusterIP
			cr.Spec.Service.Type = &st
		}
	}
	if cr.Spec.Service.Port == 0 {
		if defaults.Service.Port != 0 {
			cr.Spec.Service.Port = defaults.Service.Port
		} else {
			cr.Spec.Service.Port = 8080
		}
	}
}

// GetCondition ...
func GetCondition(conditionType appsodyv1alpha1.AppsodyApplicationStatusConditionType, status *appsodyv1alpha1.AppsodyApplicationStatus) *appsodyv1alpha1.AppsodyApplicationStatusCondition {
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return &status.Conditions[i]
		}
	}

	return nil
}

// SetCondition ...
func SetCondition(condition appsodyv1alpha1.AppsodyApplicationStatusCondition, status *appsodyv1alpha1.AppsodyApplicationStatus) {
	for i := range status.Conditions {
		if status.Conditions[i].Type == condition.Type {
			status.Conditions[i] = condition
			return
		}
	}

	status.Conditions = append(status.Conditions, condition)
}
