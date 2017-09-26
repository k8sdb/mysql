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

// FakeSnapshots implements SnapshotInterface
type FakeSnapshots struct {
	Fake *FakeKubedbV1alpha1
	ns   string
}

var snapshotsResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "v1alpha1", Resource: "snapshots"}

var snapshotsKind = schema.GroupVersionKind{Group: "kubedb.com", Version: "v1alpha1", Kind: "Snapshot"}

func (c *FakeSnapshots) Create(snapshot *v1alpha1.Snapshot) (result *v1alpha1.Snapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(snapshotsResource, c.ns, snapshot), &v1alpha1.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Snapshot), err
}

func (c *FakeSnapshots) Update(snapshot *v1alpha1.Snapshot) (result *v1alpha1.Snapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(snapshotsResource, c.ns, snapshot), &v1alpha1.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Snapshot), err
}

func (c *FakeSnapshots) UpdateStatus(snapshot *v1alpha1.Snapshot) (*v1alpha1.Snapshot, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(snapshotsResource, "status", c.ns, snapshot), &v1alpha1.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Snapshot), err
}

func (c *FakeSnapshots) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(snapshotsResource, c.ns, name), &v1alpha1.Snapshot{})

	return err
}

func (c *FakeSnapshots) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(snapshotsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.SnapshotList{})
	return err
}

func (c *FakeSnapshots) Get(name string, options v1.GetOptions) (result *v1alpha1.Snapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(snapshotsResource, c.ns, name), &v1alpha1.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Snapshot), err
}

func (c *FakeSnapshots) List(opts v1.ListOptions) (result *v1alpha1.SnapshotList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(snapshotsResource, snapshotsKind, c.ns, opts), &v1alpha1.SnapshotList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SnapshotList{}
	for _, item := range obj.(*v1alpha1.SnapshotList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested snapshots.
func (c *FakeSnapshots) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(snapshotsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched snapshot.
func (c *FakeSnapshots) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Snapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(snapshotsResource, c.ns, name, data, subresources...), &v1alpha1.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Snapshot), err
}
