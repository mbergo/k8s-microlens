package common

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type ResourceMetrics struct {
	clientset *kubernetes.Clientset
	formatter *Formatter
}

func NewResourceMetrics(clientset *kubernetes.Clientset, formatter *Formatter) *ResourceMetrics {
	return &ResourceMetrics{
		clientset: clientset,
		formatter: formatter,
	}
}

// formatCPU converts CPU cores to a human-readable format
func (rm *ResourceMetrics) formatCPU(cpu int64) string {
	if cpu < 1000 {
		return fmt.Sprintf("%dm", cpu)
	}
	return fmt.Sprintf("%.2f", float64(cpu)/1000)
}

// formatMemory converts memory bytes to a human-readable format
func (rm *ResourceMetrics) formatMemory(bytes int64) string {
	sizes := []string{"B", "Ki", "Mi", "Gi", "Ti"}
	if bytes == 0 {
		return "0B"
	}
	i := math.Floor(math.Log(float64(bytes)) / math.Log(1024))
	if i >= float64(len(sizes)) {
		i = float64(len(sizes)) - 1
	}
	value := float64(bytes) / math.Pow(1024, i)
	return fmt.Sprintf("%.2f%s", value, sizes[int(i)])
}

// GetNodeMetrics fetches metrics for all nodes
func (rm *ResourceMetrics) GetNodeMetrics() (*metricsv1beta1.NodeMetricsList, error) {
	raw, err := rm.clientset.RESTClient().Get().
		Resource("nodes").
		SubResource("proxy").
		Name("metrics").
		DoRaw(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting node metrics: %v", err)
	}

	metrics := &metricsv1beta1.NodeMetricsList{}
	if err := json.Unmarshal(raw, metrics); err != nil {
		return nil, fmt.Errorf("error parsing node metrics: %v", err)
	}

	return metrics, nil
}

// ShowNodeMetrics displays metrics for all nodes
func (rm *ResourceMetrics) ShowNodeMetrics() error {
	fmt.Println("\n[Node Metrics]")

	metrics, err := rm.GetNodeMetrics()
	if err != nil {
		return err
	}

	for _, node := range metrics.Items {
		rm.formatter.PrintResource("├──", "Node", node.Name)
		rm.formatter.Indent()

		nodeInfo, err := rm.clientset.CoreV1().Nodes().Get(context.Background(), node.Name, metav1.GetOptions{})
		if err != nil {
			rm.formatter.PrintInfo("", "Error getting node info: %v", err)
			continue
		}

		capacity := nodeInfo.Status.Capacity
		allocatable := nodeInfo.Status.Allocatable

		rm.formatter.PrintInfo("", "Capacity:")
		rm.formatter.PrintInfo("", "  CPU: %s", capacity.Cpu().String())
		rm.formatter.PrintInfo("", "  Memory: %s", rm.formatMemory(capacity.Memory().Value()))
		rm.formatter.PrintInfo("", "  Pods: %s", capacity.Pods().String())

		rm.formatter.PrintInfo("", "Allocatable:")
		rm.formatter.PrintInfo("", "  CPU: %s", allocatable.Cpu().String())
		rm.formatter.PrintInfo("", "  Memory: %s", rm.formatMemory(allocatable.Memory().Value()))
		rm.formatter.PrintInfo("", "  Pods: %s", allocatable.Pods().String())

		pods, err := rm.clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node.Name,
		})
		if err != nil {
			rm.formatter.PrintInfo("", "Error getting pod list: %v", err)
			continue
		}

		rm.formatter.PrintInfo("", "Current State:")
		rm.formatter.PrintInfo("", "  Running Pods: %d", len(pods.Items))

		var totalCPURequests, totalMemoryRequests int64
		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				if container.Resources.Requests != nil {
					totalCPURequests += container.Resources.Requests.Cpu().MilliValue()
					totalMemoryRequests += container.Resources.Requests.Memory().Value()
				}
			}
		}

		cpuPercentage := float64(totalCPURequests) / float64(allocatable.Cpu().MilliValue()) * 100
		memoryPercentage := float64(totalMemoryRequests) / float64(allocatable.Memory().Value()) * 100

		rm.formatter.PrintInfo("", "  CPU Usage: %.2f%% (%s/%s)",
			cpuPercentage,
			rm.formatCPU(totalCPURequests),
			allocatable.Cpu().String())
		rm.formatter.PrintInfo("", "  Memory Usage: %.2f%% (%s/%s)",
			memoryPercentage,
			rm.formatMemory(totalMemoryRequests),
			rm.formatMemory(allocatable.Memory().Value()))

		rm.formatter.Outdent()
	}

	return nil
}

// ShowResourceUtilization shows resource utilization for pods in a namespace
func (rm *ResourceMetrics) ShowResourceUtilization(namespace string) error {
	fmt.Printf("\n[Resource Utilization: %s]\n", namespace)

	pods, err := rm.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting pods: %v", err)
	}

	var totalRequestCPU, totalRequestMemory int64
	var totalLimitCPU, totalLimitMemory int64

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if container.Resources.Requests != nil {
				totalRequestCPU += container.Resources.Requests.Cpu().MilliValue()
				totalRequestMemory += container.Resources.Requests.Memory().Value()
			}
			if container.Resources.Limits != nil {
				totalLimitCPU += container.Resources.Limits.Cpu().MilliValue()
				totalLimitMemory += container.Resources.Limits.Memory().Value()
			}
		}
	}

	rm.formatter.PrintInfo("", "Namespace Summary:")
	rm.formatter.PrintInfo("", "CPU:")
	rm.formatter.PrintInfo("", "  Requests: %s", rm.formatCPU(totalRequestCPU))
	rm.formatter.PrintInfo("", "  Limits: %s", rm.formatCPU(totalLimitCPU))

	rm.formatter.PrintInfo("", "Memory:")
	rm.formatter.PrintInfo("", "  Requests: %s", rm.formatMemory(totalRequestMemory))
	rm.formatter.PrintInfo("", "  Limits: %s", rm.formatMemory(totalLimitMemory))

	return nil
}
