package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ListManagedResources finds all resources in a namespace managed by us.
func (c *Client) ListManagedResources(namespace string) ([]unstructured.Unstructured, error) {
	var managedResources []unstructured.Unstructured

	// Define the common resource types we want to check for pruning
	gvrs := []schema.GroupVersionResource{
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "configmaps"},
		{Group: "", Version: "v1", Resource: "secrets"},
		{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
	}

	// --- THIS IS THE FIX ---
	// Create the label selector using our shared constants
	labelSelector := fmt.Sprintf("%s=%s", ManagedByLabel, FieldManager)
	// --- END FIX ---

	for _, gvr := range gvrs {
		list, err := c.dynamic.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector, // Use the correct selector
		})

		if err != nil {
			fmt.Printf("Warning: could not list %s: %v\n", gvr.Resource, err)
			continue
		}

		managedResources = append(managedResources, list.Items...)
	}

	fmt.Printf("Found %d managed resources in namespace %s\n", len(managedResources), namespace)
	return managedResources, nil
}
