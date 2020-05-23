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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "kubedb.dev/apimachinery/apis/catalog/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMemcachedVersions implements MemcachedVersionInterface
type FakeMemcachedVersions struct {
	Fake *FakeCatalogV1alpha1
}

var memcachedversionsResource = schema.GroupVersionResource{Group: "catalog.kubedb.com", Version: "v1alpha1", Resource: "memcachedversions"}

var memcachedversionsKind = schema.GroupVersionKind{Group: "catalog.kubedb.com", Version: "v1alpha1", Kind: "MemcachedVersion"}

// Get takes name of the memcachedVersion, and returns the corresponding memcachedVersion object, and an error if there is any.
func (c *FakeMemcachedVersions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MemcachedVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(memcachedversionsResource, name), &v1alpha1.MemcachedVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MemcachedVersion), err
}

// List takes label and field selectors, and returns the list of MemcachedVersions that match those selectors.
func (c *FakeMemcachedVersions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MemcachedVersionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(memcachedversionsResource, memcachedversionsKind, opts), &v1alpha1.MemcachedVersionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MemcachedVersionList{ListMeta: obj.(*v1alpha1.MemcachedVersionList).ListMeta}
	for _, item := range obj.(*v1alpha1.MemcachedVersionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested memcachedVersions.
func (c *FakeMemcachedVersions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(memcachedversionsResource, opts))
}

// Create takes the representation of a memcachedVersion and creates it.  Returns the server's representation of the memcachedVersion, and an error, if there is any.
func (c *FakeMemcachedVersions) Create(ctx context.Context, memcachedVersion *v1alpha1.MemcachedVersion, opts v1.CreateOptions) (result *v1alpha1.MemcachedVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(memcachedversionsResource, memcachedVersion), &v1alpha1.MemcachedVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MemcachedVersion), err
}

// Update takes the representation of a memcachedVersion and updates it. Returns the server's representation of the memcachedVersion, and an error, if there is any.
func (c *FakeMemcachedVersions) Update(ctx context.Context, memcachedVersion *v1alpha1.MemcachedVersion, opts v1.UpdateOptions) (result *v1alpha1.MemcachedVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(memcachedversionsResource, memcachedVersion), &v1alpha1.MemcachedVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MemcachedVersion), err
}

// Delete takes name of the memcachedVersion and deletes it. Returns an error if one occurs.
func (c *FakeMemcachedVersions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(memcachedversionsResource, name), &v1alpha1.MemcachedVersion{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMemcachedVersions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(memcachedversionsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MemcachedVersionList{})
	return err
}

// Patch applies the patch and returns the patched memcachedVersion.
func (c *FakeMemcachedVersions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MemcachedVersion, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(memcachedversionsResource, name, pt, data, subresources...), &v1alpha1.MemcachedVersion{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MemcachedVersion), err
}
