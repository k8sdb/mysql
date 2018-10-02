// +build !ignore_autogenerated

/*
Copyright 2018 The Kmodules Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by openapi-gen. DO NOT EDIT.

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"kmodules.xyz/offshoot-api/api/v1.ObjectMeta":          schema_kmodulesxyz_offshoot_api_api_v1_ObjectMeta(ref),
		"kmodules.xyz/offshoot-api/api/v1.PodSpec":             schema_kmodulesxyz_offshoot_api_api_v1_PodSpec(ref),
		"kmodules.xyz/offshoot-api/api/v1.PodTemplateSpec":     schema_kmodulesxyz_offshoot_api_api_v1_PodTemplateSpec(ref),
		"kmodules.xyz/offshoot-api/api/v1.ServiceSpec":         schema_kmodulesxyz_offshoot_api_api_v1_ServiceSpec(ref),
		"kmodules.xyz/offshoot-api/api/v1.ServiceTemplateSpec": schema_kmodulesxyz_offshoot_api_api_v1_ServiceTemplateSpec(ref),
	}
}

func schema_kmodulesxyz_offshoot_api_api_v1_ObjectMeta(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create.",
				Properties: map[string]spec.Schema{
					"annotations": {
						SchemaProps: spec.SchemaProps{
							Description: "Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations",
							Type:        []string{"object"},
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
				},
			},
		},
		Dependencies: []string{},
	}
}

func schema_kmodulesxyz_offshoot_api_api_v1_PodSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Properties: map[string]spec.Schema{
					"args": {
						SchemaProps: spec.SchemaProps{
							Description: "Arguments to the entrypoint. The docker image's CMD is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container's environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell",
							Type:        []string{"array"},
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
					"nodeSelector": {
						SchemaProps: spec.SchemaProps{
							Description: "NodeSelector is a selector which must be true for the pod to fit on a node. Selector which must match a node's labels for the pod to be scheduled on that node. More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/",
							Type:        []string{"object"},
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
					"resources": {
						SchemaProps: spec.SchemaProps{
							Description: "Compute Resources required by the sidecar container.",
							Ref:         ref("k8s.io/api/core/v1.ResourceRequirements"),
						},
					},
					"affinity": {
						SchemaProps: spec.SchemaProps{
							Description: "If specified, the pod's scheduling constraints",
							Ref:         ref("k8s.io/api/core/v1.Affinity"),
						},
					},
					"schedulerName": {
						SchemaProps: spec.SchemaProps{
							Description: "If specified, the pod will be dispatched by specified scheduler. If not specified, the pod will be dispatched by default scheduler.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"tolerations": {
						SchemaProps: spec.SchemaProps{
							Description: "If specified, the pod's tolerations.",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.Toleration"),
									},
								},
							},
						},
					},
					"imagePullSecrets": {
						SchemaProps: spec.SchemaProps{
							Description: "ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec. If specified, these secrets will be passed to individual puller implementations for them to use.",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.LocalObjectReference"),
									},
								},
							},
						},
					},
					"env": {
						SchemaProps: spec.SchemaProps{
							Description: "List of environment variables to set in the container. Cannot be updated.",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.EnvVar"),
									},
								},
							},
						},
					},
					"initContainers": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-patch-merge-key": "name",
								"x-kubernetes-patch-strategy":  "merge",
							},
						},
						SchemaProps: spec.SchemaProps{
							Description: "List of initialization containers belonging to the pod. Init containers are executed in order prior to containers being started. If any init container fails, the pod is considered to have failed and is handled according to its restartPolicy. The name for an init container or normal container must be unique among all containers. Init containers may not have Lifecycle actions, Readiness probes, or Liveness probes. The resourceRequirements of an init container are taken into account during scheduling by finding the highest request/limit for each resource type, and then using the max of of that value or the sum of the normal containers. Limits are applied to init containers in a similar fashion. Init containers cannot currently be added or removed. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.Container"),
									},
								},
							},
						},
					},
					"priorityClassName": {
						SchemaProps: spec.SchemaProps{
							Description: "If specified, indicates the pod's priority. \"system-node-critical\" and \"system-cluster-critical\" are two special keywords which indicate the highest priorities with the former being the highest priority. Any other name must be defined by creating a PriorityClass object with that name. If not specified, the pod priority will be default or zero if there is no default.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"priority": {
						SchemaProps: spec.SchemaProps{
							Description: "The priority value. Various system components use this field to find the priority of the pod. When Priority Admission Controller is enabled, it prevents users from setting this field. The admission controller populates this field from PriorityClassName. The higher the value, the higher the priority.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"securityContext": {
						SchemaProps: spec.SchemaProps{
							Description: "SecurityContext holds pod-level security attributes and common container settings. Optional: Defaults to empty.  See type description for default values of each field.",
							Ref:         ref("k8s.io/api/core/v1.PodSecurityContext"),
						},
					},
					"livenessProbe": {
						SchemaProps: spec.SchemaProps{
							Description: "Periodic probe of container liveness. Container will be restarted if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes",
							Ref:         ref("k8s.io/api/core/v1.Probe"),
						},
					},
					"readinessProbe": {
						SchemaProps: spec.SchemaProps{
							Description: "Periodic probe of container service readiness. Container will be removed from service endpoints if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes",
							Ref:         ref("k8s.io/api/core/v1.Probe"),
						},
					},
					"lifecycle": {
						SchemaProps: spec.SchemaProps{
							Description: "Actions that the management system should take in response to container lifecycle events. Cannot be updated.",
							Ref:         ref("k8s.io/api/core/v1.Lifecycle"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"k8s.io/api/core/v1.Affinity", "k8s.io/api/core/v1.Container", "k8s.io/api/core/v1.EnvVar", "k8s.io/api/core/v1.Lifecycle", "k8s.io/api/core/v1.LocalObjectReference", "k8s.io/api/core/v1.PodSecurityContext", "k8s.io/api/core/v1.Probe", "k8s.io/api/core/v1.ResourceRequirements", "k8s.io/api/core/v1.Toleration"},
	}
}

func schema_kmodulesxyz_offshoot_api_api_v1_PodTemplateSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "PodTemplateSpec describes the data a pod should have when created from a template",
				Properties: map[string]spec.Schema{
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Description: "Standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata",
							Ref:         ref("kmodules.xyz/offshoot-api/api/v1.ObjectMeta"),
						},
					},
					"controller": {
						SchemaProps: spec.SchemaProps{
							Description: "Workload controller's metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata",
							Ref:         ref("kmodules.xyz/offshoot-api/api/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Description: "Specification of the desired behavior of the pod. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status",
							Ref:         ref("kmodules.xyz/offshoot-api/api/v1.PodSpec"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"kmodules.xyz/offshoot-api/api/v1.ObjectMeta", "kmodules.xyz/offshoot-api/api/v1.PodSpec"},
	}
}

func schema_kmodulesxyz_offshoot_api_api_v1_ServiceSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceSpec describes the attributes that a user creates on a service.",
				Properties: map[string]spec.Schema{
					"clusterIP": {
						SchemaProps: spec.SchemaProps{
							Description: "clusterIP is the IP address of the service and is usually assigned randomly by the master. If an address is specified manually and is not in use by others, it will be allocated to the service; otherwise, creation of the service will fail. This field can not be changed through updates. Valid values are \"None\", empty string (\"\"), or a valid IP address. \"None\" can be specified for headless services when proxying is not required. Only applies to types ClusterIP, NodePort, and LoadBalancer. Ignored if type is ExternalName. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"type": {
						SchemaProps: spec.SchemaProps{
							Description: "type determines how the Service is exposed. Defaults to ClusterIP. Valid options are ExternalName, ClusterIP, NodePort, and LoadBalancer. \"ExternalName\" maps to the specified externalName. \"ClusterIP\" allocates a cluster-internal IP address for load-balancing to endpoints. Endpoints are determined by the selector or if that is not specified, by manual construction of an Endpoints object. If clusterIP is \"None\", no virtual IP is allocated and the endpoints are published as a set of endpoints rather than a stable IP. \"NodePort\" builds on ClusterIP and allocates a port on every node which routes to the clusterIP. \"LoadBalancer\" builds on NodePort and creates an external load-balancer (if supported in the current cloud) which routes to the clusterIP. More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services---service-types",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"externalIPs": {
						SchemaProps: spec.SchemaProps{
							Description: "externalIPs is a list of IP addresses for which nodes in the cluster will also accept traffic for this service.  These IPs are not managed by Kubernetes.  The user is responsible for ensuring that traffic arrives at a node with this IP.  A common example is external load-balancers that are not part of the Kubernetes system.",
							Type:        []string{"array"},
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
					"loadBalancerIP": {
						SchemaProps: spec.SchemaProps{
							Description: "Only applies to Service Type: LoadBalancer LoadBalancer will get created with the IP specified in this field. This feature depends on whether the underlying cloud-provider supports specifying the loadBalancerIP when a load balancer is created. This field will be ignored if the cloud-provider does not support the feature.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"loadBalancerSourceRanges": {
						SchemaProps: spec.SchemaProps{
							Description: "If specified and supported by the platform, this will restrict traffic through the cloud-provider load-balancer will be restricted to the specified client IPs. This field will be ignored if the cloud-provider does not support the feature.\" More info: https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/",
							Type:        []string{"array"},
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
					"externalTrafficPolicy": {
						SchemaProps: spec.SchemaProps{
							Description: "externalTrafficPolicy denotes if this Service desires to route external traffic to node-local or cluster-wide endpoints. \"Local\" preserves the client source IP and avoids a second hop for LoadBalancer and Nodeport type services, but risks potentially imbalanced traffic spreading. \"Cluster\" obscures the client source IP and may cause a second hop to another node, but should have good overall load-spreading.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"healthCheckNodePort": {
						SchemaProps: spec.SchemaProps{
							Description: "healthCheckNodePort specifies the healthcheck nodePort for the service. If not specified, HealthCheckNodePort is created by the service api backend with the allocated nodePort. Will use user-specified nodePort value if specified by the client. Only effects when Type is set to LoadBalancer and ExternalTrafficPolicy is set to Local.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
				},
			},
		},
		Dependencies: []string{},
	}
}

func schema_kmodulesxyz_offshoot_api_api_v1_ServiceTemplateSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceTemplateSpec describes the data a service should have when created from a template",
				Properties: map[string]spec.Schema{
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Description: "Standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata",
							Ref:         ref("kmodules.xyz/offshoot-api/api/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Description: "Specification of the desired behavior of the service. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status",
							Ref:         ref("kmodules.xyz/offshoot-api/api/v1.ServiceSpec"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"kmodules.xyz/offshoot-api/api/v1.ObjectMeta", "kmodules.xyz/offshoot-api/api/v1.ServiceSpec"},
	}
}
