package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppsodyApplicationSpec defines the desired state of AppsodyApplication
// +k8s:openapi-gen=true
type AppsodyApplicationSpec struct {
	ApplicationImage     string                         `json:"applicationImage"`
	Replicas             *int32                         `json:"replicas,omitempty"`
	Autoscaling          *AppsodyApplicationAutoScaling `json:"autoscaling,omitempty"`
	PullPolicy           *corev1.PullPolicy             `json:"pullPolicy,omitempty"`
	PullSecret           *string                        `json:"pullSecret,omitempty"`
	Volumes              []corev1.Volume                `json:"volumes,omitempty"`
	VolumeMounts         []corev1.VolumeMount           `json:"volumeMounts,omitempty"`
	ResourceConstraints  *corev1.ResourceRequirements   `json:"resourceConstraints,omitempty"`
	ReadinessProbe       *corev1.Probe                  `json:"readinessProbe,omitempty"`
	LivenessProbe        *corev1.Probe                  `json:"livenessProbe,omitempty"`
	Service              *AppsodyApplicationService     `json:"service,omitempty"`
	Expose               *bool                          `json:"expose,omitempty"`
	EnvFrom              []corev1.EnvFromSource         `json:"envFrom,omitempty"`
	Env                  []corev1.EnvVar                `json:"env,omitempty"`
	ServiceAccountName   *string                        `json:"serviceAccountName,omitempty"`
	Architecture         []string                       `json:"architecture,omitempty"`
	Storage              *AppsodyApplicationStorage     `json:"storage,omitempty"`
	CreateKnativeService *bool                          `json:"createKnativeService,omitempty"`
	Stack                string                         `json:"stack"`
}

// AppsodyApplicationAutoScaling ...
// +k8s:openapi-gen=true
type AppsodyApplicationAutoScaling struct {
	TargetCPUUtilizationPercentage *int32 `json:"targetCPUUtilizationPercentage,omitempty"`
	MinReplicas                    *int32 `json:"minReplicas,omitempty"`
	MaxReplicas                    *int32 `json:"maxReplicas,omitempty"`
}

// AppsodyApplicationService ...
// +k8s:openapi-gen=true
type AppsodyApplicationService struct {
	Type *corev1.ServiceType `json:"type,omitempty"`

	// +kubebuilder:validation:Maximum=65536
	// +kubebuilder:validation:Minimum=1
	Port int32 `json:"port,omitempty"`
}

// AppsodyApplicationStorage ...
// +k8s:openapi-gen=true
type AppsodyApplicationStorage struct {
	Size                string                        `json:"size,omitempty"`
	MountPath           string                        `json:"mountPath"`
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

// AppsodyApplicationStatus defines the observed state of AppsodyApplication
// +k8s:openapi-gen=true
type AppsodyApplicationStatus struct {
	Conditions []AppsodyApplicationStatusCondition `json:"conditions,omitempty"`
}

// AppsodyApplicationStatusCondition ...
type AppsodyApplicationStatusCondition struct {
	LastTransitionTime metav1.Time                           `json:"lastTransitionTime,omitempty"`
	LastUpdateTime     metav1.Time                           `json:"lastUpdateTime,omitempty"`
	Reason             string                                `json:"reason,omitempty"`
	Message            string                                `json:"mesage,omitempty"`
	Status             corev1.ConditionStatus                `json:"status,omitempty"`
	Type               AppsodyApplicationStatusConditionType `json:"type,omitempty"`
}

// AppsodyApplicationStatusConditionType ...
type AppsodyApplicationStatusConditionType string

const (
	// AppsodyApplicationStatusConditionTypeInitialized ...
	AppsodyApplicationStatusConditionTypeInitialized AppsodyApplicationStatusConditionType = "Initialized"

	// AppsodyApplicationStatusConditionTypeReconciled ...
	AppsodyApplicationStatusConditionTypeReconciled AppsodyApplicationStatusConditionType = "Reconciled"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppsodyApplication is the Schema for the appsodyapplications API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.applicationImage"
// +kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".spec.service.port"
// +kubebuilder:printcolumn:name="Exposed",type="boolean",JSONPath=".spec.expose"
type AppsodyApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppsodyApplicationSpec   `json:"spec,omitempty"`
	Status AppsodyApplicationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppsodyApplicationList contains a list of AppsodyApplication
type AppsodyApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppsodyApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppsodyApplication{}, &AppsodyApplicationList{})
}
