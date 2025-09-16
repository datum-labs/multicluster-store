package mcstore

import (
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apiserver/pkg/storage"
)

// WithClusterPredicate wraps the base SelectionPredicate to constrain by cluster when all==false.
func WithClusterPredicate(base storage.SelectionPredicate, clusterID string, all bool) storage.SelectionPredicate {
	if all {
		return base
	}
	opts := effectiveOptions(DefaultOptions)
	if opts.Predicate == PredicatePathOnly {
		return base
	}
	fieldKey := opts.ClusterFieldKey
	var newField fields.Selector
	if base.Field == nil || base.Field.Empty() {
		newField = fields.OneTermEqualSelector(fieldKey, clusterID)
	} else {
		newField = fields.AndSelectors(base.Field, fields.OneTermEqualSelector(fieldKey, clusterID))
	}
	base.Field = newField
	return base
}

// EmptyPredicate returns a zero-value SelectionPredicate.
func EmptyPredicate() storage.SelectionPredicate { return storage.SelectionPredicate{} }
