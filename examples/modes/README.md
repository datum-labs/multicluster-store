Predicate modes

Demonstrates PathOnly, MetadataOnly, and Hybrid predicate composition.

Run:

```bash
go run ./examples/modes
```

Expected output (example):

```text
mode=0 prefix=/registry/core/pods/cluster-a/default/ field=
mode=1 prefix=/registry/core/pods/cluster-a/default/ field==cluster-a
mode=2 prefix=/registry/core/pods/cluster-a/default/ field==cluster-a
```


