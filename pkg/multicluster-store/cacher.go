//go:build mcstore_cacher

package mcstore

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/cacher"
)

// CacherBuilder constructs a cacher with cluster-aware indexing over a backend storage.
type CacherBuilder struct {
	Backend   storage.Interface
	KeyLayout KeyLayout
	Extractor ClusterExtractor
	Capacity  int
}

func (b *CacherBuilder) For(gr schema.GroupResource, newFunc func() runtime.Object) (storage.Interface, error) {
	lw := NewListerWatcher(b.Backend, b.KeyLayout, gr, newFunc)
	cfg := &cacher.Config{
		CacheCapacity: b.Capacity,
		ObjectType:    newFunc(),
		Storage:       b.Backend,
		ListerWatcher: lw,
		NewFunc:       newFunc,
		GetAttrsFunc:  ObjectAttrsFunc(newFunc),
	}
	return cacher.NewCacherFromConfig(cfg)
}
