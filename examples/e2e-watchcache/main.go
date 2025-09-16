package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	mc "github.com/datum-cloud/k8s-mcstore/pkg/mcstore"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type objEvent struct {
	Type string
	Key  string
	Obj  runtime.Object
	RV   int64
}

type subscriber struct {
	name    string
	cluster string
	ns      string // optional
	ch      chan objEvent
}

type watchCache struct {
	layout   *mc.DefaultKeyLayout
	gr       schema.GroupResource
	mu       sync.RWMutex
	byKey    map[string]runtime.Object
	byClus   map[string]map[string]struct{}
	byClusNs map[string]map[string]struct{}
	subsMu   sync.RWMutex
	subs     []*subscriber
}

func newWatchCache(layout *mc.DefaultKeyLayout, gr schema.GroupResource) *watchCache {
	return &watchCache{
		layout:   layout,
		gr:       gr,
		byKey:    map[string]runtime.Object{},
		byClus:   map[string]map[string]struct{}{},
		byClusNs: map[string]map[string]struct{}{},
	}
}

func (c *watchCache) indexKey(obj runtime.Object, key string) {
	cid := getCluster(obj)
	ns := getNS(obj)
	if _, ok := c.byClus[cid]; !ok {
		c.byClus[cid] = map[string]struct{}{}
	}
	c.byClus[cid][key] = struct{}{}
	idx := cid + "/" + ns
	if _, ok := c.byClusNs[idx]; !ok {
		c.byClusNs[idx] = map[string]struct{}{}
	}
	c.byClusNs[idx][key] = struct{}{}
}

func (c *watchCache) removeIndex(obj runtime.Object, key string) {
	cid := getCluster(obj)
	ns := getNS(obj)
	if m, ok := c.byClus[cid]; ok {
		delete(m, key)
	}
	idx := cid + "/" + ns
	if m, ok := c.byClusNs[idx]; ok {
		delete(m, key)
	}
}

func getCluster(obj runtime.Object) string {
	acc, _ := meta.Accessor(obj)
	anns := acc.GetAnnotations()
	if anns == nil {
		return ""
	}
	key := mc.DefaultOptions.ClusterAnnotationKey
	return anns[key]
}

func getNS(obj runtime.Object) string {
	acc, _ := meta.Accessor(obj)
	return acc.GetNamespace()
}

func (c *watchCache) subscribe(name, cluster, ns string) *subscriber {
	s := &subscriber{name: name, cluster: cluster, ns: ns, ch: make(chan objEvent, 64)}
	c.subsMu.Lock()
	c.subs = append(c.subs, s)
	c.subsMu.Unlock()
	return s
}

func (c *watchCache) broadcast(ev objEvent) {
	c.subsMu.RLock()
	subs := append([]*subscriber(nil), c.subs...)
	c.subsMu.RUnlock()
	for _, s := range subs {
		if s.cluster == "" { // watch-all
			s.ch <- ev
			continue
		}
		// filter by cluster/ns
		if getCluster(ev.Obj) != s.cluster {
			continue
		}
		if s.ns != "" && getNS(ev.Obj) != s.ns {
			continue
		}
		s.ch <- ev
	}
}

