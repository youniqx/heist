# Architectural overview

## Components

### Heist Operator

The Heist Operator manages Vault resources which are defined by [custom
Kubernetes resources](crds). To be able to do this, Heist listens for webhooks
from Kubernetes. It checks created pods in configured namespaces and injects the
Heist Agent as a sidecar and also validates CRDs with a validating webhook.

Access to any secret is strictly controlled and follows the principle of least
privilege. Each service account only gets access to Vault secrets and
functionalities you specifically configure with
[VaultBindings](crds/vaultbinding.md).

### Heist Agent

The Heist Agent is a sidecar container which handles authentication in Vault as
well as providing the necessary secrets to the application pods' shared
volume.

Secrets are provided to the Application Container in form of files in shared
volumes.

Configuration of the Heist Agent is handled by the Heist Operator. The Operator
will generate a VaultClientConfig resource based on configured permissions
with [VaultBindings](crds/vaultbinding.md).
