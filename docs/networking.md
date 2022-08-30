# Networking

## Heist Operator

The Heist operator needs to be able to send requests to the Kubernetes API and
Vault instance. Additionally, the container listens on port 9443.

## Heist Agent Sidecar

The Heist sidecar containers need to be able to talk to
the Heist operator and the Kubernetes API.
The sidecar sends Heist requests to the k8s service that does
load balancing and port redirection for Heist, which is named `heist-webhook` if
you use our chart.
