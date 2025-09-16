package main

import (
	"context"
	"fmt"

	mc "github.com/datum-labs/multicluster-store/pkg/multicluster-store"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func main() {
	extractor := mc.ContextExtractor{}
	layout := &mc.DefaultKeyLayout{Extractor: extractor}
	gr := schema.GroupResource{Group: "core", Resource: "pods"}

	// Kind root strategy (single prefix across clusters)
	kind := mc.KindRootStrategy{}
	p1, _ := kind.WatchPrefixes(context.Background(), layout, gr)
	fmt.Println("kind-root:", p1)

	// Per-cluster strategy (one prefix per cluster)
	pcs := mc.PerClusterStrategy{Source: mc.StaticClusterSource{Clusters: []string{"cluster-a", "cluster-b"}}}
	p2, _ := pcs.WatchPrefixes(context.Background(), layout, gr)
	fmt.Println("per-cluster:", p2)
}
