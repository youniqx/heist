# Injector

The injector (also called agent) is used to copy secrets as files into shared
volumes of a pod, so secrets can be mounted into containers.

## Annotations

These annotations can be used to control the injector. They need to be specified
as pod annotations.

Please note that Kubernetes Annotation values are always strings. Thus, booleans
have to be encased in double quotes to mark them as strings.

Following Annotations are configurable:

- `heist.youniqx.com/inject-agent` is used to enable injection of the heist
  agent into a Pod. Defaults to "false".
- `heist.youniqx.com/agent-image` is used to customize the injected agent image.
  If not specified it falls back to the AgentImage set in the injector config. If
  not specified in the config, it will use "youniqx/heist:latest".
- `heist.youniqx.com/agent-preload` is used to customize whether an
  InitContainer is created to make sure the secret is there before the main
  container starts. Defaults to "false".
- `heist.youniqx.com/agent-paths` is used to customize absolute paths where
  secrets can be written to. Multiple paths can be specified by separating them
  with a comma. Default is "/heist/secrets".

The following Annotations are set by the injector:

- `heist.youniqx.com/agent-status` is used by the heist operator to keep track
  of the injection status in Pods. Has the value `injected` if secret is
  injected successful.
