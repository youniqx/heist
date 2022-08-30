# How to sign and verify JWTs with a transit engine

## Introduction

In this tutorial, you will use a transit engine to sign and verify JWTs using
Heist. Then you will sign a JWT with the generated transit engine and verify it.
If you do not have a running instance of Heist yet, you can refer to the [Quick
Start Guide](../quick-start.md).

## Prerequisites

- [Familiarity with Kubernetes
  Concepts](https://www.digitalocean.com/community/tutorials/an-introduction-to-kubernetes)
- A Kubernetes Cluster with heist and HashiCorp Vault installed and configured
  with one of these setup methods:
  - [Production ready cluster, heist and Vault deployment](../deploy.md)
  - [Development setup with kind](./gettings-heist-running-kind.md)
- Familiarity with [JWTs](https://jwt.io/introduction/)
- Vault CLI installed and setup

## Step 1 - Set up a transit engine

First you are going to create a namespace which you can use for this tutorial.
Be aware that you have to grant Heist access if you have a namespaced
deployment.

```bash
kubectl create namespace heist-sign-demo
```

Switching the default `kubectl` context to the new namespace with [kubens](https://github.com/ahmetb/kubectx):

```bash
kubens heist-sign-demo
```

Or switching the default `kubectl` context to the new namespace with`kubectl`
context configuration:

```bash
kubectl config set-context --current --namespace heist-sign-demo
```

You are now set up to configure a transit engine. You can do this by creating
a `VaultTransitEngine` resource.

```bash
kubectl apply -f - <<'EOF'
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultTransitEngine
metadata:
  name: example-transit-engine
EOF
```

Creating this resource will create a transit engine following the naming
scheme `managed/transit_engine/<namespace>/<name>`. Accordingly, you should
now see a transit engine at
`managed/transit_engine/heist-sign-demo/example-transit-engine/`.

You can verify this by either checking it in your Vault UI or by using the
Vault CLI:

```bash
vault secrets list
```

This should look something like this:

```txt
Path                                                              Type         Accessor              Description
----                                                              ----         --------              -----------
cubbyhole/                                                        cubbyhole    cubbyhole_05a2a103    per-token private secret storage
identity/                                                         identity     identity_b5f0593b     identity store
managed/transit/                                                  transit      transit_86969045      n/a
managed/transit_engine/heist-sign-demo/example-transit-engine/    transit      transit_f646b8ca      n/a
```

Don't get confused by `managed/transit/`. This is a transit engine which is
created by Heist itself and lets you encrypt secrets for defining them in Heist
resources which let you hard-code or import these values.

## Step 2 - Configure a transit key

As a transit engine is comparable to a key-ring, which only holds key, you also
have to generate a transit key.

```bash
kubectl create -f - <<'EOF'
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultTransitKey
metadata:
  name: example-transit-key
spec:
  engine: example-transit-engine
  type: ed25519
EOF
```

In this example you created a key of the type `ed25519`. You can find more
information about what key types are supported in the official [Vault
documentation](https://www.vaultproject.io/api/secret/transit#type). Please be
aware that not every key type supports all operations!

Heist then also creates policies for the transit key and all possible operations
of a given key. The policies do not have a completely generic naming schema, but
you can roughly assume that it looks something like this
`managed.<resource_type>.operation.<namespace>.<resource_name>`.

```bash
vault policy list
```

The list should look something like this.

```txt
default
heist
managed.encrypt
managed.transit.key.datakey.heist-sign-demo.example-transit-key
managed.transit.key.decrypt.heist-sign-demo.example-transit-key
managed.transit.key.encrypt.heist-sign-demo.example-transit-key
managed.transit.key.hmac.heist-sign-demo.example-transit-key
managed.transit.key.read.heist-sign-demo.example-transit-key
managed.transit.key.rewrap.heist-sign-demo.example-transit-key
managed.transit.key.sign.heist-sign-demo.example-transit-key
managed.transit.key.verify.heist-sign-demo.example-transit-key
root
```

You can then read the policy which grants access to the `sign` endpoint.

```bash
vault policy read managed.transit.key.sign.heist-sign-demo.example-transit-key
```

You will see that the policy contains all permissions required for using the
`sign` endpoint of the transit key `example-transit-key` in the transit engine
`example-transit-engine`.

```txt
path "managed/transit_engine/heist-sign-demo/example-transit-engine/sign/example-transit-key" {
  capabilities = ["update"]
}
path "managed/transit_engine/heist-sign-demo/example-transit-engine/sign/example-transit-key/sha2-224" {
  capabilities = ["update"]
}
path "managed/transit_engine/heist-sign-demo/example-transit-engine/sign/example-transit-key/sha2-256" {
  capabilities = ["update"]
}
path "managed/transit_engine/heist-sign-demo/example-transit-engine/sign/example-transit-key/sha2-384" {
  capabilities = ["update"]
}
path "managed/transit_engine/heist-sign-demo/example-transit-engine/sign/example-transit-key/sha2-512" {
  capabilities = ["update"]
}
```

## Step 3 - Configure permissions for the pod service account

Later you will start a pod which will run with the namespace's `default` service
account in. Currently, this service account is not able to use the transit key
because you didn't define any permissions. To allow the pod to use the transit
key for signing and verification purposes you have to create a `VaultBinding`
resource. In this example, you will give the `default` service account
permissions to execute the `read`, `sign` and `verify` operations of the transit
key `example-transit-key`.

```bash
kubectl create -f - <<'EOF'
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultBinding
metadata:
  name: example-vault-binding
spec:
  subject:
    name: default
  transitKeys:
    - capabilities:
        - sign
        - read
        - verify
      name: example-transit-key
EOF
```

Heist creates a role in the `managed/kubernetes` authentication method which
will references Vault policies created in
[Step 2](#step-2---configure-a-transit-key).

```bash
vault list auth/managed/kubernetes/role
```

You should now be able to see a role following the name schema
`managed.k8s.<namespace>.<serviceaccount>`. In this example you can see a role
called `managed.k8s.heist-sign-demo.default`.

```txt
Keys
----
heist
managed.k8s.heist-sign-demo.default
```

You then can read the policy for a more detailed overview about what policies
are attached to the role.

```bash
vault read auth/managed/kubernetes/role/managed.k8s.heist-sign-demo.default
```

## Step 4 - Prepare JWT header and payload

Before you can sign the token you have to define its values. As you are using a
`ed25519` key, you have according to
[RFC8037](https://www.rfc-editor.org/rfc/rfc8037), set the `alg` value in the
JWT header to EdDSA.

```json
{
  "alg": "EdDSA",
  "typ": "JWT"
}
```

Afterwards you have to `base64url` encode the minified version of the JSON. If
you don't have `jq`, minify it manually or use an online tool.

```sh
jq -jc . << EOF | basenc --base64url
{
  "alg": "EdDSA",
  "typ": "JWT"
}
EOF
```

Which results in this output:

```txt
eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9
```

Additionally, you need a payload for the token.

```json
{
  "sub": "1234567890",
  "name": "John Doe",
  "admin": true
}
```

Again, you have to `base64url` encode the minified version of the JSON.

```sh
jq -jc . << EOF | basenc --base64url
{
  "sub": "1234567890",
  "name": "John Doe",
  "admin": true
}
EOF
```

Which will give us the following `base64url` encoded output:

```txt
eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9
```

Now you have to combine these values with a `.` and you are ready to sign the
token.

```txt
eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9
```

## Step 5 - Start demo pod

For demonstration purposes you can now start a pod in which you will sign and
verify the JWT.

```bash
kubectl run --image vault:latest heist-demo-sign
```

Then connect into the pod:

```bash
kubectl exec -ti pod/heist-demo-sign -- /bin/sh
```

You have to install two additional tools for this tutorial. First you need
`coreutils` for `base64url` encoding some values and `jq` to deal with some
JSONs. But you can also replace them with any tool you prefer.

```sh
apk add coreutils jq
```

Afterwards you can log in to vault with the service account token in the
container and the role which was created in
[Step 3](#step-3---configure-permissions-for-the-pod-service-account).

```sh
export VAULT_ADDR="http://vault.vault.svc.cluster.local"
vault write auth/managed/kubernetes/login jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" role="managed.k8s.heist-sign-demo.d
efault"
```

## Step 6 - Sign the token

There are multiple ways how you can access the transit engine but here is an
example how you can sign the previously generated token with the Vault CLI. The
`marshaling_algorithm` is required to get a  `base64url` encoded string which
can be  appended to your token to create a valid JWT. If you don't use this flag
or sign it in the UI you have to manually make sure that the value is
`base64url` encoded.

```sh
vault write managed/transit_engine/heist-sign-demo/example-transit-engine/sign/example-transit-key \
  marshaling_algorithm=jws \
  input="$(echo "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9" | basenc --base64 -w 0)"
```

This command will produce an output like this:

```txt
Key            Value
---            -----
key_version    1
signature      vault:v1:iMUhmqr5A8XZ9KhbCgEYNJMbOUzWmNJLQjVkGb47mRrWvu-JeMW49qrJgDLk9KKNDq9gA2JM7w9saRuhadcRAg
```

You now have to strip the leading `vault:v1:` from the signature and append this
signature to your previously generated token and the JWT is valid.

```txt
eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.iMUhmqr5A8XZ9KhbCgEYNJMbOUzWmNJLQjVkGb47mRrWvu-JeMW49qrJgDLk9KKNDq9gA2JM7w9saRuhadcRAg
```

## Step 7 - Verify the token

To verify the JWT you have to split the token at the second `.`. The first part
is the header + payload and the second the signature. Again you should use the
`marshaling_algorithm` argument to tell Vault that the signature is `base64url`
encoded. Otherwise, you have to make sure that you provide a valid `base64`
encoded signature with correct padding.

```sh
vault write managed/transit_engine/heist-sign-demo/example-transit-engine/verify/example-transit-key \
  marshaling_algorithm=jws \
  signature="vault:v1:iMUhmqr5A8XZ9KhbCgEYNJMbOUzWmNJLQjVkGb47mRrWvu-JeMW49qrJgDLk9KKNDq9gA2JM7w9saRuhadcRAg" \
  input="$(echo "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9" | basenc --base64 -w 0)"
```

You should be presented with an output like this:

```txt
Key      Value
---      -----
valid    true
```
