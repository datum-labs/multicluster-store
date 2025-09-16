package mcstore

import (
	"context"
)

// ClusterExtractor returns the logical cluster ID for the request.
// If all==true, the caller is requesting cross-cluster scope.
type ClusterExtractor interface {
	ExtractCluster(ctx context.Context) (clusterID string, all bool, err error)
}

type contextKey int

const (
	// ctxKeyClusterID carries the logical cluster identifier in the request context.
	ctxKeyClusterID contextKey = iota
	// ctxKeyAllClusters indicates the request targets all clusters.
	ctxKeyAllClusters
)

// WithCluster returns a child context that carries cluster selection.
func WithCluster(parent context.Context, clusterID string, all bool) context.Context {
	ctx := context.WithValue(parent, ctxKeyClusterID, clusterID)
	ctx = context.WithValue(ctx, ctxKeyAllClusters, all)
	return ctx
}

// ContextExtractor reads cluster selection from context values set by WithCluster.
type ContextExtractor struct{}

func (ContextExtractor) ExtractCluster(ctx context.Context) (string, bool, error) {
	cid, _ := ctx.Value(ctxKeyClusterID).(string)
	all, _ := ctx.Value(ctxKeyAllClusters).(bool)
	return cid, all, nil
}
