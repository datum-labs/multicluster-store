package mcstore

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// WatchStrategy produces one or more etcd prefixes to list/watch for a given kind.
// Implementations can choose a single kind root prefix or multiple per-cluster prefixes.
type WatchStrategy interface {
	WatchPrefixes(ctx context.Context, layout KeyLayout, gr schema.GroupResource) ([]string, error)
}

// KindRootStrategy returns a single prefix: the kind root across all clusters.
type KindRootStrategy struct{}

func (KindRootStrategy) WatchPrefixes(ctx context.Context, layout KeyLayout, gr schema.GroupResource) ([]string, error) {
	return []string{layout.KeyRoot(gr)}, nil
}

// ClusterSource returns the set of logical clusters to watch.
type ClusterSource interface {
	ListClusters(ctx context.Context) ([]string, error)
}

// StaticClusterSource is a simple static implementation for tests/examples.
type StaticClusterSource struct{ Clusters []string }

func (s StaticClusterSource) ListClusters(ctx context.Context) ([]string, error) {
	return s.Clusters, nil
}

// PerClusterStrategy returns one prefix per (cluster, kind), using the layout to compute per-cluster roots.
type PerClusterStrategy struct{ Source ClusterSource }

func (p PerClusterStrategy) WatchPrefixes(ctx context.Context, layout KeyLayout, gr schema.GroupResource) ([]string, error) {
	clusters, err := p.Source.ListClusters(ctx)
	if err != nil {
		return nil, err
	}
	prefixes := make([]string, 0, len(clusters))
	for _, c := range clusters {
		cctx := WithCluster(ctx, c, false)
		prefix, err := layout.KeyForList(cctx, gr, "")
		if err != nil {
			return nil, err
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}
