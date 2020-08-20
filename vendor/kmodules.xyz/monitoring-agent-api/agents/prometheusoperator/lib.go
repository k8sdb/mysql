/*
Copyright The Kmodules Authors.

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

package prometheusoperator

import (
	"context"

	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	api "kmodules.xyz/monitoring-agent-api/api/v1"
	prom_util "kmodules.xyz/monitoring-agent-api/prometheus/v1"

	promapi "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	prom "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
)

// PrometheusCoreosOperator creates `ServiceMonitor` so that CoreOS Prometheus operator can generate necessary config for Prometheus.
type PrometheusCoreosOperator struct {
	at         api.AgentType
	k8sClient  kubernetes.Interface
	promClient prom.MonitoringV1Interface
}

func New(at api.AgentType, k8sClient kubernetes.Interface, promClient prom.MonitoringV1Interface) api.Agent {
	return &PrometheusCoreosOperator{
		at:         at,
		k8sClient:  k8sClient,
		promClient: promClient,
	}
}

func (agent *PrometheusCoreosOperator) GetType() api.AgentType {
	return agent.at
}

func (agent *PrometheusCoreosOperator) CreateOrUpdate(sp api.StatsAccessor, new *api.AgentSpec) (kutil.VerbType, error) {
	if !agent.isOperatorInstalled() {
		return kutil.VerbUnchanged, errors.New("cluster does not support CoreOS Prometheus operator")
	}

	svc, err := agent.k8sClient.CoreV1().Services(sp.GetNamespace()).Get(context.TODO(), sp.ServiceName(), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	var portName string
	for _, p := range svc.Spec.Ports {
		if p.Port == new.Prometheus.Exporter.Port {
			portName = p.Name
		}
	}
	if portName == "" {
		return kutil.VerbUnchanged, errors.New("no port found in stats service")
	}

	smMeta := metav1.ObjectMeta{
		Name:      sp.ServiceMonitorName(),
		Namespace: sp.GetNamespace(),
	}
	owner := metav1.NewControllerRef(svc, corev1.SchemeGroupVersion.WithKind("Service"))

	_, vt, err := prom_util.CreateOrPatchServiceMonitor(context.TODO(), agent.promClient, smMeta, func(in *promapi.ServiceMonitor) *promapi.ServiceMonitor {
		in.Labels = new.Prometheus.ServiceMonitor.Labels
		core_util.UpsertMap(in.Labels, svc.Labels)
		in.Spec.NamespaceSelector = promapi.NamespaceSelector{
			MatchNames: []string{sp.GetNamespace()},
		}
		in.Spec.Endpoints = []promapi.Endpoint{
			{
				Port:        portName,
				Interval:    new.Prometheus.ServiceMonitor.Interval,
				Path:        sp.Path(),
				HonorLabels: true,
			},
		}
		in.Spec.Selector = metav1.LabelSelector{
			MatchLabels: svc.Labels,
		}
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		return in
	}, metav1.PatchOptions{})

	return vt, err
}

func (agent *PrometheusCoreosOperator) Delete(sp api.StatsAccessor) (kutil.VerbType, error) {
	if !agent.isOperatorInstalled() {
		return kutil.VerbUnchanged, errors.New("cluster does not support CoreOS Prometheus operator")
	}

	err := agent.promClient.ServiceMonitors(sp.GetNamespace()).Delete(context.TODO(), sp.ServiceMonitorName(), metav1.DeleteOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return kutil.VerbDeleted, nil
}

func (agent *PrometheusCoreosOperator) isOperatorInstalled() bool {
	if resourceList, err := agent.k8sClient.Discovery().ServerPreferredResources(); discovery.IsGroupDiscoveryFailedError(err) || err == nil {
		for _, resources := range resourceList {
			gv, err := schema.ParseGroupVersion(resources.GroupVersion)
			if err != nil {
				return false
			}
			if gv.Group != promapi.SchemeGroupVersion.Group {
				continue
			}
			for _, resource := range resources.APIResources {
				if resource.Kind == promapi.PrometheusesKind || resource.Kind == promapi.ServiceMonitorsKind {
					return true
				}
			}
		}
	}
	return false
}
