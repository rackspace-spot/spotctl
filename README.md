# Rackspace Spot CLI (spotctl)

A command-line tool to manage Rackspace Spot resources including Kubernetes and VM cloudspaces, node pools, organizations, and more.

- For Kubernetes resources documentation, see [docs/kubernetes.md](docs/kubernetes.md)
- For VM resources documentation, see [docs/virtualmachines.md](docs/virtualmachines.md) 

## Features
- Complete resource lifecycle management
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

## General commands across all resources types (kubernetes & vm)

### Authentication
- `spotctl configure` - Configure spotctl

### Server Classes
- `spotctl serverclasses list` - List available server classes
- `spotctl serverclasses get <name>` - Get details of a server class

### Regions
- `spotctl regions list` - List available regions
- `spotctl regions get <name>` - Get details of a region

### Organizations
- `spotctl organizations list` - List organizations
- `spotctl organizations get --name <name>` - Get organization details by name

### Pricing
- `spotctl pricing get <serverclass>` - Get pricing information
- `spotctl pricing get-all` - Get pricing information for all server classes

## 🧑‍💻 Support
For documentation, please refer to the [official Rackspace Spot documentation](https://spot.rackspace.com/docs/en). For support, ask your questions in the [Rackspace community discussions](https://github.com/rackerlabs/spot/discussions), or drop us an email.

## 📜 License
**Copyright © Rackspace US, Inc. or its affiliates. All Rights Reserved.**  

`SPDX-License-Identifier: Apache-2.0`
