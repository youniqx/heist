# Heist Documentation

## CRD Overview

At the moment Heist can manage KV and PKI secret engines. There are 4 CRDs
related to managing those engines:

- [**VaultKVSecretEngine**](crds/vaultkvsecretengine.md): Creates a KV secret
  engine in Vault
- [**VaultKVSecret**](crds/vaultkvsecret.md): Creates a KV secret in a secret
  engine created with [**VaultKVSecretEngine**](crds/vaultkvsecretengine.md).
- [**VaultCertificateAuthority**](crds/vaultcertificateauthority.md): Creates
  a new PKI engine, can import or auto generate certificate and keys.
- [**VaultCertificateRole**](crds/vaultcertificaterole.md): Creates a
  Certificate role in a PKI engine created with
  [**VaultCertificateAuthority**](crds/vaultcertificateauthority.md) to enable
  issuing certificates.

Access management is controlled with the
[**VaultBinding**](crds/vaultbinding.md) CRD. When you use one of the above
four CRDs Heist creates policies in Vault which grant access to those
resources. The [**VaultBinding**](crds/vaultbinding.md) CRD then binds
those policies to the listed service accounts.

There also is the [**VaultSyncSecret**](crds/vaultsyncsecret.md) CRD which can
be used to sync values from Vault to Kubernetes Secrets. This is useful for
things like image pull secrets, where storing the value in a Kubernetes Secret
is mandatory. This is also useful for TLS secrets. If you create a
[**VaultCertificateRole**](crds/vaultcertificaterole.md) and configure it for
TLS, then Heist can sync the keys to a secret and will update them automatically
before they expire. This allows you to use short-lived TLS credentials based on
your own PKI.

The `VaultClientConfig` CRD is completely managed by Heist. Users should not
create `VaultClientConfig` objects themselves. They contain configuration for
pods and service accounts which is consumed by the Heist Agent.

## Heist Agent

The Heist Agent is a sidecar container Heist can inject into your pods. It will
automatically inject the secrets configured in the [**VaultBinding**](crds/vaultbinding.md)
CRD into the Pod based on the service account the Pod uses.

The Heist Agent is not injected per default. To enable this for a deployment
add the `younix.com/heist-agent-enabled` annotation to the Pod template with
the value `"true"`.
