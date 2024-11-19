package common

import (
	"context"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ResourceProcessor struct {
	clientset *kubernetes.Clientset
	ctx       context.Context
	formatter *Formatter
}

func NewResourceProcessor(clientset *kubernetes.Clientset, ctx context.Context) *ResourceProcessor {
	return &ResourceProcessor{
		clientset: clientset,
		ctx:       ctx,
		formatter: NewFormatter(),
	}
}

func (rp *ResourceProcessor) ShowDeploymentDetails(namespace string) error {
	fmt.Println("\n[Deployment Layer]")
	deployments, err := rp.clientset.AppsV1().Deployments(namespace).List(rp.ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting deployments: %v", err)
	}

	for i, deploy := range deployments.Items {
		isLast := i == len(deployments.Items)-1
		prefix := "├──"
		if isLast {
			prefix = "└──"
		}

		rp.formatter.PrintResource(prefix, "Deployment", deploy.Name)
		rp.formatter.Indent()

		rp.formatter.PrintInfo("", "Replicas: %d/%d", deploy.Status.AvailableReplicas, *deploy.Spec.Replicas)
		rp.formatter.PrintInfo("", "Strategy: %s", deploy.Spec.Strategy.Type)

		if deploy.Spec.Strategy.RollingUpdate != nil {
			rp.formatter.PrintInfo("", "Max Surge: %s", deploy.Spec.Strategy.RollingUpdate.MaxSurge.String())
			rp.formatter.PrintInfo("", "Max Unavailable: %s", deploy.Spec.Strategy.RollingUpdate.MaxUnavailable.String())
		}

		// Show container details
		for _, container := range deploy.Spec.Template.Spec.Containers {
			rp.formatter.PrintInfo("", "Container: %s (Image: %s)", container.Name, container.Image)
			if len(container.Ports) > 0 {
				for _, port := range container.Ports {
					rp.formatter.PrintInfo("", "  Port: %d/%s", port.ContainerPort, port.Protocol)
				}
			}

			if container.Resources.Limits != nil || container.Resources.Requests != nil {
				rp.formatter.PrintInfo("", "  Resources:")
				if container.Resources.Limits != nil {
					cpu := container.Resources.Limits.Cpu()
					memory := container.Resources.Limits.Memory()
					rp.formatter.PrintInfo("", "    Limits: CPU: %v, Memory: %v", cpu, memory)
				}
				if container.Resources.Requests != nil {
					cpu := container.Resources.Requests.Cpu()
					memory := container.Resources.Requests.Memory()
					rp.formatter.PrintInfo("", "    Requests: CPU: %v, Memory: %v", cpu, memory)
				}
			}
		}

		rp.formatter.Outdent()
	}

	return nil
}

func (rp *ResourceProcessor) ShowHPADetails(namespace string) error {
	fmt.Println("\n[HPA Layer]")
	hpas, err := rp.clientset.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(rp.ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting HPAs: %v", err)
	}

	for i, hpa := range hpas.Items {
		isLast := i == len(hpas.Items)-1
		prefix := "├──"
		if isLast {
			prefix = "└──"
		}

		rp.formatter.PrintResource(prefix, "HPA", hpa.Name)
		rp.formatter.Indent()

		rp.formatter.PrintInfo("", "Target: %s/%s",
			hpa.Spec.ScaleTargetRef.Kind,
			hpa.Spec.ScaleTargetRef.Name)
		rp.formatter.PrintInfo("", "Min Replicas: %d", *hpa.Spec.MinReplicas)
		rp.formatter.PrintInfo("", "Max Replicas: %d", hpa.Spec.MaxReplicas)

		for _, metric := range hpa.Spec.Metrics {
			switch metric.Type {
			case autoscalingv2.ResourceMetricSourceType:
				if metric.Resource != nil {
					rp.formatter.PrintInfo("", "Resource Metric: %s", metric.Resource.Name)
					if metric.Resource.Target.AverageUtilization != nil {
						rp.formatter.PrintInfo("", "  Target Utilization: %d%%",
							*metric.Resource.Target.AverageUtilization)
					}
					if metric.Resource.Target.AverageValue != nil {
						rp.formatter.PrintInfo("", "  Target Value: %s",
							metric.Resource.Target.AverageValue.String())
					}
				}
			case autoscalingv2.PodsMetricSourceType:
				if metric.Pods != nil {
					rp.formatter.PrintInfo("", "Pods Metric: %s", metric.Pods.Metric.Name)
					rp.formatter.PrintInfo("", "  Target Average Value: %s",
						metric.Pods.Target.AverageValue.String())
				}
			}
		}

		if hpa.Status.CurrentReplicas > 0 {
			rp.formatter.PrintInfo("", "Current Replicas: %d", hpa.Status.CurrentReplicas)
			rp.formatter.PrintInfo("", "Desired Replicas: %d", hpa.Status.DesiredReplicas)
		}

		rp.formatter.Outdent()
	}

	return nil
}

func (rp *ResourceProcessor) ShowConfigMapUsage(namespace string) error {
	fmt.Println("\n[ConfigMap Layer]")
	configMaps, err := rp.clientset.CoreV1().ConfigMaps(namespace).List(rp.ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting configmaps: %v", err)
	}

	for i, cm := range configMaps.Items {
		isLast := i == len(configMaps.Items)-1
		prefix := "├──"
		if isLast {
			prefix = "└──"
		}

		rp.formatter.PrintResource(prefix, "ConfigMap", cm.Name)
		rp.formatter.Indent()

		rp.formatter.PrintInfo("", "Data Keys: %d", len(cm.Data))

		pods, err := rp.clientset.CoreV1().Pods(namespace).List(rp.ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error getting pods: %v", err)
		}

		found := false
		for _, pod := range pods.Items {
			usedAs := rp.getConfigMapUsageInPod(&pod, cm.Name)
			if len(usedAs) > 0 {
				if !found {
					rp.formatter.PrintInfo("", "Used by:")
					found = true
				}
				rp.formatter.PrintRelation("Pod", pod.Name, usedAs...)
			}
		}

		rp.formatter.Outdent()
	}

	return nil
}

func (rp *ResourceProcessor) ShowSecretUsage(namespace string) error {
	fmt.Println("\n[Secret Layer]")
	secrets, err := rp.clientset.CoreV1().Secrets(namespace).List(rp.ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting secrets: %v", err)
	}

	for i, secret := range secrets.Items {
		isLast := i == len(secrets.Items)-1
		prefix := "├──"
		if isLast {
			prefix = "└──"
		}

		rp.formatter.PrintResource(prefix, "Secret", secret.Name)
		rp.formatter.Indent()

		rp.formatter.PrintInfo("", "Type: %s", secret.Type)
		rp.formatter.PrintInfo("", "Data Keys: %d", len(secret.Data))

		pods, err := rp.clientset.CoreV1().Pods(namespace).List(rp.ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error getting pods: %v", err)
		}

		found := false
		for _, pod := range pods.Items {
			usedAs := rp.getSecretUsageInPod(&pod, secret.Name)
			if len(usedAs) > 0 {
				if !found {
					rp.formatter.PrintInfo("", "Used by:")
					found = true
				}
				rp.formatter.PrintRelation("Pod", pod.Name, usedAs...)
			}
		}

		rp.formatter.Outdent()
	}

	return nil
}

func (rp *ResourceProcessor) getConfigMapUsageInPod(pod *corev1.Pod, configMapName string) []string {
	var usages []string

	// Check volume mounts
	for _, volume := range pod.Spec.Volumes {
		if volume.ConfigMap != nil && volume.ConfigMap.Name == configMapName {
			usages = append(usages, fmt.Sprintf("Mounted as volume: %s", volume.Name))
		}
	}

	// Check containers
	for _, container := range pod.Spec.Containers {
		for _, envFrom := range container.EnvFrom {
			if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name == configMapName {
				usages = append(usages, fmt.Sprintf("Used in envFrom by container: %s", container.Name))
			}
		}

		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil &&
				env.ValueFrom.ConfigMapKeyRef.Name == configMapName {
				usages = append(usages, fmt.Sprintf("Used as env var '%s' in container: %s", env.Name, container.Name))
			}
		}
	}

	return usages
}

