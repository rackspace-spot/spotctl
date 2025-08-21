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

Before using spotctl, you need to configure your credentials: You need pass the organization, region and refresh token.

```bash
# Run the interactive configuration wizard
spotctl configure
```

## Available Commands

### Authentication
- `spotctl configure` - Configure spotctl

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
- `spotctl pricing get <serverclass>` - Get pricing information

## Usage Examples

### List all cloudspaces
```bash
spotctl cloudspaces list
```

### Create a new cloudspace

#### Quick Create
```bash
spotctl cloudspaces create --name my-cluster --region us-east-1 
```

#### Config File
```bash
spotctl cloudspaces create --config my-cluster-config.yaml
```

#### Command Line Arguments (json)
```bash
spotctl cloudspaces create \
  --name <name> \
  --region <region> \
  --org <org> \
  --spot-nodepool '{"desired":1,"serverclass":"gp.vs1.medium-ord","bidprice":0.08}' \
  --ondemand-nodepool '{"desired":1,"serverclass":"gp.vs1.medium-ord"}' 
```
#### Command Line Arguments (comma separated)
```bash
spotctl cloudspaces create \
  --name <name> \
  --region <region> \
  --org <org> \
  --spot-nodepool desired=1,serverclass=gp.vs1.medium-ord,bidprice=0.08 \
  --ondemand-nodepool desired=1,serverclass=gp.vs1.medium-ord
```


### Get kubeconfig for a cloudspace
```bash
spotctl cloudspaces get-config my-cluster --file ~/.kube/config-my-cluster
```

### Delete a cloudspace 
```bash
spotctl cloudspaces delete --name <my-cluster>
```

### Organizations
```bash
# List all organizations
spot-cli organizations list --output table

# Get organization details
spot-cli organizations get org-123 --output json
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


## Output Formats

| Format | Description                      | Example Command                              |
|--------|----------------------------------|----------------------------------------------|
| JSON   | Structured JSON output (default) | `spot-cli regions list --output json`        |
| Table  | Human-readable table format      | `spot-cli server-classes list --output table`|
| YAML   | YAML-formatted output            | `spot-cli organizations list --output yaml`  |



## License

Apache License 2.0
