package mcstore

// PredicateMode controls how cluster scoping is enforced in predicates.
type PredicateMode int

const (
	// PredicatePathOnly enforces scoping via key layout only; no field selector is injected.
	PredicatePathOnly PredicateMode = iota
	// PredicateMetadataOnly injects the cluster field selector; ignores path-based concerns.
	PredicateMetadataOnly
	// PredicateHybrid allows both path isolation and metadata field selectors.
	PredicateHybrid
)

// Options controls naming and behavior for cluster scoping, indexing, and globals.
type Options struct {
	// ClusterAnnotationKey is read from object metadata when present (default: "multicluster.k8s.io/cluster").
	ClusterAnnotationKey string
	// ClusterFieldKey is used when composing field selectors (default: "metadata.cluster").
	ClusterFieldKey string
	// GlobalSegment names the global bucket under a kind (default: "__global__").
	GlobalSegment string
	// Predicate selects how predicates are composed (default: PredicateHybrid).
	Predicate PredicateMode
	// WatchStrategy controls how etcd list/watch prefixes are chosen.
	// Default is a single-kind-root strategy.
	WatchStrategy WatchStrategy
	// EtcdPrefix is the base etcd key prefix (default: "/registry").
	EtcdPrefix string
}

// DefaultOptions holds the package defaults. Mutate to change global behavior.
var DefaultOptions = &Options{
	ClusterAnnotationKey: "multicluster.k8s.io/cluster",
	ClusterFieldKey:      "metadata.cluster",
	GlobalSegment:        "__global__",
	Predicate:            PredicateHybrid,
	WatchStrategy:        KindRootStrategy{},
	EtcdPrefix:           "/registry",
}

func effectiveOptions(opts *Options) *Options {
	if opts == nil {
		return DefaultOptions
	}
	// Fill zero values from defaults
	if opts.ClusterAnnotationKey == "" {
		opts.ClusterAnnotationKey = DefaultOptions.ClusterAnnotationKey
	}
	if opts.ClusterFieldKey == "" {
		opts.ClusterFieldKey = DefaultOptions.ClusterFieldKey
	}
	if opts.GlobalSegment == "" {
		opts.GlobalSegment = DefaultOptions.GlobalSegment
	}
	if opts.Predicate != PredicatePathOnly && opts.Predicate != PredicateMetadataOnly && opts.Predicate != PredicateHybrid {
		opts.Predicate = DefaultOptions.Predicate
	}
	if opts.WatchStrategy == nil {
		opts.WatchStrategy = DefaultOptions.WatchStrategy
	}
	if opts.EtcdPrefix == "" {
		opts.EtcdPrefix = DefaultOptions.EtcdPrefix
	}
	return opts
}
