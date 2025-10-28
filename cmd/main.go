package main

import (
	"fmt"
	"os"

	"github.com/MyoMyatMin/gitops-controller/internal/k8s"
)

func main() {
	fmt.Println("GitOps Controller Starting")
	k8sClient, err := k8s.NewClient("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	err = k8sClient.ListNamespaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing namespaces: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
