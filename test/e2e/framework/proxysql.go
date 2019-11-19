package framework

import (
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"github.com/appscode/go/crypto/rand"
	jsonTypes "github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/types"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	proxysqlPvcStorageSize = "50Mi"
)

func (f *Invocation) ProxySQL(backendObjName string) *api.ProxySQL {
	mode := api.LoadBalanceModeGroupReplication

	return &api.ProxySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("proxysql"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"app": f.app,
			},
		},
		Spec: api.ProxySQLSpec{
			Version:  jsonTypes.StrYo(ProxySQLCatalogName),
			Replicas: types.Int32P(1),
			Mode:     &mode,
			Backend: &api.ProxySQLBackendSpec{
				Ref: &corev1.TypedLocalObjectReference{
					APIGroup: types.StringP(kubedb.GroupName),
					Kind:     api.ResourceKindMySQL,
					Name:     backendObjName,
				},
				Replicas: types.Int32P(api.PerconaXtraDBDefaultClusterSize),
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
	}
}

func (f *Framework) CreateProxySQL(obj *api.ProxySQL) error {
	_, err := f.dbClient.KubedbV1alpha1().ProxySQLs(obj.Namespace).Create(obj)
	return err
}

func (f *Framework) GetProxySQL(meta metav1.ObjectMeta) (*api.ProxySQL, error) {
	return f.dbClient.KubedbV1alpha1().ProxySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
}

func (f *Framework) DeleteProxySQL(meta metav1.ObjectMeta) error {
	return f.dbClient.KubedbV1alpha1().ProxySQLs(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
}

func (f *Framework) EventuallyProxySQLPhase(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() api.DatabasePhase {
			db, err := f.dbClient.KubedbV1alpha1().ProxySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return db.Status.Phase
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) EventuallyProxySQLRunning(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			proxysql, err := f.dbClient.KubedbV1alpha1().ProxySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return proxysql.Status.Phase == api.DatabasePhaseRunning
		},
		time.Minute*15,
		time.Second*5,
	)
}
