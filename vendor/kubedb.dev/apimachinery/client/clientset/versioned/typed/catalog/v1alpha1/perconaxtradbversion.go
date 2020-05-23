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

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	scheme "kubedb.dev/apimachinery/client/clientset/versioned/scheme"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PerconaXtraDBVersionsGetter has a method to return a PerconaXtraDBVersionInterface.
// A group's client should implement this interface.
type PerconaXtraDBVersionsGetter interface {
	PerconaXtraDBVersions() PerconaXtraDBVersionInterface
}

// PerconaXtraDBVersionInterface has methods to work with PerconaXtraDBVersion resources.
type PerconaXtraDBVersionInterface interface {
	Create(ctx context.Context, perconaXtraDBVersion *v1alpha1.PerconaXtraDBVersion, opts v1.CreateOptions) (*v1alpha1.PerconaXtraDBVersion, error)
	Update(ctx context.Context, perconaXtraDBVersion *v1alpha1.PerconaXtraDBVersion, opts v1.UpdateOptions) (*v1alpha1.PerconaXtraDBVersion, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.PerconaXtraDBVersion, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.PerconaXtraDBVersionList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PerconaXtraDBVersion, err error)
	PerconaXtraDBVersionExpansion
}

// perconaXtraDBVersions implements PerconaXtraDBVersionInterface
type perconaXtraDBVersions struct {
	client rest.Interface
}

// newPerconaXtraDBVersions returns a PerconaXtraDBVersions
func newPerconaXtraDBVersions(c *CatalogV1alpha1Client) *perconaXtraDBVersions {
	return &perconaXtraDBVersions{
		client: c.RESTClient(),
	}
}

// Get takes name of the perconaXtraDBVersion, and returns the corresponding perconaXtraDBVersion object, and an error if there is any.
func (c *perconaXtraDBVersions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PerconaXtraDBVersion, err error) {
	result = &v1alpha1.PerconaXtraDBVersion{}
	err = c.client.Get().
		Resource("perconaxtradbversions").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PerconaXtraDBVersions that match those selectors.
func (c *perconaXtraDBVersions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PerconaXtraDBVersionList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.PerconaXtraDBVersionList{}
	err = c.client.Get().
		Resource("perconaxtradbversions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested perconaXtraDBVersions.
func (c *perconaXtraDBVersions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("perconaxtradbversions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a perconaXtraDBVersion and creates it.  Returns the server's representation of the perconaXtraDBVersion, and an error, if there is any.
func (c *perconaXtraDBVersions) Create(ctx context.Context, perconaXtraDBVersion *v1alpha1.PerconaXtraDBVersion, opts v1.CreateOptions) (result *v1alpha1.PerconaXtraDBVersion, err error) {
	result = &v1alpha1.PerconaXtraDBVersion{}
	err = c.client.Post().
		Resource("perconaxtradbversions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(perconaXtraDBVersion).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a perconaXtraDBVersion and updates it. Returns the server's representation of the perconaXtraDBVersion, and an error, if there is any.
func (c *perconaXtraDBVersions) Update(ctx context.Context, perconaXtraDBVersion *v1alpha1.PerconaXtraDBVersion, opts v1.UpdateOptions) (result *v1alpha1.PerconaXtraDBVersion, err error) {
	result = &v1alpha1.PerconaXtraDBVersion{}
	err = c.client.Put().
		Resource("perconaxtradbversions").
		Name(perconaXtraDBVersion.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(perconaXtraDBVersion).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the perconaXtraDBVersion and deletes it. Returns an error if one occurs.
func (c *perconaXtraDBVersions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("perconaxtradbversions").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *perconaXtraDBVersions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("perconaxtradbversions").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched perconaXtraDBVersion.
func (c *perconaXtraDBVersions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PerconaXtraDBVersion, err error) {
	result = &v1alpha1.PerconaXtraDBVersion{}
	err = c.client.Patch(pt).
		Resource("perconaxtradbversions").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
