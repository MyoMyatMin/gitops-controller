package k8s

import (
	"context"
	"fmt"

	"github.com/MyoMyatMin/gitops-controller/internal/log"
	"github.com/MyoMyatMin/gitops-controller/pkg/manifest"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func (c *Client) getResourceInterface(manifest manifest.Manifest) (dynamic.ResourceInterface, error) {
	obj := manifest.Object
	gvk := obj.GroupVersionKind()

	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		log.Errorf("error getting REST mapping for %s: %v", gvk, err)
		return nil, fmt.Errorf("error getting REST mapping for %s: %w", gvk, err)
	}

	var resourceInterface dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		resourceInterface = c.dynamic.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		resourceInterface = c.dynamic.Resource(mapping.Resource)
	}

	return resourceInterface, nil
}

func (c *Client) Get(manifest manifest.Manifest) (*unstructured.Unstructured, error) {
	resourceInterface, err := c.getResourceInterface(manifest)
	if err != nil {
		return nil, err
	}

	logFields := logrus.Fields{
		"kind":      manifest.Kind,
		"name":      manifest.Name,
		"namespace": manifest.Namespace,
	}
	log.WithFields(logFields).Info("Getting resource")

	return resourceInterface.Get(context.TODO(), manifest.Name, metav1.GetOptions{})
}

func (c *Client) Delete(manifest manifest.Manifest) error {
	resourceInterface, err := c.getResourceInterface(manifest)
	if err != nil {
		return err
	}

	logFields := logrus.Fields{
		"kind":      manifest.Kind,
		"name":      manifest.Name,
		"namespace": manifest.Namespace,
	}
	log.WithFields(logFields).Info("Deleting resource")

	return resourceInterface.Delete(context.TODO(), manifest.Name, metav1.DeleteOptions{})
}
