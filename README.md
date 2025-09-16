# k8s-mcstore

Kubernetes Multi-Cluster Store (mcstore)

mcstore lets a kube-apiserver serve many logical clusters/projects with a single etcd List+Watch per resource kind. Objects are stored under a cluster segment after kind, and the watch cache fans out to clients with per-cluster filtering.

## Why

- One etcd watch per kind across all clusters → lowest etcd load
- Uniform, configurable key layout → simple migrations and backups
- Pluggable extractors, predicates, and watch strategies → operator flexibility

## Key layout

```
/<prefix>/<group>/<resource>/<cluster>/<namespace?>/<name>

examples:
/registry/core/pods/cluster-a/ns-1/pod-1
/registry/core/nodes/cluster-b/node-7
/registry/apiextensions.k8s.io/customresourcedefinitions/cluster-a/crd.example.io
```

Optionally reserve a global segment (default `__global__`):
```
/registry/<group>/<resource>/__global__/<name>
```

## How it works

```
Clients (per-project)
    │ list/watch
    ▼
API Server (genericregistry.Store)
    KeyRootFunc  → "/<prefix>/<group>/<resource>/" (shared across clusters)
    KeyFunc      → "/<prefix>/<group>/<resource>/<cluster>/<ns?>/<name>"
    Predicate    → WithClusterPredicate(...)
    ▼
Watch Cache (indexes: byCluster, byClusterNamespace)
    │ single etcd List+Watch per kind
    ▼
etcd (recursive prefix watch)
```

## Configuration (Options)

Set globally via `mcstore.DefaultOptions` or pass an `Options` on `DefaultKeyLayout`:

- `ClusterAnnotationKey` (default `multicluster.k8s.io/cluster`): where objects carry their cluster
- `ClusterFieldKey` (default `metadata.cluster`): field selector key used in predicates
- `GlobalSegment` (default `__global__`): reserved cluster name for global objects
- `Predicate` (default `PredicateHybrid`): one of `PredicatePathOnly`, `PredicateMetadataOnly`, `PredicateHybrid`
- `WatchStrategy` (default `KindRootStrategy`): single kind-root prefix vs `PerClusterStrategy`
- `EtcdPrefix` (default `/registry`): base etcd key prefix to match your apiserver config

Minimal setup:
```go
mc.DefaultOptions = &mc.Options{
    ClusterAnnotationKey: "multicluster.k8s.io/cluster",
    ClusterFieldKey:      "metadata.cluster",
    GlobalSegment:        "__global__",
    Predicate:            mc.PredicateHybrid,
    WatchStrategy:        mc.KindRootStrategy{},
    EtcdPrefix:           "/registry",
}
```

## Extracting cluster/project

Pick how the request’s cluster/project is determined:

- `ContextExtractor`: use `mcstore.WithCluster(ctx, id, all)` in your router/middleware
- `PathExtractor`: derive from URL using `mcstore.WithRequestPath(ctx, path)` and `mcstore.NewPathExtractor("/clusters/", "cluster")`
- Implement your own `ClusterExtractor` (header, auth principal, etc.)

Normalize “no project” to a sentinel (e.g., `root`) so every key has a cluster segment.

## Predicate modes

- `PredicatePathOnly`: do not inject a field; rely on path/prefix (pair with per-cluster prefixes if you need strict path-only isolation)
- `PredicateMetadataOnly`: inject `<ClusterFieldKey>=<cluster>`; single shared kind watch; requires objects to be annotated
- `PredicateHybrid`: same as MetadataOnly (default), compatible with any path-scoped flows; best general choice

## Watch strategies

- `KindRootStrategy` (default): one etcd watch per kind across all clusters
- `PerClusterStrategy`: one etcd watch per (cluster, kind), useful for isolating noisy tenants

## Integrating with kube-apiserver

```go
extractor := mcstore.NewPathExtractor("/clusters/", "cluster") // or your own
layout    := &mcstore.DefaultKeyLayout{Extractor: extractor, Options: mcstore.DefaultOptions}
gr        := schema.GroupResource{Group: group, Resource: resource}

store := &genericregistry.Store{
    NewFunc:     newFunc,
    NewListFunc: newListFunc,
    KeyRootFunc: func(ctx context.Context) string { return layout.KeyRoot(gr) },
    KeyFunc: func(ctx context.Context, name string) (string, error) {
        obj := newFunc()
        // set name/namespace on obj from ctx as needed
        return layout.KeyForObject(ctx, gr, obj)
    },
    PredicateFunc: func(lbl labels.Selector, fld fields.Selector) storage.SelectionPredicate {
        base := genericregistry.MatchFunc(lbl, fld)
        cid, all, _ := extractor.ExtractCluster(requestContext)
        return mcstore.WithClusterPredicate(base, cid, all)
    },
    // Storage: your etcd3 storage (optionally wrapped by a cacher built with your WatchStrategy)
}
```

To add a watch cache, build a `ListerWatcher` using your chosen strategy (default: kind root) and index the cache by cluster and cluster/namespace.

## Examples

- Client: `go run ./examples/client` — roots, per-cluster prefixes, predicate composition
- Apiserver skeleton: `go run ./examples/apiserver` — wiring and outputs
- Etcd demo: `docker compose -f ./examples/etcd-demo/compose.yaml up -d && ETCD_ENDPOINTS=127.0.0.1:2379 go run ./examples/etcd-demo` — single prefix watch across clusters
- Watch strategies: `go run ./examples/strategy` — kind-root vs per-cluster prefixes
- Prefixes: `go run ./examples/prefixes` — built-ins, CRDs, and global segment
- Extractor: `go run ./examples/extractor` — path-based extraction
- E2E watch cache: `docker compose -f ./examples/etcd-demo/compose.yaml up -d && ETCD_ENDPOINTS=127.0.0.1:2379 go run ./examples/e2e-watchcache` — one watch, filtered fan-out

## Operations

- ResourceVersion: use etcd’s global revisions; ordering is total and monotonic across clusters
- HA: each apiserver replica keeps its own per-kind watch; no cross-replica coordination
- Sizing: watch-cache per kind sized for total objects across clusters
- Sharding: for hot kinds, shard etcd by key range; your layout stays the same

## Security & tenancy

- Enforce tenancy at REST using the extracted cluster/project
- Optional etcd ACLs per `/<prefix>/<group>/<resource>/<cluster>/` for defense in depth
- Path-only mode (PredicatePathOnly) works without metadata

## Migration

- Stand up a new keyspace and rewrite `/registry/...` keys to the new layout, or shadow-write then flip reads

## FAQ

- Do I need etcd changes? No — recursive prefix watches and revisions are already available
- Can I rename “cluster” to project/tenant? Yes — configure `Options` and your extractor
- What about global objects? Use `GlobalSegment` to store them under the kind root
- Custom etcd prefix? Set `Options.EtcdPrefix` to match your apiserver


