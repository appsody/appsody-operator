// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1beta1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplication":            schema_pkg_apis_appsody_v1beta1_AppsodyApplication(ref),
		"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationAutoScaling": schema_pkg_apis_appsody_v1beta1_AppsodyApplicationAutoScaling(ref),
		"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationService":     schema_pkg_apis_appsody_v1beta1_AppsodyApplicationService(ref),
		"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationSpec":        schema_pkg_apis_appsody_v1beta1_AppsodyApplicationSpec(ref),
		"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationStatus":      schema_pkg_apis_appsody_v1beta1_AppsodyApplicationStatus(ref),
		"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.ServiceBindingConsumes":        schema_pkg_apis_appsody_v1beta1_ServiceBindingConsumes(ref),
		"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.ServiceBindingProvides":        schema_pkg_apis_appsody_v1beta1_ServiceBindingProvides(ref),
		"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.StatusCondition":               schema_pkg_apis_appsody_v1beta1_StatusCondition(ref),
	}
}

func schema_pkg_apis_appsody_v1beta1_AppsodyApplication(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AppsodyApplication is the Schema for the appsodyapplications API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationSpec", "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_appsody_v1beta1_AppsodyApplicationAutoScaling(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AppsodyApplicationAutoScaling ...",
				Properties: map[string]spec.Schema{
					"targetCPUUtilizationPercentage": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"minReplicas": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"maxReplicas": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
				},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_appsody_v1beta1_AppsodyApplicationService(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AppsodyApplicationService ...",
				Properties: map[string]spec.Schema{
					"type": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"port": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"consumes": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-list-type": "atomic",
							},
						},
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.ServiceBindingConsumes"),
									},
								},
							},
						},
					},
					"annotations": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"string"},
										Format: "",
									},
								},
							},
						},
					},
					"protocol": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"provides": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.ServiceBindingProvides"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.ServiceBindingConsumes", "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.ServiceBindingProvides"},
	}
}

func schema_pkg_apis_appsody_v1beta1_AppsodyApplicationSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AppsodyApplicationSpec defines the desired state of AppsodyApplication",
				Properties: map[string]spec.Schema{
					"version": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"applicationImage": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"replicas": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"autoscaling": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationAutoScaling"),
						},
					},
					"pullPolicy": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"pullSecret": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"resourceConstraints": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/api/core/v1.ResourceRequirements"),
						},
					},
					"readinessProbe": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/api/core/v1.Probe"),
						},
					},
					"livenessProbe": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/api/core/v1.Probe"),
						},
					},
					"service": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationService"),
						},
					},
					"expose": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"serviceAccountName": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"storage": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationStorage"),
						},
					},
					"createKnativeService": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"stack": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"monitoring": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationMonitoring"),
						},
					},
					"createAppDefinition": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"volumeMounts": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-list-map-keys": "name",
								"x-kubernetes-list-type":     "map",
							},
						},
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.VolumeMount"),
									},
								},
							},
						},
					},
					"volumes": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-list-map-keys": "name",
								"x-kubernetes-list-type":     "map",
							},
						},
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.Volume"),
									},
								},
							},
						},
					},
					"envFrom": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-list-map-keys": "name",
								"x-kubernetes-list-type": []string{
									"map",
									"atomic",
								},
							},
						},
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.EnvFromSource"),
									},
								},
							},
						},
					},
					"env": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-list-map-keys": "name",
								"x-kubernetes-list-type":     "map",
							},
						},
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.EnvVar"),
									},
								},
							},
						},
					},
					"architecture": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-list-type": "set",
							},
						},
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"string"},
										Format: "",
									},
								},
							},
						},
					},
				},
				Required: []string{"applicationImage"},
			},
		},
		Dependencies: []string{
			"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationAutoScaling", "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationMonitoring", "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationService", "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.AppsodyApplicationStorage", "k8s.io/api/core/v1.EnvFromSource", "k8s.io/api/core/v1.EnvVar", "k8s.io/api/core/v1.Probe", "k8s.io/api/core/v1.ResourceRequirements", "k8s.io/api/core/v1.Volume", "k8s.io/api/core/v1.VolumeMount"},
	}
}

func schema_pkg_apis_appsody_v1beta1_AppsodyApplicationStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AppsodyApplicationStatus defines the observed state of AppsodyApplication",
				Properties: map[string]spec.Schema{
					"conditions": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-list-map-keys": "type",
								"x-kubernetes-list-type":     "map",
							},
						},
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.StatusCondition"),
									},
								},
							},
						},
					},
					"consumedServices": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type: []string{"array"},
										Items: &spec.SchemaOrArray{
											Schema: &spec.Schema{
												SchemaProps: spec.SchemaProps{
													Type:   []string{"string"},
													Format: "",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.StatusCondition"},
	}
}

func schema_pkg_apis_appsody_v1beta1_ServiceBindingConsumes(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceBindingConsumes represents a service to be consumed",
				Properties: map[string]spec.Schema{
					"name": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"namespace": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"category": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"mountPath": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
				Required: []string{"name", "category"},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_appsody_v1beta1_ServiceBindingProvides(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceBindingProvides represents information about",
				Properties: map[string]spec.Schema{
					"category": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"context": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"protocol": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"auth": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.ServiceBindingAuth"),
						},
					},
				},
				Required: []string{"category"},
			},
		},
		Dependencies: []string{
			"github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1.ServiceBindingAuth"},
	}
}

func schema_pkg_apis_appsody_v1beta1_StatusCondition(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "StatusCondition ...",
				Properties: map[string]spec.Schema{
					"lastTransitionTime": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.Time"),
						},
					},
					"lastUpdateTime": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.Time"),
						},
					},
					"reason": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"message": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"type": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
			},
		},
		Dependencies: []string{
			"k8s.io/apimachinery/pkg/apis/meta/v1.Time"},
	}
}
