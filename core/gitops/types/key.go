package types

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ObjectKey struct {
	Name      string
	Namespace string
}

func NewObjectKey(meta metav1.ObjectMeta) ObjectKey {
	return ObjectKey{
		Name:      meta.Name,
		Namespace: meta.Namespace,
	}
}
