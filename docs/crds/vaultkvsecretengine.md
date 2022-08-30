# VaultKVSecretEngine

Configures a KV secret engine in Vault. Heist only supports version 2 of the KV
secret engine.

Deleting a `VaultKVSecretEngine` will also work if there are still
[**VaultKVSecret**](vaultkvsecret.md) objects storing their data in them.
Restoring the engine in this case will also restore all previous secret values -
this also includes auto generated values.

## Basic Example

Here is a minimal example of a `VaultKVSecretEngine`:

```yaml
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultKVSecretEngine
metadata:
  name: example-kv-engine
```

## Full Example

Here is an example with all fields set to their default value:

```yaml
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultKVSecretEngine
metadata:
  name: example-kv-engine
spec:
  maxVersions: 10
  deleteProtection: false
```

The field `max_versions` configures how many versions of a field should be kept
in Vaults history. This is useful for keeping track of older versions of
secrets, but Heist does not support rollbacks to older version at this time.

Setting `deleteProtection` to `true` prevents the `VaultKVSecretEngine` object
from being deleted from Kubernetes. This may be useful in production
environments.
