package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func scaleDeploymentsToZero() (string, error) {
	ctx := context.TODO()

	// Initialize Kubernetes client from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", "/home/devops/.kube/config")
	if err != nil {
		return "Failed to load kubeconfig", err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "Failed to create kubernetes client", err
	}

	// List deployments in the "default" namespace
	deployments, err := clientset.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return "Failed to list deployments", err
	}

	// Scale all deployments to 0
	for _, deploy := range deployments.Items {
		deploy.Spec.Replicas = new(int32) // Pointer to int32
		*deploy.Spec.Replicas = 0         // Set to 0
		_, err := clientset.AppsV1().Deployments("default").Update(ctx, &deploy, metav1.UpdateOptions{})
		if err != nil {
			return "Failed to scale deployment", err
		}
	}

	return "Successfully scaled all deployments to 0", nil
}

func main() {
	result, err := scaleDeploymentsToZero()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(result)
}
