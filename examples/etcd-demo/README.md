Etcd-backed demo

Shows a single prefix watch per kind across clusters using etcd.

Start etcd:

```bash
docker compose -f ./examples/etcd-demo/compose.yaml up -d
```

Run demo (customizes Options to use "project"):

```bash
ETCD_ENDPOINTS=127.0.0.1:2379 go run ./examples/etcd-demo
```

Expected output (example):

```text
Watching prefix: /registry/core/pods/
event=PUT cluster=project-a key=/registry/core/pods/project-a/ns-1/pod-1
event=PUT cluster=project-b key=/registry/core/pods/project-b/ns-2/pod-2
event=PUT cluster=project-a key=/registry/core/pods/project-a/ns-1/pod-1
event=DELETE cluster=project-b key=/registry/core/pods/project-b/ns-2/pod-2
```

Tear down etcd:

```bash
docker compose -f ./examples/etcd-demo/compose.yaml down
```


