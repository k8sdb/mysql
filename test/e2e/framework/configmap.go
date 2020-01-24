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
	"bytes"
	"strings"

	shell "github.com/codeskyblue/go-sh"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) InitScriptConfigMap() (*core.ConfigMap, error) {
	sh := shell.NewSession()

	var execOut bytes.Buffer
	sh.Stdout = &execOut
	if err := sh.Command("which", "curl").Run(); err != nil {
		return nil, err
	}

	curlLoc := strings.TrimSpace(execOut.String())
	execOut.Reset()

	sh.ShowCMD = true
	if err := sh.Command(curlLoc, "-fsSL", "https://github.com/kubedb/mysql-init-scripts/raw/master/init.sql").Run(); err != nil {
		return nil, err
	}

	return &core.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-init-script",
			Namespace: f.namespace,
		},
		Data: map[string]string{
			"init.sql": execOut.String(),
		},
	}, nil
}

func (f *Invocation) GetCustomConfig(configs []string) *core.ConfigMap {
	configs = append([]string{"[mysqld]"}, configs...)
	return &core.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.app,
			Namespace: f.namespace,
		},
		Data: map[string]string{
			"my-custom.cnf": strings.Join(configs, "\n"),
		},
	}
}

func (f *Invocation) CreateConfigMap(obj *core.ConfigMap) error {
	_, err := f.kubeClient.CoreV1().ConfigMaps(obj.Namespace).Create(obj)
	return err
}

func (f *Framework) DeleteConfigMap(meta metav1.ObjectMeta) error {
	err := f.kubeClient.CoreV1().ConfigMaps(meta.Namespace).Delete(meta.Name, deleteInForeground())
	if !kerr.IsNotFound(err) {
		return err
	}
	return nil
}
