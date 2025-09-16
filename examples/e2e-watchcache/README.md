E2E watch-cache fan-out (single kind watch)

Demonstrates one etcd watch per kind across all clusters, with in-memory fan-out filtered by cluster.

Prereq: etcd via compose

```bash
docker compose -f ./examples/etcd-demo/compose.yaml up -d
```

Run:

```bash
ETCD_ENDPOINTS=127.0.0.1:2379 go run ./examples/e2e-watchcache
```

Output shows per-subscriber event streams (cluster-a, cluster-b, all), all fed from a single etcd prefix watch.


