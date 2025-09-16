package main

import (
	"context"
	"fmt"

	mc "github.com/datum-cloud/k8s-mcstore/pkg/mcstore"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func show(mode mc.PredicateMode) {
	mc.DefaultOptions = &mc.Options{Predicate: mode, GlobalSegment: "__global__", ClusterFieldKey: "metadata.cluster"}
	extractor := mc.ContextExtractor{}
	layout := &mc.DefaultKeyLayout{Extractor: extractor, Options: mc.DefaultOptions}
	gr := schema.GroupResource{Group: "core", Resource: "pods"}
	ctx := mc.WithCluster(context.Background(), "cluster-a", false)
	prefix, _ := layout.KeyForList(ctx, gr, "default")
	pred := mc.WithClusterPredicate(mc.EmptyPredicate(), "cluster-a", false)
	fieldStr := ""
	if pred.Field != nil && !pred.Field.Empty() {
		fieldStr = pred.Field.String()
	}
	fmt.Printf("mode=%v prefix=%s field=%s\n", mode, prefix, fieldStr)
}

func main() {
	show(mc.PredicatePathOnly)
	show(mc.PredicateMetadataOnly)
	show(mc.PredicateHybrid)
}
