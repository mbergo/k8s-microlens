package common

import (
	"k8s.io/client-go/kubernetes"
)

// KubernetesClient interface defines the methods we use from the clientset
type KubernetesClient interface {
	kubernetes.Interface
}
