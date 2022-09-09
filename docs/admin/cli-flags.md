# Command Line Interface

## Usage

`heist <command> <subcommand> [parameters]`

## Commands, Subcommands and Parameters

| Command          | Parameter                            | Description                                                                                   | Environment Variable               | Type                  | Example               |
|:-----------------|:-------------------------------------|:----------------------------------------------------------------------------------------------|:-----------------------------------|:----------------------|:----------------------|
| `heist operator` |                                      | Starts the Heist Operator.                                                                    |                                    |                       |                       |
|                  | `--leader-elect`                     | Enable leader election for controller manager.                                                | OPERATOR_LEADER_ELECT              | bool                  | true                  |
|                  | `--vault-address`                    | Address of the Vault instance the operator manages.                                           | VAULT_ADDRESS                      | string                | <http://0.0.0.0:1234> |
|                  | `--vault-jwt-path`                   | Path to the file containing the JWT used to authenticate in Vault when using Kubernetes Auth. | VAULT_JWT_PATH                     | string                | path/to/file          |
|                  | `--vault-role`                       | Role used by the operator to authenticate in the Vault instance when using Kubernetes Auth.   | VAULT_ROLE                         | string                | roleName              |
|                  | `--vault-token`                      | Token used by the operator to authenticate in the Vault instance when using Token Auth.       | VAULT_TOKEN                        | string                | vaulttoken            |
|                  | `--vault-ca-cert`                    | CA certs to verify Vault server certificate.                                                  | VAULT_CA_CERTS                     | string                | path/to/file          |
|                  | `--vault-kubernetes-auth-mount-path` | Path of the Kubernetes Auth Engine mounted in Vault used to authenticate in Vault.            | VAULT_KUBERNETES_AUTH_MOUNT_PATH   | string                | path/to/mount         |
|                  | `--metrics-bind-address`             | The address the metric endpoint binds to.                                                     | OPERATOR_METRICS_BIND_ADDRESS      | string                | <http://0.0.0.0:1234> |
|                  | `--health-probe-bind-address`        | The address the probe endpoint binds to.                                                      | OPERATOR_HEALTH_PROBE_BIND_ADDRESS | string                | <http://0.0.0.0:1234> |
|                  | `--webhook-port`                     | The port the webhook server listens on.                                                       | OPERATOR_WEBHOOK_PORT              | string                | 1234                  |
|                  | `--sync-secret-namespace`            | Allow list of namespaces to which values can be synced.                                       | OPERATOR_SYNC_SECRET_NAMESPACE     | list, comma separated | ns1,ns2               |

| Command             | Parameter                   | Description                                                            | Environment Variable          | Type   | Example               |
|:--------------------|:----------------------------|:-----------------------------------------------------------------------|:------------------------------|:-------|:----------------------|
| `heist agent`       |                             | Starts the Heist Agent.                                                |                               |        |                       |
|                     | `--address`                 | Address the agent will be listening on.                                | AGENT_ADDRESS                 | string | <http://0.0.0.0:1234> |
|                     | `--client-config-name`      | Name of the client config object to watch.                             | AGENT_CLIENT_CONFIG_NAME      | string | someObjectName        |
|                     | `--client-config-namespace` | Namespace containing the client config to watch.                       | AGENT_CLIENT_CONFIG_NAMESPACE | string | clientConfigNamespace |
|                     | `--kubernetes-config-path`  | Path to the Kubernetes config file.                                    | AGENT_KUBERNETES_CONFIG_PATH  | string | path/to/config        |
|                     | `--kubernetes-master-url`   | URL of the Kubernetes API server.                                      | AGENT_KUBERNETES_MASTER_URL   | string | <http://0.0.0.0:1234> |
|                     | `--secret-base-path`        | Base path for secrets synced by the agent.                             | AGENT_SECRET_BASE_PATH        | string | path/to/              |
| `heist agent serve` |                             | Starts the Agent server and serve the Agent API at the specified port. |                               |        |                       |
| `heist agent sync`  |                             | Syncs secrets once and then quit.                                      |                               |        |                       |

