package integration

import (
	"context"
	"os"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/mbergo/k8s-microlens/internal/common"
)

func setupIntegrationTest(t *testing.T) *common.ResourceProcessor {
	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test")
	}

	// Get kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Error getting home directory: %v", err)
		}
		kubeconfig = homeDir + "/.kube/config"
	}

	// Build config
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatalf("Error building kubeconfig: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("Error creating kubernetes client: %v", err)
	}

	return common.NewResourceProcessor(clientset, context.Background())
}

func TestIntegration(t *testing.T) {
	processor := setupIntegrationTest(t)

	t.Run("ProcessDefaultNamespace", func(t *testing.T) {
		err := processor.ProcessNamespace("default")
		if err != nil {
			t.Errorf("Error processing default namespace: %v", err)
		}
	})

	t.Run("ShowKubeSystemResourceRelationships", func(t *testing.T) {
		err := processor.ShowResourceRelationships("kube-system")
		if err != nil {
			t.Errorf("Error showing kube-system relationships: %v", err)
		}
	})
}
