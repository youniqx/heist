# VaultBinding

`VaultBinding` is used to grant a service account access to any Heist managed
object. This currently includes:

- Certificate Authorities
- Certificate Roles
- KV Secrets
- Transit Keys
- Heist's Default Transit Engine

Any path specified in the `VaultBinding` is configuration for the Heist Agent.
Relative paths are assumed to be relative to the folder `/heist/secrets`.

To provision secrets outside the default `/heist/secrets` folder you have to
configure a pod annotation. Please refer to the
[injector documentation](../admin/injector.md).

## Capabilities

Capabilities bind permissions to the configured subject. Most of the time this
is the service account of the pod. You have to configure them in order to use
the Heist Agent to inject them or to allow direct access from your application
to bound secrets.

### spec.capabilities

| capability | description                                                                                               |
| ---------- | --------------------------------------------------------------------------------------------------------- |
| encrypt    | Allows the service account to use the default managed transit engine `managed.encrypt` to encrypt values. |

### spec.certificateAuthorities

| capability   | description                                                                      |
| ------------ | -------------------------------------------------------------------------------- |
| read_public  | Allows the service account to read the public key of the certificate authority.  |
| read_private | Allows the service account to read the private key of the certificate authority. |

### spec.certificateRoles

| capability    | description                                                                                                                                                                                                                                                                               |
| ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| issue         | Configures the certificate to be able to issue other certificates.                                                                                                                                                                                                                        |
| sign_csr      | Configures the Certificate to be able to sign CSRs, using the fields as configured in the VaultCertificateAuthority.                                                                                                                                                                      |
| sign_verbatim | Configures the Certificate to be able to sign CSRs, using the fields provided by the CSRs. Generally speaking it is safer to use the capability sign_csr, since verbatim signing requires validation that the provided values are correct, since no validation is done by Vault or Heist. |

### spec.kvSecrets

| capability | description                                       |
| ---------- | ------------------------------------------------- |
| read       | Allows the service account to read the KV secret. |

### spec.transitKeys

| capability | description                                                                                                                                                                                                                                                                                                                                                                |
| ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| encrypt    | Allows the service account to use the transit key to encrypt data.                                                                                                                                                                                                                                                                                                         |
| decrypt    | Allows the service account to use the transit key to decrypt data.                                                                                                                                                                                                                                                                                                         |
| datakey    | Allows the service account to use the transit key to use a data key that can be used for offline de- and encryption. The data key is NOT the transit key used when encrypting or decrypting values with the API. Vault provides an example [use case](https://learn.hashicorp.com/tutorials/vault/eaas-transit#generate-data-key) with a tutorial on how to use data keys. |
| rewrap     | Allows the service account to use the transit key to rewrap an already encrypted secret with the latest version of the encryption key.                                                                                                                                                                                                                                     |
| sign       | Allows the service account to use the transit key to sign data.                                                                                                                                                                                                                                                                                                            |
| hmac       | Allows the service account to use the transit key to generate a digest of the provided data and key.                                                                                                                                                                                                                                                                       |
| verify     | Allows the service account to use the transit key to verify signed data.                                                                                                                                                                                                                                                                                                   |
| read       | Allows the service account to use the transit key to retrieve information about the transit key. The transit key itself is not exposed via the API.                                                                                                                                                                                                                        |

## Templating

Once you have bound your secrets to a subject, you can configure Heist Agent to
inject bound secrets into the pod (if applicable). The template supports
[sprig](https://masterminds.github.io/sprig/) template functions and can access
all bound secrets if all required [capabilities](#Capabilities)
have been configured correctly.

It follows the convention `{{ <secretType> <name> <field> }}`.

### caField

caField can be used to reference a certificate authority and to inject certain
fields of it. The following fields are available.

| field           | description                                                                          |
| --------------- | ------------------------------------------------------------------------------------ |
| cert_chain      | Is the field type for binding the cert chain of a certificate.                       |
| full_cert_chain | Is the field type for binding the full cert chain (including root) of a certificate. |
| private_key     | Is the field type for binding the private key of a certificate.                      |
| certificate     | Is the field type for binding the public part a certificate.                         |

`{{ caField "example" "full_cert_chain" }}`: retrieves the value of field
"full_cert_chain" from CA "example".

### certField

certField can be used to reference a certificate and to inject certain
fields of it. The following fields are available.

| field           | description                                                                          |
| --------------- | ------------------------------------------------------------------------------------ |
| cert_chain      | Is the field type for binding the cert chain of a certificate.                       |
| full_cert_chain | Is the field type for binding the full cert chain (including root) of a certificate. |
| private_key     | Is the field type for binding the private key of a certificate.                      |
| certificate     | Is the field type for binding the public part a certificate.                         |

`{{ certField "example" "full_cert_chain" }}`: retrieves the value of field
"full_cert_chain" from CA "example".

### kvSecret

kvSecret can be used to reference a VaultKVSecret and to inject certain
values of it. You can directly use the name of the key in your KV secret.

`{{ kvSecret "secret_1" "user" }}` retrieves the value of key "user" in secret_1
`{{ kvSecret "secret_1" "pass" }}` retrieves the value of key "pass" in secret_1

## Example

```yaml
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultBinding
metadata:
  name: example-binding
spec:
  subject:
    name: example-service-account
  certificateAuthorities:
    - name: example-vault-certificate-authority
      capabilities:
        - read_public
        - read_private
  certificateRoles:
    - name: example-vault-certificate
      capabilities:
        - issue
  kvSecrets:
    - name: example-kv-secret
      capabilities:
        - read
  capabilities:
    - encrypt
  transitKeys:
    - name: example-transit-key
      capabilities:
        - encrypt
        - sign
  agent:
    certificateTemplates:
      - certificateRole: example-vault-certificate
        commonName: example.com
        excludeCNFromSans: true
        otherSans:
          - "*.example.com"
    templates:
      - path: tls.key
        template: '{{ certField "example-vault-certificate" "private_key" }}'
      - path: tls.crt
        template: '{{ certField "example-vault-certificate" "certificate" }}'
      - path: cachain.crt
        template: '{{ certField "example-vault-certificate" "cert_chain" }}'
      - path: ca_cert.pem
        template: '{{ caField "example-vault-certificate-authority" "certificate" }}'
      - path: ca_key.pem
        template: '{{ caField "example-vault-certificate-authority" "private_key" }}'
      - path: example-kv-secret
        template: '{{ kvSecret "example-kv-secret" "field_name_1" }}'
```
