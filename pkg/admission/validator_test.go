/*
Copyright The KubeDB Authors.

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
package admission

import (
	"net/http"
	"testing"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	extFake "kubedb.dev/apimachinery/client/clientset/versioned/fake"
	"kubedb.dev/apimachinery/client/clientset/versioned/scheme"

	"github.com/appscode/go/types"
	admission "k8s.io/api/admission/v1beta1"
	apps "k8s.io/api/apps/v1"
	authenticationV1 "k8s.io/api/authentication/v1"
	core "k8s.io/api/core/v1"
	storageV1beta1 "k8s.io/api/storage/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clientSetScheme "k8s.io/client-go/kubernetes/scheme"
	"kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func init() {
	utilruntime.Must(scheme.AddToScheme(clientSetScheme.Scheme))
}

var requestKind = metaV1.GroupVersionKind{
	Group:   api.SchemeGroupVersion.Group,
	Version: api.SchemeGroupVersion.Version,
	Kind:    api.ResourceKindMySQL,
}

func TestMySQLValidator_Admit(t *testing.T) {
	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			validator := MySQLValidator{}

			validator.initialized = true
			validator.extClient = extFake.NewSimpleClientset(
				&catalog.MySQLVersion{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "8.0",
					},
					Spec: catalog.MySQLVersionSpec{
						Version: "8.0.0",
					},
				},
				&catalog.MySQLVersion{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "5.6",
					},
					Spec: catalog.MySQLVersionSpec{
						Version: "5.6",
					},
				},
				&catalog.MySQLVersion{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "5.7.25",
					},
					Spec: catalog.MySQLVersionSpec{
						Version: "5.7.25",
					},
				},
			)
			validator.client = fake.NewSimpleClientset(
				&core.Secret{
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "foo-auth",
						Namespace: "default",
					},
				},
				&storageV1beta1.StorageClass{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "standard",
					},
				},
			)

			objJS, err := meta.MarshalToJson(&c.object, api.SchemeGroupVersion)
			if err != nil {
				panic(err)
			}
			oldObjJS, err := meta.MarshalToJson(&c.oldObject, api.SchemeGroupVersion)
			if err != nil {
				panic(err)
			}

			req := new(admission.AdmissionRequest)

			req.Kind = c.kind
			req.Name = c.objectName
			req.Namespace = c.namespace
			req.Operation = c.operation
			req.UserInfo = authenticationV1.UserInfo{}
			req.Object.Raw = objJS
			req.OldObject.Raw = oldObjJS

			if c.heatUp {
				if _, err := validator.extClient.KubedbV1alpha1().MySQLs(c.namespace).Create(&c.object); err != nil && !kerr.IsAlreadyExists(err) {
					t.Errorf(err.Error())
				}
			}
			if c.operation == admission.Delete {
				req.Object = runtime.RawExtension{}
			}
			if c.operation != admission.Update {
				req.OldObject = runtime.RawExtension{}
			}

			response := validator.Admit(req)
			if c.result == true {
				if response.Allowed != true {
					t.Errorf("expected: 'Allowed=true'. but got response: %v", response)
				}
			} else if c.result == false {
				if response.Allowed == true || response.Result.Code == http.StatusInternalServerError {
					t.Errorf("expected: 'Allowed=false', but got response: %v", response)
				}
			}
		})
	}

}

var cases = []struct {
	testName   string
	kind       metaV1.GroupVersionKind
	objectName string
	namespace  string
	operation  admission.Operation
	object     api.MySQL
	oldObject  api.MySQL
	heatUp     bool
	result     bool
}{
	{"Create Valid MySQL",
		requestKind,
		"foo",
		"default",
		admission.Create,
		sampleMySQL(),
		api.MySQL{},
		false,
		true,
	},
	{"Create Invalid MySQL",
		requestKind,
		"foo",
		"default",
		admission.Create,
		getAwkwardMySQL(),
		api.MySQL{},
		false,
		false,
	},
	{"Edit MySQL Spec.DatabaseSecret with Existing Secret",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editExistingSecret(sampleMySQL()),
		sampleMySQL(),
		false,
		true,
	},
	{"Edit MySQL Spec.DatabaseSecret with non Existing Secret",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editNonExistingSecret(sampleMySQL()),
		sampleMySQL(),
		false,
		true,
	},
	{"Edit Status",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editStatus(sampleMySQL()),
		sampleMySQL(),
		false,
		true,
	},
	{"Edit Spec.Monitor",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editSpecMonitor(sampleMySQL()),
		sampleMySQL(),
		false,
		true,
	},
	{"Edit Invalid Spec.Monitor",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editSpecInvalidMonitor(sampleMySQL()),
		sampleMySQL(),
		false,
		false,
	},
	{"Edit Spec.TerminationPolicy",
		requestKind,
		"foo",
		"default",
		admission.Update,
		haltDatabase(sampleMySQL()),
		sampleMySQL(),
		false,
		true,
	},
	{"Delete MySQL when Spec.TerminationPolicy=DoNotTerminate",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		sampleMySQL(),
		api.MySQL{},
		true,
		false,
	},
	{"Delete MySQL when Spec.TerminationPolicy=Halt",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		haltDatabase(sampleMySQL()),
		api.MySQL{},
		true,
		true,
	},
	{"Delete Non Existing MySQL",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		api.MySQL{},
		api.MySQL{},
		false,
		true,
	},

	// For MySQL Group Replication
	{"Create valid group",
		requestKind,
		"foo",
		"default",
		admission.Create,
		validGroup(sampleMySQL()),
		api.MySQL{},
		false,
		true,
	},
	{"Create group with '.spec.topology.mode' not set",
		requestKind,
		"foo",
		"default",
		admission.Create,
		groupWithClusterModeNotSet(),
		api.MySQL{},
		false,
		false,
	},
	{"Create group with invalid '.spec.topology.mode'",
		requestKind,
		"foo",
		"default",
		admission.Create,
		groupWithInvalidClusterMode(),
		api.MySQL{},
		false,
		false,
	},
	{"Create group with single replica",
		requestKind,
		"foo",
		"default",
		admission.Create,
		groupWithSingleReplica(),
		api.MySQL{},
		false,
		false,
	},
	{"Create group with replicas more than max group size",
		requestKind,
		"foo",
		"default",
		admission.Create,
		groupWithOverReplicas(),
		api.MySQL{},
		false,
		false,
	},
	{"Create group with unsupported MySQL server version",
		requestKind,
		"foo",
		"default",
		admission.Create,
		groupWithUnsupportedServerVersion(),
		api.MySQL{},
		false,
		true,
	},
	{"Create group with Non-tri formatted MySQL server version",
		requestKind,
		"foo",
		"default",
		admission.Create,
		groupWithNonTriFormattedServerVersion(),
		api.MySQL{},
		false,
		true,
	},
	{"Create group with empty group name",
		requestKind,
		"foo",
		"default",
		admission.Create,
		groupWithEmptyGroupName(),
		api.MySQL{},
		false,
		false,
	},
	{"Create group with invalid group name",
		requestKind,
		"foo",
		"default",
		admission.Create,
		groupWithInvalidGroupName(),
		api.MySQL{},
		false,
		false,
	},
	{"Create group with baseServerID 0",
		requestKind,
		"foo",
		"default",
		admission.Create,
		groupWithBaseServerIDZero(),
		api.MySQL{},
		false,
		false,
	},
	{"Create group with baseServerID exceeded max limit",
		requestKind,
		"foo",
		"default",
		admission.Create,
		groupWithBaseServerIDExceededMaxLimit(),
		api.MySQL{},
		false,
		false,
	},
}

func sampleMySQL() api.MySQL {
	return api.MySQL{
		TypeMeta: metaV1.TypeMeta{
			Kind:       api.ResourceKindMySQL,
			APIVersion: api.SchemeGroupVersion.String(),
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindMySQL,
			},
		},
		Spec: api.MySQLSpec{
			Version:     "8.0",
			Replicas:    types.Int32P(1),
			StorageType: api.StorageTypeDurable,
			Storage: &core.PersistentVolumeClaimSpec{
				StorageClassName: types.StringP("standard"),
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceStorage: resource.MustParse("100Mi"),
					},
				},
			},
			Init: &api.InitSpec{
				ScriptSource: &api.ScriptSourceSpec{
					VolumeSource: core.VolumeSource{
						GitRepo: &core.GitRepoVolumeSource{
							Repository: "https://github.com/kubedb/mysql-init-scripts.git",
							Directory:  ".",
						},
					},
				},
			},
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			TerminationPolicy: api.TerminationPolicyDoNotTerminate,
		},
	}
}

func getAwkwardMySQL() api.MySQL {
	mysql := sampleMySQL()
	mysql.Spec.Version = "3.0"
	return mysql
}

func editExistingSecret(old api.MySQL) api.MySQL {
	old.Spec.DatabaseSecret = &core.SecretVolumeSource{
		SecretName: "foo-auth",
	}
	return old
}

func editNonExistingSecret(old api.MySQL) api.MySQL {
	old.Spec.DatabaseSecret = &core.SecretVolumeSource{
		SecretName: "foo-auth-fused",
	}
	return old
}

func editStatus(old api.MySQL) api.MySQL {
	old.Status = api.MySQLStatus{
		Phase: api.DatabasePhaseCreating,
	}
	return old
}

func editSpecMonitor(old api.MySQL) api.MySQL {
	old.Spec.Monitor = &mona.AgentSpec{
		Agent: mona.AgentPrometheusBuiltin,
		Prometheus: &mona.PrometheusSpec{
			Exporter: &mona.PrometheusExporterSpec{
				Port: 1289,
			},
		},
	}
	return old
}

// should be failed because more fields required for COreOS Monitoring
func editSpecInvalidMonitor(old api.MySQL) api.MySQL {
	old.Spec.Monitor = &mona.AgentSpec{
		Agent: mona.AgentPrometheusOperator,
	}
	return old
}

func haltDatabase(old api.MySQL) api.MySQL {
	old.Spec.TerminationPolicy = api.TerminationPolicyHalt
	return old
}

func validGroup(old api.MySQL) api.MySQL {
	old.Spec.Version = api.MySQLGRRecommendedVersion
	old.Spec.Replicas = types.Int32P(api.MySQLDefaultGroupSize)
	clusterMode := api.MySQLClusterModeGroup
	old.Spec.Topology = &api.MySQLClusterTopology{
		Mode: &clusterMode,
		Group: &api.MySQLGroupSpec{
			Name:         "dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b",
			BaseServerID: types.Int64P(api.MySQLDefaultBaseServerID),
		},
	}

	return old
}

func groupWithClusterModeNotSet() api.MySQL {
	old := validGroup(sampleMySQL())
	old.Spec.Topology.Mode = nil

	return old
}

func groupWithInvalidClusterMode() api.MySQL {
	old := validGroup(sampleMySQL())
	gr := api.MySQLClusterMode("groupReplication")
	old.Spec.Topology.Mode = &gr

	return old
}

func groupWithSingleReplica() api.MySQL {
	old := validGroup(sampleMySQL())
	old.Spec.Replicas = types.Int32P(1)

	return old
}

func groupWithOverReplicas() api.MySQL {
	old := validGroup(sampleMySQL())
	old.Spec.Replicas = types.Int32P(api.MySQLMaxGroupMembers + 1)

	return old
}

func groupWithUnsupportedServerVersion() api.MySQL {
	old := validGroup(sampleMySQL())
	old.Spec.Version = "8.0"

	return old
}

func groupWithNonTriFormattedServerVersion() api.MySQL {
	old := validGroup(sampleMySQL())
	old.Spec.Version = "5.6"

	return old
}

func groupWithEmptyGroupName() api.MySQL {
	old := validGroup(sampleMySQL())
	old.Spec.Topology.Group.Name = ""

	return old
}

func groupWithInvalidGroupName() api.MySQL {
	old := validGroup(sampleMySQL())
	old.Spec.Topology.Group.Name = "a-a-a-a-a"

	return old
}

func groupWithBaseServerIDZero() api.MySQL {
	old := validGroup(sampleMySQL())
	old.Spec.Topology.Group.BaseServerID = types.Int64P(0)

	return old
}

func groupWithBaseServerIDExceededMaxLimit() api.MySQL {
	old := validGroup(sampleMySQL())
	old.Spec.Topology.Group.BaseServerID = types.Int64P(api.MySQLMaxBaseServerID + 1)

	return old
}
