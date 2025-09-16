Apiserver skeleton

Demonstrates wiring `DefaultKeyLayout`, `WithCluster`, and `WithClusterPredicate`.

Run:

```bash
go run ./examples/apiserver
```

Expected output (example):

```text
KeyRoot: /registry/core/pods/
Cluster list prefix: /registry/core/pods/cluster-a/default/
Predicate field selector: metadata.cluster=cluster-a
Global list prefix: /registry/core/pods/__global__/
```

To build a container:

```bash
docker build -t mc-apiserver:dev ./examples/apiserver
```


