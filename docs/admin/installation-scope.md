# Installation scope

Operators need access to resources in order to manage and monitor them. To do
so, you can install Heist either namespaced or globally in your cluster. The
benefit of a namespaced deployment is, that Heist only has access to a specified
set of namespaces. The disadvantage is, that you manually have to allow
namespaces in order to use Heist.

## Global deployment

By default, Heist will be installed globally. This is not the recommended way for
production, but it makes it easier to get started with Heist. This will create
ClusterRoles and ClusterRoleBindings which will give your Heist service account
cluster wide access to all Heist resources.

> The deployed ClusterRole includes permissions to manage Roles and
> RoleBindings. This is required because Heist restricts access to
> ClientConfigs (Heist Agent configuration) to the service account used by the
> pod.

```bash
helm upgrade --install heist youniqx-oss/heist \
  --set vault.address="<vault_endpoint>"
```

## Namespaced deployment

### Watch specific namespaces

Heist can be configured to only watch a specific set of namespaces. To do this,
configure the `scope.watchedNamespaces` value in the helm chart. When
configured, the Helm chart will create ClusterRoles and ClusterRoleBindings
which will give your Heist service account cluster wide access to all Heist
resources. Additionally, Heist will only watch those configured namespaces.

> The deployed ClusterRole includes permissions to manage Roles and
> RoleBindings. This is required because Heist restricts access to
> ClientConfigs (Heist Agent configuration) to the service account used by the
> pod.

```bash
helm upgrade --install heist youniqx-oss/heist \
  --set vault.address="<vault_endpoint>" \
  --set scope.watchedNamespaces="example-ns-1,example-ns-2"
```

### Remove global permissions entirely

If you don't want Heist to have global permissions in your cluster, you can
disable it by setting `managerRoles.enabled` to false. This prevents deployment
of all RBAC resource required by Heist. You then have to manually deploy
RBAC resources to required namespaces in order for Heist to work.

```bash
helm upgrade --install heist youniqx-oss/heist \
  --set vault.address="<vault_endpoint>" \
  --set managerRoles.enabled=false
```

### Manage single namespace

There is also a third method with which you can deploy Heist without global
permissions but still use the provided RBAC resources. When setting
`scope.global` to false, the Helm chart will create Roles instead of
ClusterRoles which gives Heist only access to the namespace where it is deployed
to.

```bash
helm upgrade --install heist youniqx-oss/heist \
  --set vault.address="<vault_endpoint>" \
  --set scope.global=false
```
