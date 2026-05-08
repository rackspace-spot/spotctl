# Rackspace Spot CLI — Virtual Machines (VM) Resources

This guide covers all VM-related commands in `spotctl`: **VM Cloudspaces**, **VM Pools**, and **VM SSH Keys**. This guide is focused on VM lifecycle management in Rackspace Spot.

## Features
- Complete Virtual Machines (VM) lifecycle management
- Multiple output formats (JSON, YAML, Table)
- OAuth2 authentication
- Interactive command structure
- Cross-platform support
- Easy configuration management

## Installation

Download the latest binary from the releases page: https://github.com/rackspace-spot/spotctl/releases

Move the binary to a directory in your PATH, for example:
```bash
sudo mv spotctl /usr/local/bin/
```

## Prerequisites

- Access to a Rackspace Spot Organization
- A valid Refresh Token for your Organization
- `spotctl` configured (see [Configuration](#configuration))

## Configuration

```bash
spotctl configure
```

You will be prompted for:
- **Organization** — your org name (e.g., `org-spot`)
- **Region** — e.g., `us-central-ord-1`
- **Refresh Token** — your OAuth2 refresh token

Once configured, `--org` and `--region` flags are optional (they default to your configured values).

If the binary is not installed globally (not in your PATH), run it like this:

```bash
./spotctl <command>
```

> 💡 Note: `./` is required when running a binary from the current directory.
---

## VM SSH Keys

Manage SSH keys used by VM cloudspaces.
#### Aliases for command:  `vmsshkeys`, `vmsshkey`, `vmsk`

### List VM SSH Keys
```bash
spotctl vmsshkeys list
spotctl vmsshkeys list --org org-spot
```

### Create a VM SSH Key
```bash
spotctl vmsshkeys create \
  --name my-ssh-key \
  --public-key "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ... user@host" \
  --description "My dev SSH key"
```

### Get VM SSH Key Details
```bash
spotctl vmsshkeys get --name my-ssh-key
```

### Delete a VM SSH Key
```bash
spotctl vmsshkeys delete --name my-ssh-key
spotctl vmsshkeys delete --name my-ssh-key -y   # skip confirmation
```
---

## VM Cloudspaces

Manage VM cloudspaces — the top-level resource that groups VM pools and references an SSH key.
#### Aliases for command:  `vmcloudspaces`, `vmcloudspace`, `vmcs`

### List VM Cloudspaces
```bash
spotctl vmcloudspaces list
spotctl vmcloudspaces list --org org-spot
```

### Create a VM Cloudspace

#### Using CLI Flags
```bash
spotctl vmcloudspaces create \
  --name my-vm-cs \
  --region us-central-ord-1 \
  --vm-ssh-key-name my-ssh-key
```

With optional webhook and inline VM pool:
```bash
spotctl vmcloudspaces create \
  --name my-vm-cs \
  --region us-central-ord-1 \
  --vm-ssh-key-name my-ssh-key \
  --webhook https://example.com/webhook \
  --vm-pool "serverclass=gp.vs1.medium-ord,bidprice=0.02,desired=3,pooltype=spot,vmimage=ubuntu24.04"
```

With inline VM pool and cloud-init user data (applied to all inline pools):
```bash
spotctl vmcloudspaces create \
  --name my-vm-cs \
  --region us-central-ord-1 \
  --vm-ssh-key-name my-ssh-key \
  --vm-pool "serverclass=gp.vs1.medium-ord,bidprice=0.02,desired=3,pooltype=spot,vmimage=ubuntu24.04" \
  --vm-userdata '#!/bin/bash
apt-get update -y
apt-get install -y apache2 tree htop
systemctl enable apache2
systemctl start apache2'
```

With cloud-init user data from a script file:
```bash
spotctl vmcloudspaces create \
  --name my-vm-cs \
  --region us-central-ord-1 \
  --vm-ssh-key-name my-ssh-key \
  --vm-pool "serverclass=gp.vs1.medium-ord,bidprice=0.02,desired=3,pooltype=spot,vmimage=ubuntu24.04" \
  --vm-userdata-from-script ./cloud-init.sh
```

> **Note:** `--vm-userdata` and `--vm-userdata-from-script` are mutually exclusive. The user data is automatically base64-encoded before being sent to the API. If the input is already base64-encoded, it is sent as-is.

Optionally specify the SSH key namespace (defaults to org namespace):
```bash
spotctl vmcloudspaces create \
  --name my-vm-cs \
  --region us-central-ord-1 \
  --vm-ssh-key-name my-ssh-key \
  --vm-ssh-key-namespace custom-namespace
```

#### Using a Config File (YAML)

Create a file `vm-cloudspace.yaml`:
```yaml
vmCloudSpace:
  name: my-vm-cs
  org: org-spot
  region: us-central-ord-1
  vmSshKeyRef:
    name: my-ssh-key
  webhook: https://example.com/webhook
vmPools:
  - serverClass: gp.vs1.medium-ord
    bidPrice: "0.02"
    desired: 3
    poolType: spot
    vmImage: ubuntu24.04
    vmUserData: "IyEvYmluL2Jhc2gKYXB0LWdldCB1cGRhdGU="   # base64-encoded cloud-init script (optional)
```

```bash
spotctl vmcloudspaces create --config vm-cloudspace.yaml
```

#### Using a Config File (JSON)
```json
{
  "vmCloudSpace": {
    "name": "my-vm-cs",
    "org": "org-spot",
    "region": "us-central-ord-1",
    "vmSshKeyRef": {
      "name": "my-ssh-key"
    },
    "webhook": "https://example.com/webhook"
  },
  "vmPools": [
    {
      "serverClass": "gp.vs1.medium-ord",
      "bidPrice": "0.02",
      "desired": 3,
      "poolType": "spot",
      "vmImage": "ubuntu24.04",
      "vmUserData": "IyEvYmluL2Jhc2gKYXB0LWdldCB1cGRhdGU="
    }
  ]
}
```

```bash
spotctl vmcloudspaces create --config vm-cloudspace.json
```

### Get VM Cloudspace Details
```bash
spotctl vmcloudspaces get --name my-vm-cs
```

### Update a VM Cloudspace

Currently only the **webhook** field can be updated:
```bash
spotctl vmcloudspaces update --name my-vm-cs --webhook https://example.com/new-webhook
```

To clear the webhook:
```bash
spotctl vmcloudspaces update --name my-vm-cs --webhook ""
```

### Delete a VM Cloudspace
```bash
spotctl vmcloudspaces delete --name my-vm-cs
spotctl vmcloudspaces delete --name my-vm-cs -y   # skip confirmation
```

---

## VM Pools

Manage VM pools within a VM cloudspace — both spot and on-demand types.
#### Aliases for command:  `vmpools`, `vmpool`, `vmp`

### List VM Pools
```bash
spotctl vmpools list --vmcloudspace my-vm-cs
```

### Create a VM Pool
```bash
spotctl vmpools create \
  --vmcloudspace my-vm-cs \
  --serverclass gp.vs1.medium-ord \
  --bidprice 0.02 \
  --desired 3 \
  --pooltype spot \
  --vmimage ubuntu24.04
```

With cloud-init user data (raw text or base64-encoded):
```bash
spotctl vmpools create \
  --vmcloudspace my-vm-cs \
  --serverclass gp.vs1.medium-ord \
  --bidprice 0.02 \
  --desired 3 \
  --pooltype spot \
  --vmimage ubuntu24.04 \
  --vm-userdata '#!/bin/bash\napt-get update && apt-get install -y nginx'
```

With cloud-init user data from a script file:
```bash
spotctl vmpools create \
  --vmcloudspace my-vm-cs \
  --serverclass gp.vs1.medium-ord \
  --bidprice 0.02 \
  --desired 3 \
  --pooltype spot \
  --vmimage ubuntu24.04 \
  --vm-userdata-from-script ./cloud-init.sh
```

> **Note:** `--vm-userdata` and `--vm-userdata-from-script` are mutually exclusive. The data is automatically base64-encoded. If already base64-encoded, it is sent as-is.

### Get VM Pool Details
```bash
spotctl vmpools get --name <vm-pool-name>
```

### Update a VM Pool
```bash
# Update desired count
spotctl vmpools update --name <vm-pool-name> --desired 5

# Update bid price
spotctl vmpools update --name <vm-pool-name> --bidprice 0.05

# Update both
spotctl vmpools update --name <vm-pool-name> --desired 5 --bidprice 0.05
```

### Delete a VM Pool
```bash
spotctl vmpools delete --name <vm-pool-name>
spotctl vmpools delete --name <vm-pool-name> -y   # skip confirmation
```

---

## Output Formats

All list/get commands support output formatting:

| Format | Flag                | Description                      |
|--------|---------------------|----------------------------------|
| JSON   | `--output json`     | Structured JSON output (default) |
| YAML   | `--output yaml`     | YAML-formatted output            |
| Table  | `--output table`    | Human-readable table format      |

Example:
```bash
spotctl vmcloudspaces list --output yaml
```

---

## Run a Full Test Flow for Virtual Machines

```bash
spotctl configure
```

Enter your **org**, **region**, and **refresh token** when prompted. This stores config at `~/.spotctl/config.yaml`.

### Run below set of commands for Virtual Machines testing

```bash
# Step 1: Create an SSH key
spotctl vmsshkeys create \
  --name test-ssh-key \
  --public-key "ssh-rsa AAAAB3... user@host"

# Step 2: Verify SSH key
spotctl vmsshkeys get --name test-ssh-key

# Step 3: Create a VM cloudspace with a pool
spotctl vmcloudspaces create \
  --name test-vm-cs \
  --region us-central-ord-1 \
  --vm-ssh-key-name test-ssh-key \
  --vm-pool "serverclass=gp.vs1.medium-ord,bidprice=0.02,desired=1,pooltype=spot,vmimage=ubuntu24.04"

# Step 4: Verify cloudspace
spotctl vmcloudspaces get --name test-vm-cs

# Step 5: List all cloudspaces
spotctl vmcloudspaces list

# Step 6: Update cloudspace webhook
spotctl vmcloudspaces update --name test-vm-cs --webhook https://example.com/hook

# Step 7: Create another VM pool
spotctl vmpools create \
  --vmcloudspace test-vm-cs \
  --serverclass gp.vs1.medium-ord \
  --bidprice 0.05 \
  --desired 2 \
  --pooltype spot \
  --vmimage ubuntu24.04 \
  --vm-userdata '#!/bin/bash\napt-get update && apt-get install -y nginx'

# Step 8: List VM pools
spotctl vmpools list --vmcloudspace test-vm-cs

# Step 9: Update a VM pool
spotctl vmpools update --name <pool-name-from-step-7> --desired 3

# Step 10: Clean up
spotctl vmpools delete --name <pool-name> -y
spotctl vmcloudspaces delete --name test-vm-cs -y
spotctl vmsshkeys delete --name test-ssh-key -y
```

---

## Quick Reference

| Resource        | Command                  | Subcommands                          |
|-----------------|--------------------------|--------------------------------------|
| VM SSH Keys     | `spotctl vmsshkeys`      | `list`, `create`, `get`, `delete`    |
| VM Cloudspaces  | `spotctl vmcloudspaces`  | `list`, `create`, `get`, `update`, `delete` |
| VM Pools        | `spotctl vmpools`        | `list`, `create`, `get`, `update`, `delete` |
