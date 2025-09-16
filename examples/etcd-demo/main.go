package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	mc "github.com/datum-cloud/k8s-mcstore/pkg/mcstore"
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func newObj(name, namespace string, labels map[string]string, annotations map[string]string) runtime.Object {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("v1")
	u.SetKind("Pod")
	u.SetName(name)
	u.SetNamespace(namespace)
	u.SetLabels(labels)
	u.SetAnnotations(annotations)
	return u
}

func main() {
	// Customize mcstore behavior (optional)
	mc.DefaultOptions = &mc.Options{
		ClusterAnnotationKey: "example.io/project",
		ClusterFieldKey:      "metadata.project",
		GlobalSegment:        "__global__",
		Predicate:            mc.PredicateHybrid,
	}

	endpoints := os.Getenv("ETCD_ENDPOINTS")
	if endpoints == "" {
		endpoints = "127.0.0.1:2379"
	}
	eps := strings.Split(endpoints, ",")

	cli, err := clientv3.New(clientv3.Config{Endpoints: eps, DialTimeout: 5 * time.Second})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	extractor := mc.ContextExtractor{}
	layout := &mc.DefaultKeyLayout{Extractor: extractor, Options: mc.DefaultOptions}
	gr := schema.GroupResource{Group: "core", Resource: "pods"}

	ctxAll := mc.WithCluster(context.Background(), "", true)
	root := layout.KeyRoot(gr)
	fmt.Println("Watching prefix:", root)

	watchCtx, cancel := context.WithCancel(ctxAll)
	defer cancel()
	go func() {
		wch := cli.Watch(watchCtx, root, clientv3.WithPrefix())
		for resp := range wch {
			for _, ev := range resp.Events {
				// derive cluster from key path: /registry/<group>/<resource>/<cluster>/...
				key := string(ev.Kv.Key)
				parts := strings.Split(key, "/")
				cluster := "?"
				if len(parts) >= 5 {
					cluster = parts[4]
				}
				fmt.Printf("event=%s cluster=%s key=%s\n", ev.Type, cluster, key)
			}
		}
	}()

	// Give the watcher a moment to establish
	time.Sleep(300 * time.Millisecond)

	// Write a couple of objects in different "projects"
	objA := newObj("pod-1", "ns-1", map[string]string{"app": "demo"}, map[string]string{"example.io/project": "project-a"})
	objB := newObj("pod-2", "ns-2", map[string]string{"app": "demo"}, map[string]string{"example.io/project": "project-b"})

	ctxA := mc.WithCluster(context.Background(), "project-a", false)
	keyA, err := layout.KeyForObject(ctxA, gr, objA)
	if err != nil {
		panic(err)
	}
	ctxB := mc.WithCluster(context.Background(), "project-b", false)
	keyB, err := layout.KeyForObject(ctxB, gr, objB)
	if err != nil {
		panic(err)
	}

	ba, _ := json.Marshal(objA)
	bb, _ := json.Marshal(objB)

	if _, err := cli.Put(context.Background(), keyA, string(ba)); err != nil {
		panic(err)
	}
	if _, err := cli.Put(context.Background(), keyB, string(bb)); err != nil {
		panic(err)
	}

	// Update one object (via meta.Accessor)
	acc, _ := meta.Accessor(objA)
	anns := acc.GetAnnotations()
	if anns == nil {
		anns = map[string]string{}
	}
	anns["note"] = "updated"
	acc.SetAnnotations(anns)
	ba2, _ := json.Marshal(objA)
	if _, err := cli.Put(context.Background(), keyA, string(ba2)); err != nil {
		panic(err)
	}

	// Delete the other object
	if _, err := cli.Delete(context.Background(), keyB); err != nil {
		panic(err)
	}

	// Let events flush
	time.Sleep(1 * time.Second)
}
