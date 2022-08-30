package agent

import (
	"crypto/x509"
	"encoding/pem"
	"sync"
	"time"

	"github.com/youniqx/heist/pkg/vault"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/kvsecret"
	"github.com/youniqx/heist/pkg/vault/pki"
)

const (
	fallbackCertificateExpiryDuration = 168 * time.Hour
	certificateExpiryTTLDivisor       = 2
)

type kvCacheEntry struct {
	ExpiresAt time.Time
	Secret    *kvsecret.KvSecret
}

type certificateCacheEntry struct {
	ExpiresAt   time.Time
	Certificate *pki.Certificate
}

type agentCache struct {
	API              vault.API
	CertificateMutex sync.Mutex
	CertificateCache map[string]*certificateCacheEntry
	KVMutex          sync.Mutex
	KVCache          map[string]*kvCacheEntry
}

func newCache(api vault.API) *agentCache {
	return &agentCache{
		API:              api,
		CertificateCache: make(map[string]*certificateCacheEntry),
		KVCache:          make(map[string]*kvCacheEntry),
	}
}

func (c *agentCache) ReadKvSecret(enginePath core.MountPathEntity, secretPath core.SecretPathEntity) (*kvsecret.KvSecret, error) {
	mountPath, err := enginePath.GetMountPath()
	if err != nil {
		return nil, err
	}

	path, err := secretPath.GetSecretPath()
	if err != nil {
		return nil, err
	}

	cacheKey := mountPath + "|" + path

	c.KVMutex.Lock()
	defer c.KVMutex.Unlock()

	cacheEntry := c.KVCache[cacheKey]
	if cacheEntry != nil && cacheEntry.ExpiresAt.After(time.Now()) {
		return cacheEntry.Secret, nil
	}

	kvSecret, err := c.API.ReadKvSecret(enginePath, secretPath)
	if err != nil {
		return nil, err
	}

	c.KVCache[cacheKey] = &kvCacheEntry{
		ExpiresAt: time.Now().Add(time.Minute),
		Secret:    kvSecret,
	}

	return kvSecret, nil
}

func (c *agentCache) IssueCertificate(enginePath core.MountPathEntity, role core.RoleNameEntity, options *pki.IssueCertOptions) (*pki.Certificate, error) {
	mountPath, err := enginePath.GetMountPath()
	if err != nil {
		return nil, err
	}

	roleName, err := role.GetRoleName()
	if err != nil {
		return nil, err
	}

	cacheKey := mountPath + "|" + roleName

	c.CertificateMutex.Lock()
	defer c.CertificateMutex.Unlock()

	cacheEntry := c.CertificateCache[cacheKey]
	if cacheEntry != nil && cacheEntry.ExpiresAt.After(time.Now()) {
		return cacheEntry.Certificate, nil
	}

	certificate, err := c.API.IssueCertificate(enginePath, role, options)
	if err != nil {
		return nil, err
	}

	var expiresAt time.Time
	block, _ := pem.Decode([]byte(certificate.Certificate))
	switch parsedCert, err := x509.ParseCertificate(block.Bytes); {
	case err != nil:
		expiresAt = time.Now().Add(fallbackCertificateExpiryDuration)
	default:
		duration := time.Until(parsedCert.NotAfter)
		expiresAt = time.Now().Add(duration / certificateExpiryTTLDivisor)
	}

	c.CertificateCache[cacheKey] = &certificateCacheEntry{
		ExpiresAt:   expiresAt,
		Certificate: certificate,
	}

	return certificate, nil
}
