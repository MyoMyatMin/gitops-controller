package manifest

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type Manifest struct {
	FilePath  string
	Kind      string
	Name      string
	Namespace string
	Object    *unstructured.Unstructured
}
