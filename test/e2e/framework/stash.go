package framework

import (
	"time"

	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat_api "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/controller"
	"stash.appscode.dev/stash/apis/stash/v1alpha1"
	stashV1alpha1 "stash.appscode.dev/stash/apis/stash/v1alpha1"
	"stash.appscode.dev/stash/apis/stash/v1beta1"
)

var (
	StashMySQLBackupTask  = "my-backup-8.0.14"
	StashMySQLRestoreTask = "my-restore-8.0.14"
)

func (f *Framework) FoundStashCRDs() bool {
	return controller.FoundStashCRDs(f.apiExtKubeClient)
}

func (f *Invocation) BackupConfiguration(meta metav1.ObjectMeta) *v1beta1.BackupConfiguration {
	return &v1beta1.BackupConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: f.namespace,
		},
		Spec: v1beta1.BackupConfigurationSpec{
			Task: v1beta1.TaskRef{
				Name: StashMySQLBackupTask,
			},
			Repository: core.LocalObjectReference{
				Name: meta.Name,
			},
			Schedule: "*/3 * * * *",
			Target: &v1beta1.BackupTarget{
				Ref: v1beta1.TargetRef{
					APIVersion: appcat_api.SchemeGroupVersion.String(),
					Kind:       appcat_api.ResourceKindApp,
					Name:       meta.Name,
				},
			},
			RetentionPolicy: v1alpha1.RetentionPolicy{
				KeepLast: 5,
				Prune:    true,
			},
		},
	}
}

func (f *Framework) CreateBackupConfiguration(backupCfg *v1beta1.BackupConfiguration) error {
	_, err := f.stashClient.StashV1beta1().BackupConfigurations(backupCfg.Namespace).Create(backupCfg)
	return err
}

func (f *Framework) DeleteBackupConfiguration(meta metav1.ObjectMeta) error {
	return f.stashClient.StashV1beta1().BackupConfigurations(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
}

func (f *Framework) EventuallyBackupSessionPhase(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() (phase v1beta1.BackupSessionPhase) {
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

func (f *Invocation) BackupSession(meta metav1.ObjectMeta) *v1beta1.BackupSession {
	return &v1beta1.BackupSession{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: f.namespace,
		},
		Spec: v1beta1.BackupSessionSpec{
			BackupConfiguration: core.LocalObjectReference{
				Name: meta.Name,
			},
		},
	}
}

func (f *Framework) CreateBackupSession(bc *v1beta1.BackupSession) error {
	_, err := f.stashClient.StashV1beta1().BackupSessions(bc.Namespace).Create(bc)

	return err

}

func (f *Framework) DeleteBackupSession(meta metav1.ObjectMeta) error {
	err := f.stashClient.StashV1beta1().BackupSessions(meta.Namespace).Delete(meta.Name, deleteInBackground())
	return err
}

func (f *Invocation) RestoreSession(meta, oldMeta metav1.ObjectMeta) *v1beta1.RestoreSession {
	return &v1beta1.RestoreSession{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: f.namespace,
			Labels: map[string]string{
				"app":                 f.app,
				api.LabelDatabaseKind: api.ResourceKindMySQL,
			},
		},
		Spec: v1beta1.RestoreSessionSpec{
			Task: v1beta1.TaskRef{
				Name: StashMySQLRestoreTask,
			},
			Repository: core.LocalObjectReference{
				Name: oldMeta.Name,
			},
			Rules: []v1beta1.Rule{
				{
					Snapshots: []string{"latest"},
				},
			},
			Target: &v1beta1.RestoreTarget{
				Ref: v1beta1.TargetRef{
					APIVersion: appcat_api.SchemeGroupVersion.String(),
					Kind:       appcat_api.ResourceKindApp,
					Name:       meta.Name,
				},
			},
		},
	}
}

func (f *Framework) CreateRestoreSession(restoreSession *v1beta1.RestoreSession) error {
	_, err := f.stashClient.StashV1beta1().RestoreSessions(restoreSession.Namespace).Create(restoreSession)
	return err
}

func (f *Framework) DeleteRestoreSession(meta metav1.ObjectMeta) error {
	err := f.stashClient.StashV1beta1().RestoreSessions(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
	return err
}

func (f *Framework) EventuallyRestoreSessionPhase(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(func() v1beta1.RestoreSessionPhase {
		restoreSession, err := f.stashClient.StashV1beta1().RestoreSessions(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return restoreSession.Status.Phase
	},
		time.Minute*7,
		time.Second*7,
	)
}
