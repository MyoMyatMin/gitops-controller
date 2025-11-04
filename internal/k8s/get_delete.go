package k8s

import (
	"context"
	"fmt"

	"github.com/MyoMyatMin/gitops-controller/internal/sync"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func (c *Client) getResourceInterface(manifest sync.Manifest) (dynamic.ResourceInterface, error) {
	obj := manifest.Object
	gvk := obj.GroupVersionKind()

	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST mapping for %s: %v", gvk.String(), err)
	}

	var resourceInterface dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		resourceInterface = c.dynamic.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		resourceInterface = c.dynamic.Resource(mapping.Resource)
	}
	return resourceInterface, nil
}

func (c *Client) Get(manifest sync.Manifest) (*unstructured.Unstructured, error) {
	resourceInterface, err := c.getResourceInterface(manifest)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Getting: Kind=%s, Name=%s, Namespace=%s\n",
		manifest.Kind, manifest.Name, manifest.Namespace)

	return resourceInterface.Get(context.TODO(), manifest.Name, metav1.GetOptions{})
}

func (c *Client) Delete(manifest sync.Manifest) error {
	resourceInterface, err := c.getResourceInterface(manifest)
	if err != nil {
		return err
	}

	fmt.Printf("Deleting: Kind=%s, Name=%s, Namespace=%s\n",
		manifest.Kind, manifest.Name, manifest.Namespace)
	return resourceInterface.Delete(context.TODO(), manifest.Name, metav1.DeleteOptions{})
}
