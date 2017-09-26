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

// FakeElasticsearchs implements ElasticsearchInterface
type FakeElasticsearchs struct {
	Fake *FakeKubedbV1alpha1
	ns   string
}

var elasticsearchsResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "v1alpha1", Resource: "elasticsearchs"}

var elasticsearchsKind = schema.GroupVersionKind{Group: "kubedb.com", Version: "v1alpha1", Kind: "Elasticsearch"}

func (c *FakeElasticsearchs) Create(elasticsearch *v1alpha1.Elasticsearch) (result *v1alpha1.Elasticsearch, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(elasticsearchsResource, c.ns, elasticsearch), &v1alpha1.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Elasticsearch), err
}

func (c *FakeElasticsearchs) Update(elasticsearch *v1alpha1.Elasticsearch) (result *v1alpha1.Elasticsearch, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(elasticsearchsResource, c.ns, elasticsearch), &v1alpha1.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Elasticsearch), err
}

func (c *FakeElasticsearchs) UpdateStatus(elasticsearch *v1alpha1.Elasticsearch) (*v1alpha1.Elasticsearch, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(elasticsearchsResource, "status", c.ns, elasticsearch), &v1alpha1.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Elasticsearch), err
}

func (c *FakeElasticsearchs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(elasticsearchsResource, c.ns, name), &v1alpha1.Elasticsearch{})

	return err
}

func (c *FakeElasticsearchs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(elasticsearchsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ElasticsearchList{})
	return err
}

func (c *FakeElasticsearchs) Get(name string, options v1.GetOptions) (result *v1alpha1.Elasticsearch, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(elasticsearchsResource, c.ns, name), &v1alpha1.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Elasticsearch), err
}

func (c *FakeElasticsearchs) List(opts v1.ListOptions) (result *v1alpha1.ElasticsearchList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(elasticsearchsResource, elasticsearchsKind, c.ns, opts), &v1alpha1.ElasticsearchList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ElasticsearchList{}
	for _, item := range obj.(*v1alpha1.ElasticsearchList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested elasticsearchs.
func (c *FakeElasticsearchs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(elasticsearchsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched elasticsearch.
func (c *FakeElasticsearchs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Elasticsearch, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(elasticsearchsResource, c.ns, name, data, subresources...), &v1alpha1.Elasticsearch{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Elasticsearch), err
}
