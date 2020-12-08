/*
Copyright AppsCode Inc. and Contributors

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

package util

import (
	"fmt"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"gomodules.xyz/x/log"
	appsv1 "k8s.io/api/apps/v1"
)

func MySQLTLSArgs(db *api.MySQL) []string {
	tlsArgs := []string{
		"--ssl-capath=/etc/mysql/certs",
		"--ssl-ca=/etc/mysql/certs/ca.crt",
		"--ssl-cert=/etc/mysql/certs/server.crt",
		"--ssl-key=/etc/mysql/certs/server.key",
	}
	return func() []string {
		if db.Spec.RequireSSL {
			return append(tlsArgs, MySQLRequireSSLArg())
		}
		return tlsArgs
	}()
}

func MySQLRequireSSLArg() string {
	return "--require-secure-transport=ON"
}

func MySQLExporterTLSArg() string {
	return "--config.my-cnf=/etc/mysql/certs/exporter.cnf"
}

func CheckMySQLTLSConfigRemoved(db *api.MySQL, sts *appsv1.StatefulSet) bool {
	for _, container := range sts.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL {
			for _, arg := range container.Args {
				for _, tlsArg := range MySQLTLSArgs(db) {
					if strings.Contains(arg, tlsArg) {
						log.Info(fmt.Sprintf("TLS Args is not removed yet for statefulSet: %s/%s", sts.Namespace, sts.Name))
						return false
					}
				}
			}

			for _, volumeMount := range container.VolumeMounts {
				if volumeMount.Name == api.DBTLSVolume {
					log.Info(fmt.Sprintf("TLS volumeMount is not removed yet for statefulSet: %s/%s", sts.Namespace, sts.Name))
					return false
				}
			}

			for _, volume := range sts.Spec.Template.Spec.Volumes {
				if volume.Name == api.DBTLSVolume {
					log.Info(fmt.Sprintf("TLS volume is not removed yet for statefulSet: %s/%s", sts.Namespace, sts.Name))
					return false
				}
			}
		}

		if container.Name == api.ContainerExporterName {
			for _, arg := range container.Args {
				if strings.Contains(arg, MySQLExporterTLSArg()) {
					log.Info(fmt.Sprintf("TLS Args is not removed yet for statefulSet %s/%s", sts.Namespace, sts.Name))
					return false
				}
			}

			for _, volumeMount := range container.VolumeMounts {
				if volumeMount.Name == api.DBExporterTLSVolume {
					log.Info(fmt.Sprintf("TLS volumeMount is not removed yet for statefulSet: %s/%s", sts.Namespace, sts.Name))
					return false
				}
			}

			for _, volume := range sts.Spec.Template.Spec.Volumes {
				if volume.Name == api.DBExporterTLSVolume {
					log.Info(fmt.Sprintf("TLS volume is not removed yet for statefulSet: %s/%s", sts.Namespace, sts.Name))
					return false
				}
			}
		}
	}
	return true
}

func CheckMySQLTLSConfigAdded(db *api.MySQL, sts *appsv1.StatefulSet) bool {
	for _, container := range sts.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularMySQL {
			for _, tlsArg := range MySQLTLSArgs(db) {
				var tlsArgExist bool
				for _, containerArg := range container.Args {
					if strings.Contains(containerArg, tlsArg) {
						tlsArgExist = true
					}
				}
				if !tlsArgExist {
					log.Info(fmt.Sprintf("TLS Args is not set yet for statefulSet: %s/%s", sts.Namespace, sts.Name))
					return false
				}
			}

			var volumeMountExist bool
			for _, volumeMount := range container.VolumeMounts {
				if volumeMount.Name == api.DBTLSVolume {
					volumeMountExist = true
					break
				}
			}
			if !volumeMountExist {
				log.Info(fmt.Sprintf("TLS volumeMount is not set yet for statefulSet: %s/%s", sts.Namespace, sts.Name))
				return false
			}

			var volumeExist bool
			for _, volume := range sts.Spec.Template.Spec.Volumes {
				if volume.Name == api.DBTLSVolume {
					volumeExist = true
					break
				}
			}
			if !volumeExist {
				log.Info(fmt.Sprintf("TLS volume is not set yet for statefulSet: %s/%s", sts.Namespace, sts.Name))
				return false
			}
		}

		if container.Name == api.ContainerExporterName {
			var exporterTLSArgExist bool
			for _, containerArg := range container.Args {
				if strings.Contains(containerArg, MySQLExporterTLSArg()) {
					exporterTLSArgExist = true
				}
			}
			if !exporterTLSArgExist {
				log.Info(fmt.Sprintf("TLS Args is not set yet for statefulSet: %s/%s", sts.Namespace, sts.Name))
				return false
			}

			var volumeMountExist bool
			for _, volumeMount := range container.VolumeMounts {
				if volumeMount.Name == api.DBExporterTLSVolume {
					volumeMountExist = true
					break
				}
			}
			if !volumeMountExist {
				log.Info(fmt.Sprintf("TLS volumeMount is not set yet for statefulSet: %s/%s", sts.Namespace, sts.Name))
				return false
			}

			var volumeExit bool
			for _, volume := range sts.Spec.Template.Spec.Volumes {
				if volume.Name == api.DBExporterTLSVolume {
					volumeExit = true
					break
				}
			}
			if !volumeExit {
				log.Info(fmt.Sprintf("TLS volume is not set yet for statefulSet: %s/%s", sts.Namespace, sts.Name))
				return false
			}
		}
	}
	return true
}
