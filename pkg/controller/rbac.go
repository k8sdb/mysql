package controller

import (
	kutilcore "github.com/appscode/kutil/core/v1"
	kutilrbac "github.com/appscode/kutil/rbac/v1beta1"
	"github.com/k8sdb/apimachinery/apis/kubedb"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	rbac "k8s.io/client-go/pkg/apis/rbac/v1beta1"
)

func (c *Controller) deleteRole(mysql *tapi.MySQL) error {
	// Delete existing Roles
	if err := c.Client.RbacV1beta1().Roles(mysql.Namespace).Delete(mysql.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRole(mysql *tapi.MySQL) error {
	// Create new Roles
	_, err := kutilrbac.EnsureRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      mysql.OffshootName(),
			Namespace: mysql.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{kubedb.GroupName},
					Resources:     []string{tapi.ResourceTypeMySQL},
					ResourceNames: []string{mysql.Name},
					Verbs:         []string{"get"},
				},
				{
					// TODO. Use this if secret is necessary, Otherwise remove it
					APIGroups:     []string{apiv1.GroupName},
					Resources:     []string{"secrets"},
					ResourceNames: []string{mysql.Spec.DatabaseSecret.SecretName},
					Verbs:         []string{"get"},
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) deleteServiceAccount(mysql *tapi.MySQL) error {
	// Delete existing ServiceAccount
	if err := c.Client.CoreV1().ServiceAccounts(mysql.Namespace).Delete(mysql.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createServiceAccount(mysql *tapi.MySQL) error {
	// Create new ServiceAccount
	_, err := kutilcore.EnsureServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      mysql.OffshootName(),
			Namespace: mysql.Namespace,
		},
		func(in *apiv1.ServiceAccount) *apiv1.ServiceAccount {
			return in
		},
	)
	return err
}

func (c *Controller) deleteRoleBinding(mysql *tapi.MySQL) error {
	// Delete existing RoleBindings
	if err := c.Client.RbacV1beta1().RoleBindings(mysql.Namespace).Delete(mysql.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRoleBinding(mysql *tapi.MySQL) error {
	// Ensure new RoleBindings
	_, err := kutilrbac.EnsureRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      mysql.OffshootName(),
			Namespace: mysql.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     mysql.OffshootName(),
			}
			in.Subjects = []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      mysql.OffshootName(),
					Namespace: mysql.Namespace,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) createRBACStuff(mysql *tapi.MySQL) error {
	// Delete Existing Role
	if err := c.deleteRole(mysql); err != nil {
		return err
	}
	// Create New Role
	if err := c.createRole(mysql); err != nil {
		return err
	}

	// Create New ServiceAccount
	if err := c.createServiceAccount(mysql); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	// Create New RoleBinding
	if err := c.createRoleBinding(mysql); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func (c *Controller) deleteRBACStuff(mysql *tapi.MySQL) error {
	// Delete Existing Role
	if err := c.deleteRole(mysql); err != nil {
		return err
	}

	// Delete ServiceAccount
	if err := c.deleteServiceAccount(mysql); err != nil {
		return err
	}

	// Delete New RoleBinding
	if err := c.deleteRoleBinding(mysql); err != nil {
		return err
	}

	return nil
}
