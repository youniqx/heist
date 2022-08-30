# Authorization in Vault

As explained in [Authentication in Vault](authentication-in-vault.md) Heist is
using Kubernetes authentication in Vault. This means that we can authorize
Kubernetes subject to interact with Vault resources.

By default, Heist does not give any Kubernetes subject access to any Vault
resource. To manage access to Vault resources, you have to configure
[capabilities](../crds/vaultbinding.md#capabilities) in a
[VaultBinding](../crds/vaultbinding.md). This will create a Kubernetes Auth Role
which binds the required Vault policies to the specified Kubernetes subject.

## Vault policies

Whenever a Heist resource is detected by Heist, the resource gets created with
all applicable Vault policies. For example, if a VaultKVSecret is created Heist will
also create a policy for each [capability available for VaultKVSecrets](../crds/vaultbinding.md#speckvsecrets)
for this specific VaultKVSecret. Vault policies managed by Heist are
prefixed with `managed.`. The naming schema afterwards is different for each
resource type (transit key, PKI, etc.) but the naming is done in a way that it
describes the affected resource and the allowed operation.
