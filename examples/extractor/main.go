package main

import (
	"context"
	"fmt"

	mc "github.com/datum-labs/multicluster-store/pkg/multicluster-store"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func main() {
	// Suppose your request path is /clusters/cluster-a/apis/core/v1/pods
	reqPath := "/clusters/cluster-a/apis/core/v1/pods"
	ctx := mc.WithRequestPath(context.Background(), reqPath)
	extractor := mc.NewPathExtractor("/clusters/", "cluster")
	layout := &mc.DefaultKeyLayout{Extractor: extractor}

	gr := schema.GroupResource{Group: "core", Resource: "pods"}

	// Extract cluster from path and compute a list prefix
	c, all, _ := extractor.ExtractCluster(ctx)
	fmt.Println("extracted:", c, "all:", all)

	// Build a list prefix
	ctxWithCluster := mc.WithCluster(ctx, c, false)
	prefix, _ := layout.KeyForList(ctxWithCluster, gr, "default")
	fmt.Println("prefix:", prefix)
}