func (c *watchCache) seedFromEtcd(ctx context.Context, cli *clientv3.Client) error {
	root := c.layout.KeyRoot(c.gr)
	resp, err := cli.Get(ctx, root, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, kv := range resp.Kvs {
		obj := &unstructured.Unstructured{}
		if err := json.Unmarshal(kv.Value, obj); err != nil {
			continue
		}
		c.byKey[string(kv.Key)] = obj
		c.indexKey(obj, string(kv.Key))
	}
	return nil
}

func (c *watchCache) runWatcher(ctx context.Context, cli *clientv3.Client) {
	root := c.layout.KeyRoot(c.gr)
	wch := cli.Watch(ctx, root, clientv3.WithPrefix(), clientv3.WithPrevKV())
	for wr := range wch {
		for _, ev := range wr.Events {
			var obj runtime.Object
			var key string
			var rv int64
			key = string(ev.Kv.Key)
			rv = ev.Kv.ModRevision
			switch ev.Type {
			case mvccpb.PUT:
				u := &unstructured.Unstructured{}
				_ = json.Unmarshal(ev.Kv.Value, u)
				c.mu.Lock()
				old := c.byKey[key]
				c.byKey[key] = u
				if old != nil {
					c.removeIndex(old, key)
				}
				c.indexKey(u, key)
				c.mu.Unlock()
				obj = u
				c.broadcast(objEvent{Type: "PUT", Key: key, Obj: obj, RV: rv})
			case mvccpb.DELETE:
				c.mu.Lock()
				old := c.byKey[key]
				delete(c.byKey, key)
				if old != nil {
					c.removeIndex(old, key)
				}
				c.mu.Unlock()
				if old != nil {
					obj = old
				} else {
					obj = &unstructured.Unstructured{}
				}
				c.broadcast(objEvent{Type: "DELETE", Key: key, Obj: obj, RV: rv})
			}
		}
	}
}

func newObj(name, namespace, cluster string) runtime.Object {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("v1")
	u.SetKind("Pod")
	u.SetName(name)
	u.SetNamespace(namespace)
	anns := map[string]string{}
	anns[mc.DefaultOptions.ClusterAnnotationKey] = cluster
	u.SetAnnotations(anns)
	return u
}

func main() {
	// Configure mcstore for demo
	mc.DefaultOptions = &mc.Options{
		ClusterAnnotationKey: "multicluster.k8s.io/cluster",
		ClusterFieldKey:      "metadata.cluster",
		GlobalSegment:        "__global__",
		Predicate:            mc.PredicateHybrid,
		WatchStrategy:        mc.KindRootStrategy{},
	}

	endpoints := os.Getenv("ETCD_ENDPOINTS")
	if endpoints == "" {
		endpoints = "127.0.0.1:2379"
	}
	cli, err := clientv3.New(clientv3.Config{Endpoints: strings.Split(endpoints, ","), DialTimeout: 5 * time.Second})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	extractor := mc.ContextExtractor{}
	layout := &mc.DefaultKeyLayout{Extractor: extractor, Options: mc.DefaultOptions}
	gr := schema.GroupResource{Group: "core", Resource: "pods"}

	wc := newWatchCache(layout, gr)
	if err := wc.seedFromEtcd(context.Background(), cli); err != nil {
		panic(err)
	}

	// Subscribe three watchers: cluster-a, cluster-b, and all-clusters
	subA := wc.subscribe("watcher-a", "cluster-a", "")
	subB := wc.subscribe("watcher-b", "cluster-b", "")
	subAll := wc.subscribe("watcher-all", "", "")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go wc.runWatcher(ctx, cli)

	// Write some objects under single kind root; one watch stream feeds all subs
	write := func(obj runtime.Object, cluster string) {
		key, _ := layout.KeyForObject(mc.WithCluster(context.Background(), cluster, false), gr, obj)
		b, _ := json.Marshal(obj)
		if _, err := cli.Put(context.Background(), key, string(b)); err != nil {
			panic(err)
		}
	}

	del := func(obj runtime.Object, cluster string) {
		key, _ := layout.KeyForObject(mc.WithCluster(context.Background(), cluster, false), gr, obj)
		if _, err := cli.Delete(context.Background(), key); err != nil {
			panic(err)
		}
	}

	// Ensure clean slate
	cli.Delete(context.Background(), layout.KeyRoot(gr), clientv3.WithPrefix())
	time.Sleep(200 * time.Millisecond)

	// ADDEDs
	write(newObj("pod-1", "ns-1", "cluster-a"), "cluster-a")
	write(newObj("pod-2", "ns-2", "cluster-b"), "cluster-b")
	// UPDATE
	write(newObj("pod-1", "ns-1", "cluster-a"), "cluster-a")
	// DELETE
	del(newObj("pod-2", "ns-2", "cluster-b"), "cluster-b")

	// Collect a short window of events
	drain := func(s *subscriber) []string {
		out := []string{}
		timeout := time.After(800 * time.Millisecond)
		for {
			select {
			case ev := <-s.ch:
				out = append(out, fmt.Sprintf("%s %s rv=%d", ev.Type, ev.Key, ev.RV))
			case <-timeout:
				sort.Strings(out)
				return out
			}
		}
	}

	fmt.Println("watcher-a:")
	fmt.Println(strings.Join(drain(subA), "\n"))
	fmt.Println("watcher-b:")
	fmt.Println(strings.Join(drain(subB), "\n"))
	fmt.Println("watcher-all:")
	fmt.Println(strings.Join(drain(subAll), "\n"))
}
