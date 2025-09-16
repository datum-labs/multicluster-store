package mcstore

import (
	"context"
	"path"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// KeyLayout defines how to compute storage keys and roots.
type KeyLayout interface {
	KeyRoot(gr schema.GroupResource) string
	KeyForObject(ctx context.Context, gr schema.GroupResource, obj runtime.Object) (string, error)
	KeyForList(ctx context.Context, gr schema.GroupResource, namespace string) (string, error)
}

type DefaultKeyLayout struct {
	Extractor ClusterExtractor
	Options   *Options
}

func (l *DefaultKeyLayout) KeyRoot(gr schema.GroupResource) string {
	group := gr.Group
	if group == "" {
		group = "core"
	}
	opts := effectiveOptions(l.Options)
	// Ensure trailing slash for prefix watches/lists
	return ensureTrailingSlash(path.Join(opts.EtcdPrefix, group, gr.Resource))
}

func (l *DefaultKeyLayout) KeyForObject(ctx context.Context, gr schema.GroupResource, obj runtime.Object) (string, error) {
	cluster, _, err := l.Extractor.ExtractCluster(ctx)
	if err != nil {
		return "", err
	}
	ns := getNamespace(obj)
	name := getName(obj)
	root := strings.TrimSuffix(l.KeyRoot(gr), "/")
	global := effectiveOptions(l.Options).GlobalSegment
	if ns == "" {
		if cluster == global {
			return path.Join(root, global, name), nil
		}
		return path.Join(root, cluster, name), nil
	}
	if cluster == global {
		return path.Join(root, global, ns, name), nil
	}
	return path.Join(root, cluster, ns, name), nil
}

func (l *DefaultKeyLayout) KeyForList(ctx context.Context, gr schema.GroupResource, namespace string) (string, error) {
	cluster, all, err := l.Extractor.ExtractCluster(ctx)
	if err != nil {
		return "", err
	}
	root := l.KeyRoot(gr)
	global := effectiveOptions(l.Options).GlobalSegment
	if all {
		// Cross-cluster list: list on the kind root
		return root, nil
	}
	if namespace == "" {
		if cluster == global {
			return ensureTrailingSlash(path.Join(strings.TrimSuffix(root, "/"), global)), nil
		}
		return ensureTrailingSlash(path.Join(strings.TrimSuffix(root, "/"), cluster)), nil
	}
	if cluster == global {
		return ensureTrailingSlash(path.Join(strings.TrimSuffix(root, "/"), global, namespace)), nil
	}
	return ensureTrailingSlash(path.Join(strings.TrimSuffix(root, "/"), cluster, namespace)), nil
}

func ensureTrailingSlash(s string) string {
	if strings.HasSuffix(s, "/") {
		return s
	}
	return s + "/"
}
