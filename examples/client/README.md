Client example

Demonstrates key roots/prefixes and predicate composition.

Run:

```bash
go run ./examples/client
```

Expected output (example):

```text
List/Watch root for kind (all clusters): /registry/core/pods/
List prefix cluster-a ns-1: /registry/core/pods/cluster-a/ns-1/
List prefix cluster-b all-namespaces: /registry/core/pods/cluster-b/
List prefix global: /registry/core/pods/__global__/
All clusters predicate field empty? true
Cluster-a predicate field: metadata.cluster=cluster-a
```


