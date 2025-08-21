# Rackspace Spot CLI (spotctl)

A command-line interface for managing Rackspace Spot resources with full CRUD operations for all resource types including cloudspaces, node pools, organizations, and more.

## Features

- Complete resource lifecycle management
- Multiple output formats (JSON, YAML, Table)
- OAuth2 authentication
- Interactive command structure
- Cross-platform support
- Easy configuration management

## Installation

### Prerequisites

- Go 1.16 or later
- Rackspace Spot API credentials

### Option 1: Install using Go

```bash
# Install the latest version
go install github.com/rackspace-spot/spotctl@latest

# Verify installation
spotctl --version
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/rackspace-spot/spotctl.git
cd spotctl

# Build the binary
go build -o spotctl

# Move to a directory in your PATH
sudo mv spotctl /usr/local/bin/

# Verify installation
spotctl --help
```

## Configuration

Before using spotctl, you need to configure your credentials:

```bash
# Run the interactive configuration wizard
spotctl configure

# Or set environment variables manually
export SPOT_ORG_ID="your_organization_id"
export SPOT_API_TOKEN="your_api_token"
export SPOT_REGION="your_preferred_region"  # Optional
```

## Available Commands

### Authentication
- `spotctl auth` - Authenticate with Rackspace Spot

### Cloudspaces (Kubernetes Clusters)
- `spotctl cloudspaces list` - List all cloudspaces
- `spotctl cloudspaces get <name>` - Get details of a specific cloudspace
- `spotctl cloudspaces create` - Create a new cloudspace
- `spotctl cloudspaces delete <name>` - Delete a cloudspace
- `spotctl cloudspaces get-config <name>` - Get kubeconfig for a cloudspace

### Node Pools
- `spotctl nodepools spot list` - List spot node pools
- `spotctl nodepools spot create` - Create a spot node pool
- `spotctl nodepools ondemand list` - List on-demand node pools
- `spotctl nodepools ondemand create` - Create an on-demand node pool

### Server Classes
- `spotctl serverclasses list` - List available server classes
- `spotctl serverclasses get <name>` - Get details of a server class

### Regions
- `spotctl regions list` - List available regions
- `spotctl regions get <name>` - Get details of a region

### Organizations
- `spotctl organizations list` - List organizations
- `spotctl organizations get <id>` - Get organization details

### Pricing
- `spotctl pricing get` - Get pricing information

## Usage Examples

### List all cloudspaces
```bash
spotctl cloudspaces list
```

### Create a new cloudspace
```bash
spotctl cloudspaces create --name my-cluster --region us-east-1
```

### Get kubeconfig for a cloudspace
```bash
spotctl cloudspaces get-config my-cluster --file ~/.kube/config-my-cluster
```

### Create a spot node pool
```bash
spotctl nodepools spot create \
  --name my-spot-pool \
  --cloudspace my-cluster \
  --server-class gp.vs1.medium-iad \
  --desired 3 \
  --bid-price 0.05
```

## Output Formats

Most commands support multiple output formats. Use the `-o` or `--output` flag:

```bash
# JSON (default)
spotctl cloudspaces list -o json

# YAML
spotctl cloudspaces list -o yaml

# Table (human-readable)
spotctl cloudspaces list -o table
```

## Troubleshooting

Enable verbose output for debugging:
```bash
spotctl -v=3 cloudspaces list
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

### Organizations
```bash
# List all organizations
spot-cli organizations list --output table

# Get organization details
spot-cli organizations get org-123 --output json
```

### Cloudspaces
```bash
# Create a new cloudspace
spot-cli cloudspaces create \
  --name prod-cluster \
  --org org-123 \
  --region us-east-iad-1 \
  --kubernetes-version 1.28

# List cloudspaces in an organization
spot-cli cloudspaces list --org org-123

# Delete a cloudspace
spot-cli cloudspaces delete --org org-123 --name staging-cluster
```

### Node Pools

#### Spot Node Pools
```bash
# Create spot node pool
spot-cli nodepools spot create \
  --name spot-workers \
  --namespace org-123 \
  --cloudspace prod-cluster \
  --server-class gp.vs1.medium-iad \
  --desired 5 \
  --bid-price 0.85

# List spot pools
spot-cli nodepools spot list --namespace org-123 --output yaml
```

#### On-Demand Node Pools
```bash
# Create on-demand node pool
spot-cli nodepools ondemand create \
  --name critical-workers \
  --namespace org-123 \
  --cloudspace prod-cluster \
  --server-class mem.vs1.large-iad \
  --desired 3

# List on-demand pools
spot-cli nodepools ondemand list --namespace org-123
```

### Utilities
```bash
# List available regions
spot-cli regions list --output table

# Show server class details
spot-cli server-classes list

# Get price history
spot-cli price-history get --server-class gp.vs1.medium-iad
```

## Example Workflow

```bash
# Authenticate
export SPOT_REFRESH_TOKEN="your_token"
spot-cli auth

# Create infrastructure
spot-cli cloudspaces create --name my-cluster --org org-123 --region us-east-iad-1
spot-cli nodepools spot create --name spot-pool --namespace org-123 --cloudspace my-cluster --server-class gp.vs1.medium-iad --desired 5 --bid-price 0.75
spot-cli nodepools ondemand create --name ondemand-pool --namespace org-123 --cloudspace my-cluster --server-class mem.vs1.large-iad --desired 2

# Query resources
spot-cli cloudspaces list --org org-123 --output table
spot-cli nodepools spot list --namespace org-123 --output json
```

## Output Formats

| Format | Description                      | Example Command                              |
|--------|----------------------------------|----------------------------------------------|
| JSON   | Structured JSON output (default) | `spot-cli regions list --output json`        |
| Table  | Human-readable table format      | `spot-cli server-classes list --output table`|
| YAML   | YAML-formatted output            | `spot-cli organizations list --output yaml`  |

## Troubleshooting

Enable verbose logging with `-v` flag:
```bash
spot-cli -v cloudspaces list --namespace org-123
```

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/improvement`)
3. Commit changes (`git commit -am 'Add new feature'`)
4. Push to branch (`git push origin feature/improvement`)
5. Create Pull Request

## License

Apache License 2.0
