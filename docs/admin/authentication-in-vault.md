# Authentication in Vault

## Vault Setup

Heist relies on Kubernetes Authentication to authenticate its own
requests as well as those of applications.

Heist takes care of configuring Vault once it can log in to Vault.
But initially you have to configure the following manually:

- Create a Kubernetes Auth Method
- Apply the Heist Policy
- Create a Service Account Binding to authorize
  the Heist Service Account to use the Heist Policy

### Creating the Kubernetes Auth Method

The Kubernetes Auth Method must be mounted on the path `managed/kubernetes`.

```shell
vault auth enable -path=managed/kubernetes kubernetes
```

You can find more information in the official Vault documentation:
[Kubernetes Auth Method](https://www.vaultproject.io/docs/auth/kubernetes)

### Apply Heist Policy

Create a file called `heist.hcl` with the following policy:

```text
path "managed/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "sys/mounts" {
  capabilities = ["read"]
}

path "sys/plugins/reload/backend" {
  capabilities = ["update"]
}

path "sys/tools/random/*" {
  capabilities = ["update"]
}

path "sys/policies/acl/*" {
  capabilities = ["create", "read", "update", "delete"]
}

path "sys/mounts/managed/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "auth/managed/kubernetes/role/*" {
  capabilities = ["create", "read", "update", "delete"]
}
```

Apply the policy with this command:

```shell
vault policy write heist heist.hcl
```

### Create Heist Service Account Binding

The last step is to create a role where the policy is bound to Heists'
Service Account:

```shell
vault write auth/managed/kubernetes/role/heist \
    bound_service_account_names=<HEIST_SERVICE_ACCOUNT> \
    bound_service_account_namespaces=<HEIST_NAMESPACE> \
    policies=heist \
    ttl=1h
```
