# k8s-microlens 🔍

A lightweight, fast, and efficient Kubernetes resource visualization tool that provides a clear and hierarchical view of your cluster resources and their relationships.

[![Go Report Card](https://goreportcard.com/badge/github.com/mbergo/k8s-microlens)](https://goreportcard.com/report/github.com/mbergo/k8s-microlens)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Build Status](https://github.com/mbergo/k8s-microlens/actions/workflows/go.yml/badge.svg)


## Features ✨

- **Resource Discovery**: Automatically maps relationships between different Kubernetes resources
- **Clear Visualization**: Presents resources in a clear, hierarchical tree structure
- **Color-Coded Output**: Uses colors and symbols for better readability
- **Comprehensive Resource Coverage**:
  - Ingresses (with TLS details)
  - Services (with endpoint information)
  - Deployments (with replica details)
  - Pods (with status and node assignment)
  - ConfigMaps (with usage tracking)
  - Secrets (with secure usage information)
  - HPAs (with scaling metrics)

## Installation 📦

### Prerequisites

- Go 1.19 or later
- Access to a Kubernetes cluster
- Valid kubeconfig file

### From Source

```bash
# Clone the repository
git clone https://github.com/mbergo/k8s-microlens.git
cd k8s-microlens

# Build the binary
go build -o k8s-microlens cmd/mapper/main.go

# (Optional) Move to PATH
sudo mv k8s-microlens /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/mbergo/k8s-microlens@latest
```

## Usage 🚀

### Basic Usage

```bash
# Show resources in all namespaces
k8s-microlens

# Show resources in a specific namespace
k8s-microlens -n default

# Exclude specific namespaces
k8s-microlens --exclude-ns kube-system --exclude-ns kube-public
```

### Command Line Options

```
Flags:
  -n, --namespace string     Process only the specified namespace
  --exclude-ns string       Exclude specified namespaces (can be specified multiple times)
  -h, --help               Show help message
  -v, --version            Show version information
```

## Example Output 📝

```
Kubernetes MicroLens
----------------------------------------
Generated at: 2024-11-18 15:04:05

External Traffic
│
[Ingress Layer]
├── ● Ingress/frontend-ingress
│   ✓ TLS Enabled
│   ➜ Service/frontend-svc via host: example.com
└── ● Ingress/api-ingress
    ➜ Service/api-svc via host: api.example.com

[Service Layer]
├── ● Service/frontend-svc
│   ℹ Type: ClusterIP
│   ℹ Ports: 80→8080/TCP
│   ➜ Pod/frontend-pod-1
│      ✓ Running
└── ● Service/api-svc
    ℹ Type: ClusterIP
    ℹ Ports: 8080→8080/TCP
    ➜ Pod/api-pod-1
       ✓ Running
```

## Output Legend 📚

- `●` Resource indicator
- `✓` Success/Enabled status
- `✗` Error/Disabled status
- `ℹ` Information
- `➜` Relationship/Connection
- `├──` Tree branch
- `└──` Last tree branch

## Project Structure 📁

```
.
├── cmd/
│   └── mapper/
│       └── main.go           # Application entry point
├── internal/
│   └── common/
│       ├── formatting.go     # Output formatting utilities
│       └── resources.go      # Resource processing logic
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## Development 🛠️

### Running Tests

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Coding Style

Follow Go best practices and conventions:
- Use `gofmt` for code formatting
- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Write tests for new functionality
- Document exported functions and types

## Troubleshooting 🔧

### Common Issues

1. **Permission Errors**
   ```
   Error getting resources: forbidden
   ```
   Solution: Ensure your kubeconfig has sufficient permissions

2. **Kubeconfig Not Found**
   ```
   Error building kubeconfig
   ```
   Solution: Set KUBECONFIG environment variable or ensure config exists in `~/.kube/config`

3. **Connection Issues**
   ```
   Error connecting to cluster
   ```
   Solution: Verify cluster access and network connectivity

## Roadmap 🗺️

- [ ] Support for Custom Resource Definitions (CRDs)
- [ ] Export functionality (JSON, YAML, DOT formats)
- [ ] Interactive mode with real-time updates
- [ ] Resource metrics integration
- [ ] Custom output formatting templates
- [ ] WebUI interface

## License 📄

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support ❤️

If you find this project useful, please consider:
- Starring the repository
- Reporting issues
- Contributing to the code
- Sharing with others

## Author ✍️

**Marcus Bergo**
- Twitter: [@mbergo](https://twitter.com/mbergo)
- LinkedIn: [Marcus Bergo](https://linkedin.com/in/marcusbergo)

## Acknowledgments 🙏

- The Kubernetes community
- Contributors to the project
- Users who provide feedback and suggestions

---

Made with ❤️ by Marcus Bergo and contributors
