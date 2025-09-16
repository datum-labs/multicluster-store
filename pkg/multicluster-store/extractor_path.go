package mcstore

import (
	"context"
	"strings"
)

type ctxKey int

const (
	ctxKeyRequestPath ctxKey = iota
)

// WithRequestPath stores an incoming request path in context for path-based extraction.
func WithRequestPath(parent context.Context, path string) context.Context {
	return context.WithValue(parent, ctxKeyRequestPath, path)
}

// PathExtractor extracts a cluster id from a request path based on a template like "/clusters/{cluster}/apis/...".
type PathExtractor struct{ Prefix, Param string }

// NewPathExtractor returns an extractor that reads cluster id from the request path.
// Example: NewPathExtractor("/clusters/", "cluster") expects /clusters/<value>/...
func NewPathExtractor(prefix string, param string) PathExtractor {
	return PathExtractor{Prefix: prefix, Param: param}
}

func (p PathExtractor) ExtractCluster(ctx context.Context) (string, bool, error) {
	rp, _ := ctx.Value(ctxKeyRequestPath).(string)
	if rp == "" {
		// no path provided; treat as all-clusters if omitted
		return "", true, nil
	}
	// naive split-based extraction: find segment after prefix
	// e.g., prefix=/clusters/ => /clusters/<cluster>/...
	if !strings.HasPrefix(rp, p.Prefix) {
		return "", true, nil
	}
	rest := strings.TrimPrefix(rp, p.Prefix)
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", true, nil
	}
	return parts[0], false, nil
}