func (rp *ResourceProcessor) getSecretUsageInPod(pod *corev1.Pod, secretName string) []string {
	var usages []string

	// Check volume mounts
	for _, volume := range pod.Spec.Volumes {
		if volume.Secret != nil && volume.Secret.SecretName == secretName {
			usages = append(usages, fmt.Sprintf("Mounted as volume: %s", volume.Name))
		}
	}

	// Check containers
	for _, container := range pod.Spec.Containers {
		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil && envFrom.SecretRef.Name == secretName {
				usages = append(usages, fmt.Sprintf("Used in envFrom by container: %s", container.Name))
			}
		}

		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil &&
				env.ValueFrom.SecretKeyRef.Name == secretName {
				usages = append(usages, fmt.Sprintf("Used as env var '%s' in container: %s", env.Name, container.Name))
			}
		}
	}

	return usages
}

func (rp *ResourceProcessor) ProcessNamespace(namespace string) error {
	rp.formatter.PrintHeader(fmt.Sprintf("Analyzing namespace: %s", namespace))
	rp.formatter.PrintLine()

	if err := rp.ShowResourceRelationships(namespace); err != nil {
		return err
	}

	if err := rp.ShowDeploymentDetails(namespace); err != nil {
		return err
	}

	if err := rp.ShowHPADetails(namespace); err != nil {
		return err
	}

	if err := rp.ShowConfigMapUsage(namespace); err != nil {
		return err
	}

	if err := rp.ShowSecretUsage(namespace); err != nil {
		return err
	}

	rp.formatter.PrintLine()
	return nil
}

