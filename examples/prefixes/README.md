Prefixes for built-ins, CRDs, and global

Shows list prefixes for different GroupResources and the global segment.

Run:

```bash
go run ./examples/prefixes
```

Expected output (example):

```text
pods ns-1: /registry/core/pods/cluster-a/ns-1/
widgets all-ns: /registry/acme.io/widgets/cluster-a/
global pods: /registry/core/pods/__global__/
global widgets: /registry/acme.io/widgets/__global__/
```


