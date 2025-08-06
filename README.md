# Rackspace Spot CLI

A command-line interface for managing Rackspace Spot resources with full CRUD operations for all resource types.

## Features

- Complete resource lifecycle management
- Multiple output formats (JSON, YAML, Table)
- OAuth2 authentication
- Interactive command structure
- Cross-platform support

## Installation

### Prerequisites

- Go 1.23.5 or later
- Rackspace Spot API credentials

### Build from Source

```bash
git clone https://github.com/rackspace-spot/spot-sdk.git
cd spot-sdk/spot-cli
go build -o spot-cli
```

### Install Globally

```bash
go install github.com/rackerlabs/spot-cli@latest
```

## Configuration

```bash
# Set refresh token
export SPOT_REFRESH_TOKEN="your_refresh_token"

# Verify authentication
spot-cli auth
```

## Command Reference

### Authentication
```bash
spot-cli auth
```

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