// Adding relationships

func (rp *ResourceProcessor) ShowResourceRelationships(namespace string) error {
	fmt.Println("External Traffic")
	fmt.Println("│")

	// Handle Ingresses
	fmt.Println("[Ingress Layer]")
	ingresses, err := rp.clientset.NetworkingV1().Ingresses(namespace).List(rp.ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting ingresses: %v", err)
	}

	for i, ingress := range ingresses.Items {
		isLast := i == len(ingresses.Items)-1
		prefix := "├──"
		if isLast {
			prefix = "└──"
		}

		rp.formatter.PrintResource(prefix, "Ingress", ingress.Name)
		rp.formatter.Indent()

		// Check TLS
		if len(ingress.Spec.TLS) > 0 {
			rp.formatter.PrintStatus("TLS Enabled", true)
			for _, tls := range ingress.Spec.TLS {
				rp.formatter.PrintInfo("", "Hosts: %v", tls.Hosts)
				if tls.SecretName != "" {
					rp.formatter.PrintInfo("", "TLS Secret: %s", tls.SecretName)
				}
			}
		}

		// Process rules
		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					details := []string{
						fmt.Sprintf("via host: %s", rule.Host),
						fmt.Sprintf("path: %s", path.Path),
						fmt.Sprintf("pathType: %s", path.PathType),
					}

					if path.Backend.Service != nil {
						rp.formatter.PrintRelation("Service", path.Backend.Service.Name, details...)
						if path.Backend.Service.Port.Number > 0 {
							rp.formatter.PrintInfo("", "  Port: %d", path.Backend.Service.Port.Number)
						}
					}
				}
			}
		}
		rp.formatter.Outdent()
	}

	// Handle Services
	fmt.Println("\n[Service Layer]")
	services, err := rp.clientset.CoreV1().Services(namespace).List(rp.ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting services: %v", err)
	}

	for i, service := range services.Items {
		isLast := i == len(services.Items)-1
		prefix := "├──"
		if isLast {
			prefix = "└──"
		}

		rp.formatter.PrintResource(prefix, "Service", service.Name)
		rp.formatter.Indent()

		// Show service details
		rp.formatter.PrintInfo("", "Type: %s", service.Spec.Type)
		if service.Spec.ClusterIP != "" {
			rp.formatter.PrintInfo("", "ClusterIP: %s", service.Spec.ClusterIP)
		}
		if len(service.Spec.ExternalIPs) > 0 {
			rp.formatter.PrintInfo("", "External IPs: %v", service.Spec.ExternalIPs)
		}

		// Show port mappings
		for _, port := range service.Spec.Ports {
			portInfo := fmt.Sprintf("Port: %d→%d/%s", port.Port, port.TargetPort.IntVal, port.Protocol)
			if port.NodePort > 0 {
				portInfo += fmt.Sprintf(" (NodePort: %d)", port.NodePort)
			}
			rp.formatter.PrintInfo("", portInfo)
		}

		// Show selector and find matching pods
		if len(service.Spec.Selector) > 0 {
			rp.formatter.PrintInfo("", "Selector: %v", service.Spec.Selector)

			labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{
				MatchLabels: service.Spec.Selector,
			})
			pods, err := rp.clientset.CoreV1().Pods(namespace).List(rp.ctx, metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			if err != nil {
				return fmt.Errorf("error getting pods for service %s: %v", service.Name, err)
			}

			if len(pods.Items) > 0 {
				rp.formatter.PrintInfo("", "Connected Pods:")
				for _, pod := range pods.Items {
					details := []string{
						fmt.Sprintf("Status: %s", pod.Status.Phase),
						fmt.Sprintf("Node: %s", pod.Spec.NodeName),
					}
					if pod.Status.PodIP != "" {
						details = append(details, fmt.Sprintf("PodIP: %s", pod.Status.PodIP))
					}
					rp.formatter.PrintRelation("Pod", pod.Name, details...)
				}
			} else {
				rp.formatter.PrintStatus("No pods found matching selector", false)
			}
		}

		// Show any associated endpoints
		endpoints, err := rp.clientset.CoreV1().Endpoints(namespace).Get(rp.ctx, service.Name, metav1.GetOptions{})
		if err == nil && len(endpoints.Subsets) > 0 {
			rp.formatter.PrintInfo("", "Endpoints:")
			for _, subset := range endpoints.Subsets {
				for _, addr := range subset.Addresses {
					target := ""
					if addr.TargetRef != nil {
						target = fmt.Sprintf(" (%s: %s)", addr.TargetRef.Kind, addr.TargetRef.Name)
					}
					rp.formatter.PrintInfo("", "  %s%s", addr.IP, target)
				}
			}
		}

		rp.formatter.Outdent()
	}

	return nil
}
