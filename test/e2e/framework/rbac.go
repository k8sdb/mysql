/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"context"

	"gomodules.xyz/x/crypto/rand"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	podsecuritypolicies = "podsecuritypolicies"
	rbacApiGroup        = "rbac.authorization.k8s.io"
	GET                 = "get"
	LIST                = "list"
	PATCH               = "patch"
	CREATE              = "create"
	UPDATE              = "update"
	USE                 = "use"
	POLICY              = "policy"
	Role                = "Role"
	ServiceAccount      = "ServiceAccount"
	CustomSecretSuffix  = "custom-secret"
	mysqlUser           = "root"
)

func (i *Invocation) ServiceAccount() *core.ServiceAccount {
	return &core.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix(i.app + "-mysql"),
			Namespace: i.namespace,
		},
	}
}

func (i *Invocation) RoleForMySQL(meta metav1.ObjectMeta) *rbac.Role {
	return &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix(i.app + "-mysql"),
			Namespace: i.namespace,
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{
					POLICY,
				},
				ResourceNames: []string{
					meta.Name,
				},
				Resources: []string{
					podsecuritypolicies,
				},
				Verbs: []string{
					USE,
				},
			},
		},
	}
}
func (i *Invocation) RoleForSnapshot(meta metav1.ObjectMeta) *rbac.Role {
	return &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix(i.app + "-mysql"),
			Namespace: i.namespace,
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{
					POLICY,
				},
				ResourceNames: []string{
					meta.Name,
				},
				Resources: []string{
					podsecuritypolicies,
				},
				Verbs: []string{
					USE,
				},
			},
		},
	}
}

func (i *Invocation) RoleBinding(saName string, roleName string) *rbac.RoleBinding {
	return &rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix(i.app + "-mysql"),
			Namespace: i.namespace,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbacApiGroup,
			Kind:     Role,
			Name:     roleName,
		},
		Subjects: []rbac.Subject{
			{
				Kind:      ServiceAccount,
				Namespace: i.namespace,
				Name:      saName,
			},
		},
	}
}

func (f *Framework) CreateServiceAccount(obj *core.ServiceAccount) error {
	_, err := f.kubeClient.CoreV1().ServiceAccounts(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	return err
}

func (f *Framework) CreateRole(obj *rbac.Role) error {
	_, err := f.kubeClient.RbacV1().Roles(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	return err
}
func (f *Framework) CreateRoleBinding(obj *rbac.RoleBinding) error {
	_, err := f.kubeClient.RbacV1().RoleBindings(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	return err
}
