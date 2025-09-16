package main

import (
	"context"
	"fmt"

	mc "github.com/datum-cloud/k8s-mcstore/pkg/mcstore"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func main() {
	extractor := mc.ContextExtractor{}
	layout := &mc.DefaultKeyLayout{Extractor: extractor, Options: &mc.Options{GlobalSegment: "__global__"}}

	// Built-in like Pods
	pods := schema.GroupResource{Group: "core", Resource: "pods"}
	// CRD-like resource
	widgets := schema.GroupResource{Group: "acme.io", Resource: "widgets"}

	ctxA := mc.WithCluster(context.Background(), "cluster-a", false)
	p1, _ := layout.KeyForList(ctxA, pods, "ns-1")
	w1, _ := layout.KeyForList(ctxA, widgets, "")

	ctxGlobal := mc.WithCluster(context.Background(), "__global__", false)
	gPods, _ := layout.KeyForList(ctxGlobal, pods, "")
	gWidgets, _ := layout.KeyForList(ctxGlobal, widgets, "")

	fmt.Println("pods ns-1:", p1)
	fmt.Println("widgets all-ns:", w1)
	fmt.Println("global pods:", gPods)
	fmt.Println("global widgets:", gWidgets)
}
