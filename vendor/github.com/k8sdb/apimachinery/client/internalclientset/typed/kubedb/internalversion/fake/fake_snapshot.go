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
	kubedb "github.com/k8sdb/apimachinery/apis/kubedb"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSnapshots implements SnapshotInterface
type FakeSnapshots struct {
	Fake *FakeKubedb
	ns   string
}

var snapshotsResource = schema.GroupVersionResource{Group: "kubedb.com", Version: "", Resource: "snapshots"}

var snapshotsKind = schema.GroupVersionKind{Group: "kubedb.com", Version: "", Kind: "Snapshot"}

func (c *FakeSnapshots) Create(snapshot *kubedb.Snapshot) (result *kubedb.Snapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(snapshotsResource, c.ns, snapshot), &kubedb.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Snapshot), err
}

func (c *FakeSnapshots) Update(snapshot *kubedb.Snapshot) (result *kubedb.Snapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(snapshotsResource, c.ns, snapshot), &kubedb.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Snapshot), err
}

func (c *FakeSnapshots) UpdateStatus(snapshot *kubedb.Snapshot) (*kubedb.Snapshot, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(snapshotsResource, "status", c.ns, snapshot), &kubedb.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Snapshot), err
}

func (c *FakeSnapshots) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(snapshotsResource, c.ns, name), &kubedb.Snapshot{})

	return err
}

func (c *FakeSnapshots) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(snapshotsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kubedb.SnapshotList{})
	return err
}

func (c *FakeSnapshots) Get(name string, options v1.GetOptions) (result *kubedb.Snapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(snapshotsResource, c.ns, name), &kubedb.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Snapshot), err
}

func (c *FakeSnapshots) List(opts v1.ListOptions) (result *kubedb.SnapshotList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(snapshotsResource, snapshotsKind, c.ns, opts), &kubedb.SnapshotList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kubedb.SnapshotList{}
	for _, item := range obj.(*kubedb.SnapshotList).Items {
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
func (c *FakeSnapshots) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubedb.Snapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(snapshotsResource, c.ns, name, data, subresources...), &kubedb.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubedb.Snapshot), err
}
