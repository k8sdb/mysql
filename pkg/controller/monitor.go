package controller

import (
	"fmt"

	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/pkg/monitor"
)

func (c *Controller) newMonitorController(mysql *api.MySQL) (monitor.Monitor, error) {
	monitorSpec := mysql.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found in %v", mysql.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return monitor.NewPrometheusController(c.Client, c.ApiExtKubeClient, c.promClient, c.opt.OperatorNamespace), nil
	}

	return nil, fmt.Errorf("Monitoring controller not found for %v", monitorSpec)
}

func (c *Controller) addMonitor(mysql *api.MySQL) error {
	ctrl, err := c.newMonitorController(mysql)
	if err != nil {
		return err
	}
	return ctrl.AddMonitor(mysql.ObjectMeta, mysql.Spec.Monitor)
}

func (c *Controller) deleteMonitor(mysql *api.MySQL) error {
	ctrl, err := c.newMonitorController(mysql)
	if err != nil {
		return err
	}
	return ctrl.DeleteMonitor(mysql.ObjectMeta, mysql.Spec.Monitor)
}

func (c *Controller) updateMonitor(oldMySQL, updatedMySQL *api.MySQL) error {
	var err error
	var ctrl monitor.Monitor
	if updatedMySQL.Spec.Monitor == nil {
		ctrl, err = c.newMonitorController(oldMySQL)
	} else {
		ctrl, err = c.newMonitorController(updatedMySQL)
	}
	if err != nil {
		return err
	}
	return ctrl.UpdateMonitor(updatedMySQL.ObjectMeta, oldMySQL.Spec.Monitor, updatedMySQL.Spec.Monitor)
}
