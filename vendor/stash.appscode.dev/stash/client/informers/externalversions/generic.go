/*
Copyright 2019 The Stash Authors.

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

// Code generated by informer-gen. DO NOT EDIT.

package externalversions

import (
	"fmt"

	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
	v1alpha1 "stash.appscode.dev/stash/apis/stash/v1alpha1"
	v1beta1 "stash.appscode.dev/stash/apis/stash/v1beta1"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=stash.appscode.com, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithResource("recoveries"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Stash().V1alpha1().Recoveries().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("repositories"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Stash().V1alpha1().Repositories().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("restics"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Stash().V1alpha1().Restics().Informer()}, nil

		// Group=stash.appscode.com, Version=v1beta1
	case v1beta1.SchemeGroupVersion.WithResource("backupconfigurations"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Stash().V1beta1().BackupConfigurations().Informer()}, nil
	case v1beta1.SchemeGroupVersion.WithResource("backupconfigurationtemplates"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Stash().V1beta1().BackupConfigurationTemplates().Informer()}, nil
	case v1beta1.SchemeGroupVersion.WithResource("backupsessions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Stash().V1beta1().BackupSessions().Informer()}, nil
	case v1beta1.SchemeGroupVersion.WithResource("functions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Stash().V1beta1().Functions().Informer()}, nil
	case v1beta1.SchemeGroupVersion.WithResource("restoresessions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Stash().V1beta1().RestoreSessions().Informer()}, nil
	case v1beta1.SchemeGroupVersion.WithResource("tasks"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Stash().V1beta1().Tasks().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
