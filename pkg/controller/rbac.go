package controller

import (
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	core "k8s.io/api/core/v1"
	policy_v1beta1 "k8s.io/api/policy/v1beta1"
	rbac "k8s.io/api/rbac/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	core_util "kmodules.xyz/client-go/core/v1"
	policy_util "kmodules.xyz/client-go/policy/v1beta1"
	rbac_util "kmodules.xyz/client-go/rbac/v1beta1"
)

func (c *Controller) createServiceAccount(mysql *api.MySQL, saName string) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mysql)
	if rerr != nil {
		return rerr
	}
	// Create new ServiceAccount
	_, _, err := core_util.CreateOrPatchServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      saName,
			Namespace: mysql.Namespace,
		},
		func(in *core.ServiceAccount) *core.ServiceAccount {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			return in
		},
	)
	return err
}

func (c *Controller) ensureRole(mysql *api.MySQL, name string) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mysql)
	if rerr != nil {
		return rerr
	}

	// Create new Role for ElasticSearch and it's Snapshot
	_, _, err := rbac_util.CreateOrPatchRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      name,
			Namespace: mysql.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{policy_v1beta1.GroupName},
					Resources:     []string{"podsecuritypolicies"},
					Verbs:         []string{"use"},
					ResourceNames: []string{name},
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) createRoleBinding(mysql *api.MySQL, name string) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mysql)
	if rerr != nil {
		return rerr
	}
	// Ensure new RoleBindings for ElasticSearch and it's Snapshot
	_, _, err := rbac_util.CreateOrPatchRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      name,
			Namespace: mysql.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     name,
			}
			in.Subjects = []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      name,
					Namespace: mysql.Namespace,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) ensurePSP(mysql *api.MySQL, name string) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, mysql)
	if rerr != nil {
		return rerr
	}

	// Ensure Pod Security Policies for ElasticSearch and it's Snapshot
	noEscalation := false
	_, _, err := policy_util.CreateOrPatchPodSecurityPolicy(c.Client,
		metav1.ObjectMeta{
			Name: name,
		},
		func(in *policy_v1beta1.PodSecurityPolicy) *policy_v1beta1.PodSecurityPolicy {
			in.OwnerReferences = []metav1.OwnerReference{
				{
					APIVersion: ref.APIVersion,
					Kind:       ref.Kind,
					Name:       ref.Name,
					UID:        ref.UID,
				},
			}
			in.Spec = policy_v1beta1.PodSecurityPolicySpec{
				Privileged:               false,
				AllowPrivilegeEscalation: &noEscalation,
				Volumes: []policy_v1beta1.FSType{
					policy_v1beta1.All,
				},
				HostIPC:     false,
				HostNetwork: false,
				HostPID:     false,
				RunAsUser: policy_v1beta1.RunAsUserStrategyOptions{
					Rule: policy_v1beta1.RunAsUserStrategyRunAsAny,
				},
				SELinux: policy_v1beta1.SELinuxStrategyOptions{
					Rule: policy_v1beta1.SELinuxStrategyRunAsAny,
				},
				FSGroup: policy_v1beta1.FSGroupStrategyOptions{
					Rule: policy_v1beta1.FSGroupStrategyRunAsAny,
				},
				SupplementalGroups: policy_v1beta1.SupplementalGroupsStrategyOptions{
					Rule: policy_v1beta1.SupplementalGroupsStrategyRunAsAny,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) ensureRBACStuff(mysql *api.MySQL) error {
	// Create New ServiceAccount
	if err := c.createServiceAccount(mysql, mysql.OffshootName()); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	if err := c.ensurePSP(mysql, mysql.OffshootName()); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	// Create New Role
	if err := c.ensureRole(mysql, mysql.OffshootName()); err != nil {
		return err
	}

	// Create New RoleBinding
	if err := c.createRoleBinding(mysql, mysql.OffshootName()); err != nil {
		return err
	}

	// Create New SNapshot ServiceAccount
	if err := c.createServiceAccount(mysql, mysql.SnapshotSAName()); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	// Create New PSP for Snapshot
	if err := c.ensurePSP(mysql, mysql.SnapshotSAName()); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	// Create New Role for Snapshot
	if err := c.ensureRole(mysql, mysql.SnapshotSAName()); err != nil {
		return err
	}

	// Create New RoleBinding for Snapshot
	if err := c.createRoleBinding(mysql, mysql.SnapshotSAName()); err != nil {
		return err
	}

	return nil
}
