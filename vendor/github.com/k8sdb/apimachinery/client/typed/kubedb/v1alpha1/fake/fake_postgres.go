/*
Copyright 2017 The KubeDB Authors.

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

package fake

import (
	v1alpha1 "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePostgreses implements PostgresInterface
type FakePostgreses struct {
	Fake *FakeKubedbV1alpha1
	ns   string
}

var postgresesResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "v1alpha1", Resource: "postgreses"}

var postgresesKind = schema.GroupVersionKind{Group: "kubedb.com", Version: "v1alpha1", Kind: "Postgres"}

func (c *FakePostgreses) Create(postgres *v1alpha1.Postgres) (result *v1alpha1.Postgres, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(postgresesResource, c.ns, postgres), &v1alpha1.Postgres{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Postgres), err
}

func (c *FakePostgreses) Update(postgres *v1alpha1.Postgres) (result *v1alpha1.Postgres, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(postgresesResource, c.ns, postgres), &v1alpha1.Postgres{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Postgres), err
}

func (c *FakePostgreses) UpdateStatus(postgres *v1alpha1.Postgres) (*v1alpha1.Postgres, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(postgresesResource, "status", c.ns, postgres), &v1alpha1.Postgres{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Postgres), err
}

func (c *FakePostgreses) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(postgresesResource, c.ns, name), &v1alpha1.Postgres{})

	return err
}

func (c *FakePostgreses) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(postgresesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.PostgresList{})
	return err
}

func (c *FakePostgreses) Get(name string, options v1.GetOptions) (result *v1alpha1.Postgres, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(postgresesResource, c.ns, name), &v1alpha1.Postgres{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Postgres), err
}

func (c *FakePostgreses) List(opts v1.ListOptions) (result *v1alpha1.PostgresList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(postgresesResource, postgresesKind, c.ns, opts), &v1alpha1.PostgresList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PostgresList{}
	for _, item := range obj.(*v1alpha1.PostgresList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested postgreses.
func (c *FakePostgreses) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(postgresesResource, c.ns, opts))

}

// Patch applies the patch and returns the patched postgres.
func (c *FakePostgreses) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Postgres, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(postgresesResource, c.ns, name, data, subresources...), &v1alpha1.Postgres{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Postgres), err
}
