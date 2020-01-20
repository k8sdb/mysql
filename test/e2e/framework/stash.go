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
package framework

import (
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/controller"

	"github.com/appscode/go/wait"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
	appcat_api "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	stash_util "stash.appscode.dev/stash/apis"
	"stash.appscode.dev/stash/apis/stash/v1alpha1"
	stashV1alpha1 "stash.appscode.dev/stash/apis/stash/v1alpha1"
	stashv1beta1 "stash.appscode.dev/stash/apis/stash/v1beta1"
)

//var (
//	//StashMySQLBackupTask  = "my-backup-8.0.14"
//	//StashMySQLRestoreTask = "my-restore-8.0.14"
//	StashMySQLBackupTask  = "my-backup-" + DBCatalogName
//	StashMySQLRestoreTask = "my-restore-" + DBCatalogName
//)
//
//const (
//	MySQLBackupTaskPrefix  = "mysql-backup-"
//	MySQLRestoreTaskPrefix = "mysql-restore-"
//	//MySQLBackupFunction  = "mysql-backup-8.0.14"
//	//MySQLRestoreFunction = "mysql-restore-8.0.14"
//)

func (f *Framework) FoundStashCRDs() bool {
	return controller.FoundStashCRDs(f.apiExtKubeClient)
}

func (f *Invocation) BackupConfiguration(meta metav1.ObjectMeta) *stashv1beta1.BackupConfiguration {
	return &stashv1beta1.BackupConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name + "-bc",
			Namespace: f.namespace,
		},
		Spec: stashv1beta1.BackupConfigurationSpec{
			Repository: core.LocalObjectReference{
				Name: meta.Name,
			},
			Schedule: "*/2 * * * *",
			RetentionPolicy: v1alpha1.RetentionPolicy{
				KeepLast: 5,
				Name:     "keep-last-5",
				Prune:    true,
			},
			BackupConfigurationTemplateSpec: stashv1beta1.BackupConfigurationTemplateSpec{
				Task: stashv1beta1.TaskRef{
					Name: "mysql-backup-" + DBCatalogName,
				},
				Target: &stashv1beta1.BackupTarget{
					Ref: stashv1beta1.TargetRef{
						APIVersion: appcat_api.SchemeGroupVersion.String(),
						Kind:       appcat_api.ResourceKindApp,
						Name:       meta.Name,
					},
				},
			},
		},
	}
}

func (f *Framework) CreateBackupConfiguration(backupCfg *stashv1beta1.BackupConfiguration) error {
	_, err := f.stashClient.StashV1beta1().BackupConfigurations(backupCfg.Namespace).Create(backupCfg)
	return err
}

func (f *Framework) DeleteBackupConfiguration(meta metav1.ObjectMeta) error {
	return f.stashClient.StashV1beta1().BackupConfigurations(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
}

func (f *Framework) WaitUntilBackkupSessionBeCreated(bcMeta metav1.ObjectMeta) (bs *stashv1beta1.BackupSession, err error) {
	err = wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		bsList, err := f.stashClient.StashV1beta1().BackupSessions(bcMeta.Namespace).List(metav1.ListOptions{
			LabelSelector: labels.Set{
				stash_util.LabelInvokerType: stashv1beta1.ResourceKindBackupConfiguration,
				stash_util.LabelInvokerName: bcMeta.Name,
			}.String(),
		})
		if err != nil {
			return false, err
		}
		if len(bsList.Items) == 0 {
			return false, nil
		}

		bs = &bsList.Items[0]

		return true, nil
	})

	return
}

func (f *Framework) EventuallyBackupSessionPhase(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() (phase stashv1beta1.BackupSessionPhase) {
			bs, err := f.stashClient.StashV1beta1().BackupSessions(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return bs.Status.Phase
		},
	)
}

func (f *Invocation) Repository(meta metav1.ObjectMeta, secretName string) *stashV1alpha1.Repository {
	return &stashV1alpha1.Repository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: f.namespace,
		},
	}
}

func (f *Framework) CreateRepository(repo *stashV1alpha1.Repository) error {
	_, err := f.stashClient.StashV1alpha1().Repositories(repo.Namespace).Create(repo)
	return err
}

func (f *Framework) DeleteRepository(meta metav1.ObjectMeta) error {
	err := f.stashClient.StashV1alpha1().Repositories(meta.Namespace).Delete(meta.Name, deleteInBackground())
	return err
}

func (f *Invocation) RestoreSession(meta, oldMeta metav1.ObjectMeta) *stashv1beta1.RestoreSession {
	return &stashv1beta1.RestoreSession{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name + "-rs",
			Namespace: f.namespace,
			Labels: map[string]string{
				"app":                 f.app,
				api.LabelDatabaseKind: api.ResourceKindMySQL,
			},
		},
		Spec: stashv1beta1.RestoreSessionSpec{
			Task: stashv1beta1.TaskRef{
				Name: "mysql-restore-" + DBCatalogName,
			},
			Repository: core.LocalObjectReference{
				Name: oldMeta.Name,
			},
			Rules: []stashv1beta1.Rule{
				{
					Snapshots: []string{"latest"},
				},
			},
			Target: &stashv1beta1.RestoreTarget{
				Ref: stashv1beta1.TargetRef{
					APIVersion: appcat_api.SchemeGroupVersion.String(),
					Kind:       appcat_api.ResourceKindApp,
					Name:       meta.Name,
				},
			},
		},
	}
}

func (f *Framework) CreateRestoreSession(restoreSession *stashv1beta1.RestoreSession) error {
	_, err := f.stashClient.StashV1beta1().RestoreSessions(restoreSession.Namespace).Create(restoreSession)
	return err
}

func (f *Framework) DeleteRestoreSession(meta metav1.ObjectMeta) error {
	err := f.stashClient.StashV1beta1().RestoreSessions(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
	return err
}

func (f *Framework) EventuallyRestoreSessionPhase(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(func() stashv1beta1.RestoreSessionPhase {
		restoreSession, err := f.stashClient.StashV1beta1().RestoreSessions(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return restoreSession.Status.Phase
	},
		time.Minute*7,
		time.Second*7,
	)
}