| Command       | Parameter                  | Description                                                           | Environment Variable           | Type   | Example               |
|:--------------|:---------------------------|:----------------------------------------------------------------------|:-------------------------------|:-------|:----------------------|
| `heist setup` |                            | Configures Vault for use with the Heist Operator.                     |                                |        |                       |
|               | `--heist-service-account`  | Name of the service account used by the Heist Operator.               | SETUP_HEIST_SERVICE_ACCOUNT    | string | heistServiceAccount   |
|               | `--heist-namespace`        | Namespace containing the Heist deployment.                            | SETUP_HEIST_NAMESPACE          | string | heistNamespace        |
|               | `--heist-role-name`        | Name of the role Heist uses to authenticate in Vault.                 | SETUP_HEIST_ROLE_NAME          | string | heistRoleName         |
|               | `--heist-policy-name`      | Name of the policy containing ACL roles for the Heist Operator.       | SETUP_HEIST_POLICY_NAME        | string | heistPolicyName       |
|               | `--vault-token`            | Token used to authenticate in Vault.                                  | SETUP_VAULT_TOKEN              | string | vaulttoken            |
|               | `--vault-ca-cert`          | CA certs to verify Vault server certificate.                          | SETUP_VAULT_CA_CERTS           | string | path/to/file          |
|               | `--vault-scheme`           | Scheme used to connect to Vault (http or https)                       | SETUP_AGENT_ADDRESS            | string | https                 |
|               | `--kubernetes-host`        | Kubernetes API Server Host.                                           | SETUP_KUBERNETES_HOST          | string | <http://0.0.0.0:1234> |
|               | `--kubernetes-jwt-issuer`  | Issuer of service account JWTs in the Kubernetes cluster.             | SETUP_KUBERNETES_JWT_ISSUER    | string | someIssuerName        |
|               | `--kubernetes-jwt-ca-cert` | CA certificate used to validate service account JWTs.                 | SETUP_KUBERNETES_JWT_CA_ISSUER | string | path/to/file          |
|               | `--kubernetes-jwt-pem-key` | One or more keys in PEM format used to validate service account JWTs. | SETUP_KUBERNETES_JWT_PEM_KEYS  | string | path/to/file          |

| Command              | Parameter                  | Description                                                           | Environment Variable           | Type   | Example             |
|:---------------------|:---------------------------|:----------------------------------------------------------------------|:-------------------------------|:-------|:--------------------|
| `heist setup static` |                            | Configures a Vault instance for use with the Heist Operator.          |                                |        |                     |
|                      | `--heist-service-account`  | Name of the service account used by the Heist Operator.               | SETUP_HEIST_SERVICE_ACCOUNT    | string | heistServiceAccount |
|                      | `--heist-namespace`        | Namespace containing the Heist deployment.                            | SETUP_HEIST_NAMESPACE          | string | heistNamespace      |
|                      | `--heist-role-name`        | Name of the role Heist uses to authenticate in Vault.                 | SETUP_HEIST_ROLE_NAME          | string | heistRoleName       |
|                      | `--heist-policy-name`      | Name of the policy containing ACL roles for the Heist Operator.       | SETUP_HEIST_POLICY_NAME        | string | heistPolicyName     |
|                      | `--vault-url`              | URL to the Vault instance you want to configure.                      | SETUP_VAULT_URL                | string | <https://some.url>  |
|                      | `--vault-token`            | Token used to authenticate in Vault.                                  | SETUP_VAULT_TOKEN              | string | vaulttoken          |
|                      | `--vault-ca-cert`          | CA certs to verify Vault server certificate.                          | SETUP_VAULT_CA_CERTS           | string | path/to/file        |
|                      | `--vault-scheme`           | Scheme used to connect to Vault (http or https)                       | SETUP_AGENT_ADDRESS            | string | https               |
|                      | `--kubernetes-jwt-issuer`  | Issuer of service account JWTs in the Kubernetes cluster.             | SETUP_KUBERNETES_JWT_ISSUER    | string | someIssuerName      |
|                      | `--kubernetes-jwt-ca-cert` | CA certificate used to validate service account JWTs.                 | SETUP_KUBERNETES_JWT_CA_ISSUER | string | path/to/file        |
|                      | `--kubernetes-jwt-pem-key` | One or more keys in PEM format used to validate service account JWTs. | SETUP_KUBERNETES_JWT_PEM_KEYS  | string | path/to/file        |

