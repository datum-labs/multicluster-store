//go:build mcstore_cacher

package mcstore

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/storage"
)

// listerWatcherAdapter implements cacher's ListerWatcher interface using our KeyLayout.
type listerWatcherAdapter struct {
	backend storage.Interface
	layout  KeyLayout
	gr      schema.GroupResource
	newFunc func() runtime.Object
}

func NewListerWatcher(backend storage.Interface, layout KeyLayout, gr schema.GroupResource, newFunc func() runtime.Object) storage.ListerWatcher {
	return &listerWatcherAdapter{backend: backend, layout: layout, gr: gr, newFunc: newFunc}
}

func (a *listerWatcherAdapter) List(ctx context.Context, options storage.ListOptions) (runtime.Object, string, error) {
	// Always list on the kind root to cover all clusters; filtering is done by cacher indexes/predicate.
	prefix := a.layout.KeyRoot(a.gr)
	opts := storage.ListOptions{ResourceVersion: options.ResourceVersion, Predicate: options.Predicate, Recursive: true}
	return a.backend.List(ctx, prefix, opts)
}

func (a *listerWatcherAdapter) Watch(ctx context.Context, options storage.ListOptions) (storage.WatchInterface, error) {
	prefix := a.layout.KeyRoot(a.gr)
	opts := storage.ListOptions{ResourceVersion: options.ResourceVersion, Predicate: options.Predicate, Recursive: true}
	return a.backend.Watch(ctx, prefix, opts)
}
