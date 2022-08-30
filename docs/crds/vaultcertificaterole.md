# VaultCertificateRole

Configures a certificate role in a PKI Engine. Requires defining a
`VaultCertificateAuthority`.

## Basic Example

Here is a minimal example:

```yaml
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultCertificateRole
metadata:
  name: example-certificate
spec:
  issuer: example-certificate-authority
  settings:
    keyBits: 4096
    keyType: rsa
  subject:
    organization: [ Example LLC ]
```

## Full Example

Here is an example with all fields set to their default values:

```yaml
apiVersion: heist.youniqx.com/v1alpha1
kind: VaultCertificateRole
metadata:
  name: example-certificate
spec:
  issuer: example-certificate-authority
  subject:
    country: []
    locality: []
    ou: []
    organization: []
    postalCode: []
    province: []
    streetAddress: []
  settings:
    allowBareDomains: false
    allowedDomains: []
    allowAnyName: false
    enforceHostNames: false
    allowGlobDomains: false
    allowSubdomains: false
    requireCN: false
    ttl: ""
    keyBits: 2048
    keyType: rsa
    allowIPSans: false
    allowLocalhost: false
    allowedDomainsTemplate: false
    allowedOtherSans: []
    allowedURISans: []
    basicConstraintsValidForNonCA: false
    clientFlag: false
    codeSigningFlag: false
    emailProtectionFlag: false
    extKeyUsage: []
    extKeyUsageOIDS: []
    keyUsage: []
    maxTTL: ""
    notBeforeDuration: ""
    policyIdentifiers: []
    serverFlag: false
    useCSRCommonName: false
    useCSRSans: false
```

The configuration under `settings` maps directly to the values you can configure
in the Vault API. Refer to here for more information:
https://www.vaultproject.io/api/secret/pki#create-update-role