| Command           | Parameter                  | Description                                                              | Environment Variable           | Type   | Example               |
|:------------------|:---------------------------|:-------------------------------------------------------------------------|:-------------------------------|:-------|:----------------------|
| `heist setup k8s` |                            | Configures an in-cluster Vault instance for use with the Heist Operator. |                                |        |                       |
|                   | `--heist-service-account`  | Name of the service account used by the Heist Operator.                  | SETUP_HEIST_SERVICE_ACCOUNT    | string | heistServiceAccount   |
|                   | `--heist-namespace`        | Namespace containing the Heist deployment.                               | SETUP_HEIST_NAMESPACE          | string | heistNamespace        |
|                   | `--heist-role-name`        | Name of the role Heist uses to authenticate in Vault.                    | SETUP_HEIST_ROLE_NAME          | string | heistRoleName         |
|                   | `--heist-policy-name`      | Name of the policy containing ACL roles for the Heist Operator.          | SETUP_HEIST_POLICY_NAME        | string | heistPolicyName       |
|                   | `--vault-url`              | URL to the Vault instance you want to configure.                         | SETUP_VAULT_URL                | string | <https://some.url>    |
|                   | `--vault-token`            | Token used to authenticate in Vault.                                     | SETUP_VAULT_TOKEN              | string | vaulttoken            |
|                   | `--vault-ca-cert`          | CA certs to verify Vault server certificate.                             | SETUP_VAULT_CA_CERTS           | string | path/to/file          |
|                   | `--vault-scheme`           | Scheme used to connect to Vault (http or https)                          | SETUP_AGENT_ADDRESS            | string | https                 |
|                   | `--kubernetes-host`        | Kubernetes API Server Host.                                              | SETUP_KUBERNETES_HOST          | string | <http://0.0.0.0:1234> |
|                   | `--kubernetes-jwt-issuer`  | Issuer of service account JWTs in the Kubernetes cluster.                | SETUP_KUBERNETES_JWT_ISSUER    | string | someIssuerName        |
|                   | `--kubernetes-jwt-ca-cert` | CA certificate used to validate service account JWTs.                    | SETUP_KUBERNETES_JWT_CA_ISSUER | string | path/to/file          |
|                   | `--kubernetes-jwt-pem-key` | One or more keys in PEM format used to validate service account JWTs.    | SETUP_KUBERNETES_JWT_PEM_KEYS  | string | path/to/file          |

## Completion

Generate completion code:
`heist completion [bash|zsh|fish|powershell]`

### Bash

#### Linux

Generate completion script as described above and
put the generated completion-file in the completions subdirectory of
`$BASH_COMPLETION_USER_DIR`(defaults to `$XDG_DATA_HOME/bash-completion`
or `~/.local/share/bash-completion` if `$XDG_DATA_HOME` is not set)
to have them loaded automatically on demand when the respective command
is being completed.

You can find more info on where to put your completion files
[here](https://github.com/scop/bash-completion/blob/master/README.md#faq).

#### macOS

Execute the following command which generates the completion script and puts it
in `/usr/local/etc/bash_completion.d/heist`

`$ heist completion bash > /usr/local/etc/bash_completion.d/heist`

### Zsh

If shell completion is not already enabled in your environment
you will need to enable it. You can execute the following command once:

`$ echo "autoload -U compinit; compinit" >> ~/.zshrc`

To load completions for each session, execute once:

`$ heist completion zsh > "${fpath[1]}/_heist"`

You will need to start a new shell for this setup to take effect.

### fish

`$ heist completion fish | source`

To load completions for each session, execute once:

`$ heist completion fish > ~/.config/fish/completions/heist.fish`

### PowerShell

`PS> heist completion powershell | Out-String | Invoke-Expression`

To load completions for every new session, run

`PS> heist completion powershell > heist.ps1`

and source this file from your PowerShell profile.
