package monitor

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	prom "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PrometheusController struct {
	kubeClient        kubernetes.Interface
	apiExtKubeClient  crd_cs.ApiextensionsV1beta1Interface
	promClient        prom.MonitoringV1Interface
	operatorNamespace string
}

func NewPrometheusController(kubeClient kubernetes.Interface, apiExtKubeClient crd_cs.ApiextensionsV1beta1Interface, promClient prom.MonitoringV1Interface, operatorNamespace string) Monitor {
	return &PrometheusController{
		kubeClient:        kubeClient,
		apiExtKubeClient:  apiExtKubeClient,
		promClient:        promClient,
		operatorNamespace: operatorNamespace,
	}
}

func (c *PrometheusController) AddMonitor(meta metav1.ObjectMeta, spec *api.MonitorSpec) error {
	if !c.SupportsCoreOSOperator() {
		return errors.New("cluster does not support CoreOS Prometheus operator")
	}
	return c.ensureServiceMonitor(meta, spec, spec)
}

func (c *PrometheusController) UpdateMonitor(meta metav1.ObjectMeta, old, new *api.MonitorSpec) error {
	if !c.SupportsCoreOSOperator() {
		return errors.New("cluster does not support CoreOS Prometheus operator")
	}
	return c.ensureServiceMonitor(meta, old, new)
}

func (c *PrometheusController) DeleteMonitor(meta metav1.ObjectMeta, spec *api.MonitorSpec) error {
	if !c.SupportsCoreOSOperator() {
		return errors.New("cluster does not support CoreOS Prometheus operator")
	}
	if err := c.promClient.ServiceMonitors(spec.Prometheus.Namespace).Delete(getServiceMonitorName(meta), nil); !kerr.IsNotFound(err) {
		return err
	}
	return nil
}

func (c *PrometheusController) SupportsCoreOSOperator() bool {
	_, err := c.apiExtKubeClient.CustomResourceDefinitions().Get(prom.PrometheusName+"."+prom.Group, metav1.GetOptions{})
	if err != nil {
		return false
	}
	_, err = c.apiExtKubeClient.CustomResourceDefinitions().Get(prom.ServiceMonitorName+"."+prom.Group, metav1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}

func (c *PrometheusController) ensureServiceMonitor(meta metav1.ObjectMeta, old, new *api.MonitorSpec) error {
	name := getServiceMonitorName(meta)
	if old != nil && (new == nil || old.Prometheus.Namespace != new.Prometheus.Namespace) {
		err := c.promClient.ServiceMonitors(old.Prometheus.Namespace).Delete(name, nil)
		if err != nil && !kerr.IsNotFound(err) {
			return err
		}
		if new == nil {
			return nil
		}
	}

	actual, err := c.promClient.ServiceMonitors(new.Prometheus.Namespace).Get(name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return c.createServiceMonitor(meta, new)
	} else if err != nil {
		return err
	}

	update := false
	if !reflect.DeepEqual(actual.Labels, new.Prometheus.Labels) {
		update = true
	}

	if !update {
		for _, e := range actual.Spec.Endpoints {
			if e.Interval != new.Prometheus.Interval {
				update = true
				break
			}
		}
	}

	if update {
		actual.Labels = new.Prometheus.Labels
		for i := range actual.Spec.Endpoints {
			actual.Spec.Endpoints[i].Interval = new.Prometheus.Interval
		}
		_, err := c.promClient.ServiceMonitors(new.Prometheus.Namespace).Update(actual)
		return err
	}

	return nil
}

func (c *PrometheusController) createServiceMonitor(meta metav1.ObjectMeta, spec *api.MonitorSpec) error {
	svc, err := c.kubeClient.CoreV1().Services(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	ports := svc.Spec.Ports
	if len(ports) == 0 {
		return errors.New("no port found in database service")
	}

	sm := &prom.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getServiceMonitorName(meta),
			Namespace: spec.Prometheus.Namespace,
			Labels:    spec.Prometheus.Labels,
		},
		Spec: prom.ServiceMonitorSpec{
			NamespaceSelector: prom.NamespaceSelector{
				MatchNames: []string{svc.Namespace},
			},
			Endpoints: []prom.Endpoint{
				{
					Port:     api.PrometheusExporterPortName,
					Interval: spec.Prometheus.Interval,
					Path:     fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/metrics", meta.Namespace, getTypeFromSelfLink(meta.SelfLink), meta.Name),
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: svc.Spec.Selector,
			},
		},
	}
	if _, err := c.promClient.ServiceMonitors(spec.Prometheus.Namespace).Create(sm); !kerr.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func getTypeFromSelfLink(selfLink string) string {
	if len(selfLink) == 0 {
		return ""
	}
	s := strings.Split(selfLink, "/")
	return s[len(s)-2]
}

func getServiceMonitorName(meta metav1.ObjectMeta) string {
	return fmt.Sprintf("kubedb-%s-%s", meta.Namespace, meta.Name)
}
