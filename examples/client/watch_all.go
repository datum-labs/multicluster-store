package main

import (
	"context"
	"fmt"
	"time"

	mc "github.com/datum-cloud/k8s-mcstore/pkg/mcstore"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// This example demonstrates computing keys and predicates for cross-cluster and per-cluster operations.
func main() {
	extractor := mc.ContextExtractor{}
	layout := &mc.DefaultKeyLayout{Extractor: extractor, Options: &mc.Options{GlobalSegment: "__global__"}}

	gr := schema.GroupResource{Group: "core", Resource: "pods"}

	// Simulate a request for all clusters
	ctxAll := mc.WithCluster(context.Background(), "", true)
	root, _ := layout.KeyForList(ctxAll, gr, "")
	fmt.Println("List/Watch root for kind (all clusters):", root)

	// Simulate per-cluster operations
	ctxA := mc.WithCluster(context.Background(), "cluster-a", false)
	ctxB := mc.WithCluster(context.Background(), "cluster-b", false)

	keyAListNS, _ := layout.KeyForList(ctxA, gr, "ns-1")
	keyBListAll, _ := layout.KeyForList(ctxB, gr, "")
	fmt.Println("List prefix cluster-a ns-1:", keyAListNS)
	fmt.Println("List prefix cluster-b all-namespaces:", keyBListAll)

	// Global objects example
	ctxGlobal := mc.WithCluster(context.Background(), "__global__", false)
	globalPrefix, _ := layout.KeyForList(ctxGlobal, gr, "")
	fmt.Println("List prefix global:", globalPrefix)

	// Show field selector composition for per-cluster predicate
	predAll := mc.WithClusterPredicate(mc.EmptyPredicate(), "", true)
	predA := mc.WithClusterPredicate(mc.EmptyPredicate(), "cluster-a", false)
	fmt.Println("All clusters predicate field empty?", predAll.Field == nil || predAll.Field.Empty())
	fmt.Println("Cluster-a predicate field:", predA.Field.String())

	// Sleep briefly to make this trivial program more obvious when run
	time.Sleep(100 * time.Millisecond)
}
