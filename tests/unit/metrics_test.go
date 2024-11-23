package unit

import (
	"context"
	"testing"

	"github.com/mbergo/k8s-microlens/internal/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestResourceMetrics(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()
	formatter := common.NewFormatter()
	metrics := common.NewResourceMetrics(clientset, formatter)

	t.Run("ShowResourceUtilization", func(t *testing.T) {
		err := metrics.ShowResourceUtilization("default")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("ShowNodeMetrics", func(t *testing.T) {
		// Create a test node
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
			Status: corev1.NodeStatus{
				Capacity: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("4"),
					corev1.ResourceMemory: resource.MustParse("8Gi"),
					corev1.ResourcePods:   resource.MustParse("110"),
				},
				Allocatable: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("3.8"),
					corev1.ResourceMemory: resource.MustParse("7.5Gi"),
					corev1.ResourcePods:   resource.MustParse("100"),
				},
			},
		}
		_, err := clientset.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Error creating test node: %v", err)
		}

		err = metrics.ShowNodeMetrics()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
