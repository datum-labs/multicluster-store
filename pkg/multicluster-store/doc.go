package mcstore

// Package mcstore provides a multi-cluster-aware storage wrapper for kube-apiserver.
//
// It implements a key layout of the form:
//   /registry/<group>/<resource>/<clusterID>/<namespace?>/<name>
// and builds a single per-kind etcd List+Watch on /registry/<group>/<resource>/,
// while fan-out to clients is filtered by cluster and namespace using indexes.
