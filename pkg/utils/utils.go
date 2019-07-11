package utils

import (
	appsodyv1alpha1 "github.com/appsody-operator/pkg/apis/appsody/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
				Resources:      cr.Spec.ResourceConstraints,
				ReadinessProbe: cr.Spec.ReadinessProbe,
				LivenessProbe:  cr.Spec.LivenessProbe,
				Env:            cr.Spec.Env,
				EnvFrom:        cr.Spec.EnvFrom,
				VolumeMounts:   cr.Spec.VolumeMounts,
			},
		},
		Volumes:       cr.Spec.Volumes,
		RestartPolicy: corev1.RestartPolicyAlways,
		DNSPolicy:     corev1.DNSClusterFirst,
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
