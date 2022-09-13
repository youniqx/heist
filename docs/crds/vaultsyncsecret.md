# VaultSyncSecret

Enables users to sync secrets from Vault to a Kubernetes Secret. Useful for
provisioning things like image pull secrets.

It is possible to sync secrets to other namespaces. However, the target
namespace has to be in an allow-list configured for Heist. The allow list can
be configured in the Operator with the environment variable
`OPERATOR_SYNC_SECRET_NAMESPACE_ALLOW_LIST` or the flag
`--sync-secret-namespace`.

For security reasons, Heist will refuse to sync values to already existing
secrets. The secret has to be new and cannot contain any non Heist managed
values.

## Basic example

Here is an example of a `VaultSyncSecret` resource:

```yaml
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultSyncSecret
spec:
  data:
    example-certificate-authority-cert:
        certificateAuthority:
          name: existing-certificate-authority
          field: certificate
    example-certificate-cert:
        certificate:
          name: existing-certificate
          field: certificate
    example-kv-secret-field-1:
        kvSecret:
          name: existing-kv-secret
          field: field_1
    example-ciphertext:
        cipherText: vault:v1:uu8YASXKaP0f62+vDS4Z0m3iv3tW0GhSiI8wU8iNU0A=
  target:
    name: example-kubernetes-secret
    namespace: example-namespace
    type: Opaque
    additionalAnnotations:
      youniqx.com/test-label: "true"
    additionalLabels:
      youniqx.com/test-label: "true"
```

## Secret Types

Heist can provision secrets of any type. However, users will have to make sure
that restrictions placed on those other secret types - like a TLS secret having
`tls.key` and `tls.crt` values - are met.

The supported secret types are:

- `Opaque`
- `kubernetes.io/tls`
- `kubernetes.io/dockercfg`
- `kubernetes.io/dockerconfigjson`
- `kubernetes.io/ssh-auth`
- `kubernetes.io/basic-auth`

Setting the secret type can be done using the `type` field under `secret`:

```yaml
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultSyncSecret
secret:
  name: example-secret
  type: kubernetes.io/tls
```

## Syncing secrets to other namespaces

Syncing secrets to the current namespace is generally possible, as long as the
target secret does not exist beforehand.

For some use cases it is required to place secrets in other namespaces - for
example TLS secrets. For that purpose admins can configure a global allow-list
of namespaces for Heist in which it is possible to sync secrets from every
namespace.

In that case users have to take care of not causing conflicts by multiple
namespaces trying to set the same secret. A simple solution would be prefixing
secret names with the source namespace.

To sync secrets to another namespace on the allow list, just set the `namespace`
field under `secret`:

```yaml
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultSyncSecret
secret:
  name: example-secret
  namespace: some-other-namespace
```

## Full example

Here is an example with all fields set to their default value:

```yaml
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultSyncSecret
secret:
  name: example-secret
  namespace: ""
  type: Opaque
kv: []
certificateAuthorities: []
certificates: []
```

Default values for object in `kv`, `certificateAuthorities` and `certificates`
are:

```yaml
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultSyncSecret
secret:
  name: example-secret
kv:
  - name: ""
    field: ""
    key: ""
certificateAuthorities:
  - name: ""
    fields:
      - field: certificate
        key: ""
certificates:
  - name: example-certificate
    common_name: ""
    exclude_cn_from_sans: false
    alternative_names: []
    fields:
      - field: certificate
        key: ""
```
