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

package internalversion

import (
	kubedb "github.com/k8sdb/apimachinery/apis/kubedb"
	scheme "github.com/k8sdb/apimachinery/client/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PostgresesGetter has a method to return a PostgresInterface.
// A group's client should implement this interface.
type PostgresesGetter interface {
	Postgreses(namespace string) PostgresInterface
}

// PostgresInterface has methods to work with Postgres resources.
type PostgresInterface interface {
	Create(*kubedb.Postgres) (*kubedb.Postgres, error)
	Update(*kubedb.Postgres) (*kubedb.Postgres, error)
	UpdateStatus(*kubedb.Postgres) (*kubedb.Postgres, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*kubedb.Postgres, error)
	List(opts v1.ListOptions) (*kubedb.PostgresList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubedb.Postgres, err error)
	PostgresExpansion
}

// postgreses implements PostgresInterface
type postgreses struct {
	client rest.Interface
	ns     string
}

// newPostgreses returns a Postgreses
func newPostgreses(c *KubedbClient, namespace string) *postgreses {
	return &postgreses{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Create takes the representation of a postgres and creates it.  Returns the server's representation of the postgres, and an error, if there is any.
func (c *postgreses) Create(postgres *kubedb.Postgres) (result *kubedb.Postgres, err error) {
	result = &kubedb.Postgres{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("postgreses").
		Body(postgres).
		Do().
		Into(result)
	return
}

// Update takes the representation of a postgres and updates it. Returns the server's representation of the postgres, and an error, if there is any.
func (c *postgreses) Update(postgres *kubedb.Postgres) (result *kubedb.Postgres, err error) {
	result = &kubedb.Postgres{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("postgreses").
		Name(postgres.Name).
		Body(postgres).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclientstatus=false comment above the type to avoid generating UpdateStatus().

func (c *postgreses) UpdateStatus(postgres *kubedb.Postgres) (result *kubedb.Postgres, err error) {
	result = &kubedb.Postgres{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("postgreses").
		Name(postgres.Name).
		SubResource("status").
		Body(postgres).
		Do().
		Into(result)
	return
}

// Delete takes name of the postgres and deletes it. Returns an error if one occurs.
func (c *postgreses) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("postgreses").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *postgreses) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("postgreses").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the postgres, and returns the corresponding postgres object, and an error if there is any.
func (c *postgreses) Get(name string, options v1.GetOptions) (result *kubedb.Postgres, err error) {
	result = &kubedb.Postgres{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("postgreses").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Postgreses that match those selectors.
func (c *postgreses) List(opts v1.ListOptions) (result *kubedb.PostgresList, err error) {
	result = &kubedb.PostgresList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("postgreses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested postgreses.
func (c *postgreses) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("postgreses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched postgres.
func (c *postgreses) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubedb.Postgres, err error) {
	result = &kubedb.Postgres{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("postgreses").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
