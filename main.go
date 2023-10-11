package main

import (
	"context"
	"encoding/json"
	"os"

	eksauth "github.com/chankh/eksutil/pkg/auth"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
)

type Payload struct {
	ClusterName string   `json:"clusterName"`
	Namespaces  []string `json:"namespaces"`
	Replicas    int32    `json:"replicas"`
}

func main() {
	if os.Getenv("ENV") == "DEBUG" {
		log.SetLevel(log.DebugLevel)
	}

	lambda.Start(handler)
}

func handler(ctx context.Context, payload Payload) (string, error) {
	cfg := &eksauth.ClusterConfig{
		ClusterName: payload.ClusterName,
	}

	clientset, err := eksauth.NewAuthClient(cfg)
	if err != nil {
		log.WithError(err).Error("Failed to create EKS client")
		return "", err
	}

	scaled := make(map[string]int32)

	for _, ns := range payload.Namespaces {
		deployments, err := clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.WithError(err).Errorf("Failed to list deployments in namespace %s", ns)
			continue
		}

		for _, deploy := range deployments.Items {
			if err := scaleDeploy(clientset, ctx, ns, deploy.Name, payload.Replicas); err == nil {
				scaled[ns+"/"+deploy.Name] = payload.Replicas
			}
		}
	}

	scaledJSON, err := json.Marshal(scaled)
	if err != nil {
		log.WithError(err).Error("Failed to marshal scaled deployments to JSON")
		return "", err
	}

	log.Info("Scaled Deployments: ", string(scaledJSON))
	return "Scaled Deployments: " + string(scaledJSON), nil
}

func scaleDeploy(client *kubernetes.Clientset, ctx context.Context, namespace, name string, replicas int32) error {
	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: replicas,
		},
	}

	_, err := client.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	if err != nil {
		log.WithError(err).Errorf("Failed to scale deployment %s in namespace %s", name, namespace)
	} else {
		log.Infof("Successfully scaled deployment %s in namespace %s to %d replicas", name, namespace, replicas)
	}
	return err
}
