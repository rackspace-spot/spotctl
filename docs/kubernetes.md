# Rackspace Spot CLI (spotctl) - Kubernetes Resources Management

This guide covers all Kubernetes-related commands in `spotctl`: **Cloudspaces**, **NodePools** - **Spot** and **Ondemand**. This guide is focused on Kubernetes resource lifecycle management in Rackspace Spot.

## Features
- Complete Kubernetes resource lifecycle management
- Multiple output formats (JSON, YAML, Table)
- OAuth2 authentication
- Interactive command structure
- Cross-platform support
- Easy configuration management

## Installation

### Prerequisites

- User should have access to Rackspace Spot Organization.
- User should have Refresh Token for the corresponding Organization.
- `<org>` refers to organization name whenever used in commands.

Download the binary from the releases page: https://github.com/rackspace-spot/spotctl/releases

Move the binary to a directory in your PATH, for example:
```bash
sudo mv spotctl /usr/local/bin/
```

Verify installation:
```bash
spotctl --version
```

## Configuration

In order to use spotctl, you need to configure your spotctl: You need pass the organization name, region and refresh token.

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
- `spotctl nodepools autopilot list` - List autopilot node pools
- `spotctl nodepools autopilot create` - Create an autopilot node pool

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

[![Video preview](../tools/interactive-cloudspace-creation.gif)](tools/interactive-cloudspace-creation.webm)

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
spotctl organizations list --output table

# Get organization details
spotctl organizations get org-123 --output json
```

### Node Pools

#### Spot Node Pools
```bash
# Create spot node pool
spotctl nodepools spot create \
  --namespace org-123 \
  --cloudspace prod-cluster \
  --serverclass gp.vs1.medium-iad \
  --desired 5 \
  --bidprice 0.85

# List spot pools
spotctl nodepools spot list --namespace org-123 --output yaml
```

#### On-Demand Node Pools
```bash
# Create on-demand node pool
spotctl nodepools ondemand create \
  --namespace org-123 \
  --cloudspace prod-cluster \
  --serverclass mem.vs1.large-iad \
  --desired 3

# List on-demand pools
spotctl nodepools ondemand list --namespace org-123
```

#### Autopilot Node Pools

An autopilot node pool takes a total vCPU target and an hourly budget ceiling. Spot picks the
eligible server classes and bids across them to fill the target, so you do not name a server class
or a node count.

```bash
# Create an autopilot node pool
spotctl nodepools autopilot create \
  --cloudspace prod-cluster \
  --vcpu-total 32 \
  --budget-per-hour 2.50

# Only pack from 4 GB/vCPU classes, 4 to 16 vCPUs per node
spotctl nodepools autopilot create \
  --cloudspace prod-cluster \
  --vcpu-total 32 \
  --budget-per-hour 2.50 \
  --memory-per-vcpu 4 \
  --vcpu-per-node-min 4 \
  --vcpu-per-node-max 16

# List autopilot pools
spotctl nodepools autopilot list --cloudspace prod-cluster --output yaml

# Show one pool with its per-server-class allocations
spotctl nodepools autopilot get --name b7ea7dd1-f421-4b81-96a5-c28a6400a420

# Raise the target. Flags you leave out keep their current values.
spotctl nodepools autopilot update \
  --name b7ea7dd1-f421-4b81-96a5-c28a6400a420 \
  --vcpu-total 64

spotctl nodepools autopilot delete --name b7ea7dd1-f421-4b81-96a5-c28a6400a420
```

Notes:

- `--region` defaults to the cloudspace's region, which is the only value the API accepts.
- `--vcpu-total` must be a multiple of the smallest eligible node size in the region, so an exact
  pack exists. The API rejects the create otherwise and names the size it wanted.
- `--budget-per-hour` is a bid ceiling, not a spend cap. You are billed the market clearing rate,
  which is always at or below the ceiling.
- Region and cloudspace are fixed once the pool exists; `update` does not offer them.

#### Some commands examples 

```bash
spotctl cloudspaces create \
  --name cli-test-205 \
  --region us-central-ord-1 \
  --org hooli \
  --spot-nodepool "serverclass=gp.vs1.medium-ord,desired=2,bidprice=0.09"

spotctl cloudspaces create \
  --name cli-test-157 \
  --region us-central-ord-1 \
  --org hooli \
  --ondemand-nodepool "serverclass=gp.vs1.medium-ord,desired=2"


spotctl nodepools spot create --cloudspace cli-test-153 --desired 1 --serverclass gp.vs1.medium-ord --bidprice 0.08

spotctl nodepools spot update --name b7ea7dd1-f421-4b81-96a5-c28a6400a420 --cloudspace cli-test-153 --desired 2 --bidprice 0.08

spotctl nodepools spot delete --name b7ea7dd1-f421-4b81-96a5-c28a6400a420 

spotctl nodepools spot get --name b7ea7dd1-f421-4b81-96a5-c28a6400a420 


Ondemand Nodepool Operations

spotctl nodepools ondemand list --cloudspace cli-test-153

spotctl nodepools ondemand get --name b7ea7dd1-f421-4b81-96a5-c28a6400a406

spotctl nodepools ondemand update --name b7ea7dd1-f421-4b81-96a5-c28a6400a406 --cloudspace cli-test-153 --desired 2

spotctl nodepools ondemand create --cloudspace cli-test-153 --desired 1 --serverclass gp.vs1.medium-ord

spotctl nodepools ondemand delete --name b7ea7dd1-f421-4b81-96a5-c28a6400a406

```

#### Some commands examples 

```bash
spotctl cloudspaces create \
  --name cli-test-205 \
  --region us-central-ord-1 \
  --org hooli \
  --spot-nodepool "serverclass=gp.vs1.medium-ord,desired=2,bidprice=0.09"

spotctl cloudspaces create \
  --name cli-test-157 \
  --region us-central-ord-1 \
  --org hooli \
  --ondemand-nodepool "serverclass=gp.vs1.medium-ord,desired=2"


spotctl nodepools spot create --cloudspace cli-test-153 --desired 1 --serverclass gp.vs1.medium-ord --bidprice 0.08

spotctl nodepools spot update --name b7ea7dd1-f421-4b81-96a5-c28a6400a420 --cloudspace cli-test-153 --desired 2 --bidprice 0.08

spotctl nodepools spot delete --name b7ea7dd1-f421-4b81-96a5-c28a6400a420 

spotctl nodepools spot get --name b7ea7dd1-f421-4b81-96a5-c28a6400a420 


Ondemand Nodepool Operations

spotctl nodepools ondemand list --cloudspace cli-test-153

spotctl nodepools ondemand get --name b7ea7dd1-f421-4b81-96a5-c28a6400a406

spotctl nodepools ondemand update --name b7ea7dd1-f421-4b81-96a5-c28a6400a406 --cloudspace cli-test-153 --desired 2

spotctl nodepools ondemand create --cloudspace cli-test-153 --desired 1 --serverclass gp.vs1.medium-ord

spotctl nodepools ondemand delete --name b7ea7dd1-f421-4b81-96a5-c28a6400a406

```



## Output Formats

| Format | Description                      | Example Command                              |
|--------|----------------------------------|----------------------------------------------|
| JSON   | Structured JSON output (default) | `spotctl regions list --output json`        |
| Table  | Human-readable table format      | `spotctl server-classes list --output table`|
| YAML   | YAML-formatted output            | `spotctl organizations list --output yaml`  |
