package mcstore

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Constants are kept for backward compatibility; prefer Options.
const (
	// AnnotationClusterID is used to carry cluster identity on objects where needed.
	AnnotationClusterID = "multicluster.k8s.io/cluster"
	// FieldClusterKey is the field selector key used to filter by cluster.
	FieldClusterKey = "metadata.cluster"
)

func getNamespace(obj runtime.Object) string {
	acc, err := meta.Accessor(obj)
	if err != nil {
		return ""
	}
	return acc.GetNamespace()
}

func getName(obj runtime.Object) string {
	acc, err := meta.Accessor(obj)
	if err != nil {
		return ""
	}
	return acc.GetName()
}

// getClusterID reads the cluster ID from metadata annotations.
func getClusterID(obj runtime.Object) string {
	acc, err := meta.Accessor(obj)
	if err != nil {
		return ""
	}
	anns := acc.GetAnnotations()
	if anns == nil {
		return ""
	}
	key := AnnotationClusterID
	if DefaultOptions != nil && DefaultOptions.ClusterAnnotationKey != "" {
		key = DefaultOptions.ClusterAnnotationKey
	}
	return anns[key]
}

// setClusterID mutates ObjectMeta to carry the cluster ID annotation.
func setClusterID(obj runtime.Object, clusterID string) {
	acc, err := meta.Accessor(obj)
	if err != nil {
		return
	}
	anns := acc.GetAnnotations()
	if anns == nil {
		anns = map[string]string{}
	}
	key := AnnotationClusterID
	if DefaultOptions != nil && DefaultOptions.ClusterAnnotationKey != "" {
		key = DefaultOptions.ClusterAnnotationKey
	}
	anns[key] = clusterID
	acc.SetAnnotations(anns)
}

// ObjectAttrsFunc returns a function compatible with cacher to extract attrs.
func ObjectAttrsFunc(example func() runtime.Object) func(obj runtime.Object) (labels map[string]string, fields map[string]string, err error) {
	return func(obj runtime.Object) (map[string]string, map[string]string, error) {
		acc, e := meta.Accessor(obj)
		if e != nil {
			return nil, nil, e
		}
		fieldsMap := map[string]string{
			"metadata.name":      acc.GetName(),
			"metadata.namespace": acc.GetNamespace(),
		}
		if cid := getClusterID(obj); cid != "" {
			// Only include cluster field when present, so selection can be optional.
			key := FieldClusterKey
			if DefaultOptions != nil && DefaultOptions.ClusterFieldKey != "" {
				key = DefaultOptions.ClusterFieldKey
			}
			fieldsMap[key] = cid
		}
		return acc.GetLabels(), fieldsMap, nil
	}
}

// Ensure metav1 imported for potential future use to avoid unused import churn.
var _ = metav1.ObjectMeta{}
