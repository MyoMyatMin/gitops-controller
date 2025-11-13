package k8s

import (
	"context"
	"fmt"

	"github.com/MyoMyatMin/gitops-controller/internal/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (c *Client) ListManagedResources(namespace string) ([]unstructured.Unstructured, error) {
	var managedResources []unstructured.Unstructured

	gvrs := []schema.GroupVersionResource{
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "configmaps"},
		{Group: "", Version: "v1", Resource: "secrets"},
		{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
	}

	labelSelector := fmt.Sprintf("%s=%s", ManagedByLabel, FieldManager)

	for _, gvr := range gvrs {
		list, err := c.dynamic.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})

		if err != nil {
			log.Warnf("Could not list %s: %v", gvr.Resource, err)
			continue

		}

		managedResources = append(managedResources, list.Items...)
	}

	log.Infof("Found %d managed resources in namespace %s", len(managedResources), namespace)
	return managedResources, nil
}
