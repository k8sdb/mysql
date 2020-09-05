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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	"github.com/appscode/go/crypto/rand"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ka "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	scs "stash.appscode.dev/apimachinery/client/clientset/versioned"
)

var (
	DockerRegistry = "kubedbci"
	DBCatalogName  = "8.0.14"
)

type Framework struct {
	restConfig       *rest.Config
	kubeClient       kubernetes.Interface
	apiExtKubeClient crd_cs.ApiextensionsV1beta1Interface
	dbClient         cs.Interface
	kaClient         ka.Interface
	appCatalogClient appcat_cs.AppcatalogV1alpha1Interface
	stashClient      scs.Interface
	namespace        string
	name             string
	StorageClass     string
}

func New(
	restConfig *rest.Config,
	kubeClient kubernetes.Interface,
	apiExtKubeClient crd_cs.ApiextensionsV1beta1Interface,
	dbClient cs.Interface,
	kaClient ka.Interface,
	appCatalogClient appcat_cs.AppcatalogV1alpha1Interface,
	stashClient scs.Interface,
	storageClass string,
) *Framework {
	return &Framework{
		restConfig:       restConfig,
		kubeClient:       kubeClient,
		apiExtKubeClient: apiExtKubeClient,
		dbClient:         dbClient,
		kaClient:         kaClient,
		appCatalogClient: appCatalogClient,
		stashClient:      stashClient,
		name:             "mysql-operator",
		namespace:        rand.WithUniqSuffix(api.ResourceSingularMySQL),
		StorageClass:     storageClass,
	}
}

func (f *Framework) Invoke() *Invocation {
	return &Invocation{
		Framework: f,
		app:       rand.WithUniqSuffix("mysql-e2e"),
	}
}

func (fi *Invocation) App() string {
	return fi.app
}

func (fi *Invocation) ExtClient() cs.Interface {
	return fi.dbClient
}

type Invocation struct {
	*Framework
	app string
}
