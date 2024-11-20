package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mbergo/k8s-microlens/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ResourceMapper holds the Kubernetes client and context
type ResourceMapper struct {
	clientset *kubernetes.Clientset
	ctx       context.Context
	formatter *common.Formatter
	processor *common.ResourceProcessor
}

// stringSliceFlag implements flag.Value interface for string slice flags
type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// NewResourceMapper creates a new ResourceMapper instance
func NewResourceMapper() (*ResourceMapper, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error getting home directory: %v", err)
		}
		kubeconfig = homeDir + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes client: %v", err)
	}

	ctx := context.Background()
	formatter := common.NewFormatter()
	processor := common.NewResourceProcessor(clientset, ctx)

	return &ResourceMapper{
		clientset: clientset,
		ctx:       ctx,
		formatter: formatter,
		processor: processor,
	}, nil
}

func (rm *ResourceMapper) getNamespaces(targetNs string, excludeNs []string) ([]string, error) {
	var namespaces []string
	if targetNs != "" {
		_, err := rm.clientset.CoreV1().Namespaces().Get(rm.ctx, targetNs, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("namespace '%s' not found", targetNs)
		}
		namespaces = []string{targetNs}
	} else {
		nsList, err := rm.clientset.CoreV1().Namespaces().List(rm.ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, ns := range nsList.Items {
			if !contains(excludeNs, ns.Name) {
				namespaces = append(namespaces, ns.Name)
			}
		}
	}
	return namespaces, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func printHelp() {
	fmt.Println("Kubernetes MicroLens - A lightweight Kubernetes resource visualization tool")
	fmt.Println("\nUsage:")
	fmt.Println("  k8s-microlens [flags]")
	fmt.Println("\nFlags:")
	fmt.Println("  -n, --namespace string     Process only the specified namespace")
	fmt.Println("  --exclude-ns string        Exclude specified namespaces (can be specified multiple times)")
	fmt.Println("  -h, --help                Show help message")
	fmt.Println("  -v, --version             Show version information")
	fmt.Println("\nExamples:")
	fmt.Println("  # Show resources in all namespaces")
	fmt.Println("  k8s-microlens")
	fmt.Println("\n  # Show resources in specific namespace")
	fmt.Println("  k8s-microlens -n default")
	fmt.Println("\n  # Exclude specific namespaces")
	fmt.Println("  k8s-microlens --exclude-ns kube-system --exclude-ns kube-public")
}

func printVersion() {
	fmt.Println("Kubernetes MicroLens v0.1.0")
	fmt.Println("A lightweight Kubernetes resource visualization tool")
	fmt.Println("\nAuthor: Marcus Bergo <marcus.bergo@gmail.com>")
	fmt.Println("Repository: https://github.com/mbergo/k8s-microlens")
}

func main() {
	var (
		namespace = flag.String("n", "", "Process only the specified namespace")
		excludeNs stringSliceFlag
		help      = flag.Bool("h", false, "Show help message")
		version   = flag.Bool("v", false, "Show version information")
	)

	flag.StringVar(namespace, "namespace", "", "Process only the specified namespace")
	flag.Var(&excludeNs, "exclude-ns", "Exclude specified namespaces (can be specified multiple times)")
	flag.BoolVar(help, "help", false, "Show help message")
	flag.BoolVar(version, "version", false, "Show version information")

	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	if *version {
		printVersion()
		os.Exit(0)
	}

	rm, err := NewResourceMapper()
	if err != nil {
		fmt.Printf("%sError initializing resource mapper: %v%s\n", common.ColorRed, err, common.ColorReset)
		os.Exit(1)
	}

	rm.formatter.PrintHeader("Kubernetes MicroLens")
	fmt.Printf("Generated at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	rm.formatter.PrintLine()

	namespaces, err := rm.getNamespaces(*namespace, excludeNs)
	if err != nil {
		fmt.Printf("%sError getting namespaces: %v%s\n", common.ColorRed, err, common.ColorReset)
		os.Exit(1)
	}

	// Process each namespace
	for _, ns := range namespaces {
		if err := rm.processor.ProcessNamespace(ns); err != nil {
			fmt.Printf("%sError processing namespace %s: %v%s\n", common.ColorRed, ns, err, common.ColorReset)
			continue
		}
	}

	rm.formatter.PrintSuccess("Resource mapping complete!")
}
