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

	// +kubebuilder:validation:Minimum=1
	MaxReplicas int32 `json:"maxReplicas,omitempty"`
}

// AppsodyApplicationService ...
// +k8s:openapi-gen=true
type AppsodyApplicationService struct {
	Type *corev1.ServiceType `json:"type,omitempty"`

	// +kubebuilder:validation:Maximum=65536
	// +kubebuilder:validation:Minimum=1
	Port int32 `json:"port,omitempty"`

	Annotations map[string]string `json:"annotations,omitempty"`
}

// AppsodyApplicationStorage ...
// +k8s:openapi-gen=true
type AppsodyApplicationStorage struct {
	Size                string                        `json:"size,omitempty"`
	MountPath           string                        `json:"mountPath,omitempty"`
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

// AppsodyApplicationStatus defines the observed state of AppsodyApplication
// +k8s:openapi-gen=true
type AppsodyApplicationStatus struct {
	Conditions []StatusCondition `json:"conditions,omitempty"`
}

// StatusCondition ...
// +k8s:openapi-gen=true
type StatusCondition struct {
	LastTransitionTime *metav1.Time           `json:"lastTransitionTime,omitempty"`
	LastUpdateTime     metav1.Time            `json:"lastUpdateTime,omitempty"`
	Reason             string                 `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
	Status             corev1.ConditionStatus `json:"status,omitempty"`
	Type               StatusConditionType    `json:"type,omitempty"`
}

// StatusConditionType ...
type StatusConditionType string

const (
	// StatusConditionTypeReconciled ...
	StatusConditionTypeReconciled StatusConditionType = "Reconciled"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppsodyApplication is the Schema for the appsodyapplications API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.applicationImage",priority="0",description="Absolute name of the deployed image containing registry and tag"
// +kubebuilder:printcolumn:name="Exposed",type="boolean",JSONPath=".spec.expose",priority="0",description="Specifies whether deployment is exposed externally via default Route"
// +kubebuilder:printcolumn:name="Reconciled",type="string",JSONPath=".status.conditions[?(@.type=='Reconciled')].status",priority="0",description="Status of the reconcile condition"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.conditions[?(@.type=='Reconciled')].reason",priority="1",description="Reason for the failure of reconcile condition"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[?(@.type=='Reconciled')].message",priority="1",description="Failure message from reconcile condition"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",priority="0",description="Age of the resource"
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
