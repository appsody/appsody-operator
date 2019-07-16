package utils

import (
	"fmt"
	"strings"

	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//GenerateDeployment ...
func GenerateDeployment(cr *appsodyv1alpha1.AppsodyApplication) *appsv1.Deployment {
	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    getLabels(cr),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: cr.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": cr.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cr.Name,
					Namespace: cr.Namespace,
					Labels:    getLabels(cr),
				},
				Spec: GeneratePodSpec(cr),
			},
		},
	}
	return &deploy
}

//GenerateStatefulSet ...
func GenerateStatefulSet(cr *appsodyv1alpha1.AppsodyApplication) *appsv1.StatefulSet {
	statefulSet := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    getLabels(cr),
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: cr.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": cr.Name,
				},
			},
			ServiceName: cr.Name + "-headless",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cr.Name,
					Namespace: cr.Namespace,
					Labels:    getLabels(cr),
				},
				Spec: GeneratePodSpec(cr),
			},
		},
	}
	if cr.Spec.Storage != nil {
		if cr.Spec.Storage.VolumeClaimTemplate != nil {
			statefulSet.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
				*cr.Spec.Storage.VolumeClaimTemplate,
			}
		} else {

			storageSize, ok := resource.ParseQuantity(cr.Spec.Storage.Size)
			if ok != nil {
				//TODO
			}
			statefulSet.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      cr.Name,
						Namespace: cr.Namespace,
						Labels:    getLabels(cr),
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: storageSize,
							},
						},
					},
				},
			}
		}
	}
	return &statefulSet
}

// GenerateSeviceAccount ...
func GenerateSeviceAccount(cr *appsodyv1alpha1.AppsodyApplication) *corev1.ServiceAccount {

	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    getLabels(cr),
		},
		ImagePullSecrets: []corev1.LocalObjectReference{},
	}

	if cr.Spec.PullSecret != "" {
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, corev1.LocalObjectReference{
			Name: cr.Spec.PullSecret,
		})
	}
	return &sa
}

// GenerateService ...
func GenerateService(cr *appsodyv1alpha1.AppsodyApplication) *corev1.Service {
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    getLabels(cr),
		},
		Spec: corev1.ServiceSpec{
			Type: cr.Spec.Service.Type,
			Selector: map[string]string{
				"app.kubernetes.io/name": cr.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Port: cr.Spec.Service.Port,
				},
			},
		},
	}

	return &service
}

func getLabels(cr *appsodyv1alpha1.AppsodyApplication) map[string]string {
	labels := map[string]string{
		"app.kubernetes.io/name":       cr.Name,
		"app.kubernetes.io/managed-by": "appsody-operator",
	}
	return labels
}

// GeneratePodSpec ...
func GeneratePodSpec(cr *appsodyv1alpha1.AppsodyApplication) corev1.PodSpec {
	pod := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "app",
				Image: cr.Spec.ApplicationImage,
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: cr.Spec.Service.Port,
					},
				},
				Resources:       cr.Spec.ResourceConstraints,
				ReadinessProbe:  cr.Spec.ReadinessProbe,
				LivenessProbe:   cr.Spec.LivenessProbe,
				Env:             cr.Spec.Env,
				EnvFrom:         cr.Spec.EnvFrom,
				VolumeMounts:    cr.Spec.VolumeMounts,
				ImagePullPolicy: cr.Spec.PullPolicy,
			},
		},
		Volumes:            cr.Spec.Volumes,
		RestartPolicy:      corev1.RestartPolicyAlways,
		DNSPolicy:          corev1.DNSClusterFirst,
		ServiceAccountName: cr.Name,
	}
	if cr.Spec.ServiceAccountName != "" {
		pod.ServiceAccountName = cr.Spec.ServiceAccountName
	}
	return pod
}

// GenerateHPA ...
func GenerateHPA(cr *appsodyv1alpha1.AppsodyApplication) autoscalingv1.HorizontalPodAutoscaler {
	hpa := autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    getLabels(cr),
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			MaxReplicas:                    *cr.Spec.Autoscaling.MaxReplicas,
			MinReplicas:                    cr.Spec.Autoscaling.MinReplicas,
			TargetCPUUtilizationPercentage: cr.Spec.Autoscaling.TargetCPUUtilizationPercentage,
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       cr.Name,
			},
		},
	}
	return hpa
}

// GenerateHeadlessSvc ...
func GenerateHeadlessSvc(cr *appsodyv1alpha1.AppsodyApplication) *corev1.Service {
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-headless",
			Namespace: cr.Namespace,
			Labels:    getLabels(cr),
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Selector: map[string]string{
				"app.kubernetes.io/name": cr.Name,
			},
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Port: cr.Spec.Service.Port,
				},
			},
		},
	}
	return &svc
}

// GenerateRoute ...
func GenerateRoute(cr *appsodyv1alpha1.AppsodyApplication) *routev1.Route {
	weight := int32(100)
	return &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "route.openshift.io/v1",
			Kind:       "Route",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    getLabels(cr),
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   cr.Name,
				Weight: &weight,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(int(cr.Spec.Service.Port)),
			},
		},
	}
}

// GenerateIngress ...
func GenerateIngress(cr *appsodyv1alpha1.AppsodyApplication) *extv1beta1.Ingress {
	return &extv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    getLabels(cr),
		},
		Spec: extv1beta1.IngressSpec{
			Rules: []extv1beta1.IngressRule{
				{
					IngressRuleValue: extv1beta1.IngressRuleValue{
						HTTP: &extv1beta1.HTTPIngressRuleValue{
							Paths: []extv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extv1beta1.IngressBackend{
										ServiceName: cr.Name,
										ServicePort: intstr.FromInt(int(cr.Spec.Service.Port)),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// ErrorIsNoMatchesForKind ...
func ErrorIsNoMatchesForKind(err error, kind string, version string) bool {
	return strings.HasPrefix(err.Error(), fmt.Sprintf("no matches for kind \"%s\" in version \"%s\"", kind, version))
}
