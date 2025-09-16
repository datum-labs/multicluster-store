package main

import (
	"context"
	"fmt"

	mc "github.com/datum-cloud/k8s-mcstore/pkg/mcstore"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Minimal apiserver wiring sketch (no real HTTP server). This demonstrates how
// to construct mcstore components you would plug into a generic apiserver.
// Build/run with: go run ./examples/apiserver
func main() {
	extractor := mc.ContextExtractor{}
	layout := &mc.DefaultKeyLayout{Extractor: extractor, Options: &mc.Options{GlobalSegment: "__global__"}}

	gr := schema.GroupResource{Group: "core", Resource: "pods"}

	// Show the kind root used for the single per-kind List/Watch across clusters.
	fmt.Println("KeyRoot:", layout.KeyRoot(gr))

	// For demonstration, compute a cluster-scoped list prefix and a field selector predicate.
	ctx := mc.WithCluster(context.Background(), "cluster-a", false)
	listPrefix, _ := layout.KeyForList(ctx, gr, "default")
	pred := mc.WithClusterPredicate(mc.EmptyPredicate(), "cluster-a", false)

	fmt.Println("Cluster list prefix:", listPrefix)
	fmt.Println("Predicate field selector:", pred.Field.String())

	// Global segment example
	ctxGlobal := mc.WithCluster(context.Background(), "__global__", false)
	globalPrefix, _ := layout.KeyForList(ctxGlobal, gr, "")
	fmt.Println("Global list prefix:", globalPrefix)
}
