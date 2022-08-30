package pki

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"path/filepath"
	"strings"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type importCertRequest struct {
	PEMBundle string `json:"pem_bundle,omitempty"`
}

func (p *pkiAPI) ImportCert(ca CAEntity, cert *ImportedCert) error {
	log := p.Core.Log().WithValues("method", "ImportCert")

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	request := &importCertRequest{
		PEMBundle: cert.PrivateKey + "\n" + cert.Certificate,
	}

	var result bytes.Buffer
	if err := p.Core.MakeRequest(core.MethodPost, filepath.Join("/v1", path, "config", "ca"), httpclient.JSON(request), httpclient.Raw(&result)); err != nil {
		log.Info("failed to import cert in pki engine", "error", err)
		return core.ErrAPIError.WithDetails("failed to import cert in pki engine").WithCause(err)
	}

	return nil
}

type parsedImportedCert struct {
	PrivateKey     string
	PrivateKeyType KeyType
	Certificate    *x509.Certificate
}

func parseImportedCert(cert *ImportedCert) (*parsedImportedCert, error) {
	privateKey, keyType, err := parsePrivateKey(cert)
	if err != nil {
		return nil, core.ErrAPIError.WithDetails("failed to parse private key block").WithCause(err)
	}

	publicKeyBlock, _ := pem.Decode([]byte(cert.Certificate))
	if publicKeyBlock == nil {
		return nil, core.ErrAPIError.WithDetails("failed to decode certificate pem")
	}

	certificate, err := x509.ParseCertificate(publicKeyBlock.Bytes)
	if err != nil {
		return nil, core.ErrAPIError.WithDetails("failed to parse certificate data").WithCause(err)
	}

	return &parsedImportedCert{
		PrivateKey:     privateKey,
		PrivateKeyType: keyType,
		Certificate:    certificate,
	}, nil
}

func parsePrivateKey(cert *ImportedCert) (string, KeyType, error) {
	privateKeyBlock, _ := pem.Decode([]byte(cert.PrivateKey))
	if privateKeyBlock == nil {
		return "", "", core.ErrAPIError.WithDetails("failed to decode private key pem")
	}

	keyType, err := determineKeyType(privateKeyBlock)
	if err != nil {
		return "", "", err
	}

	return strings.TrimSpace(cert.PrivateKey), keyType, nil
}

func determineKeyType(block *pem.Block) (KeyType, error) {
	switch block.Type {
	case "RSA PRIVATE KEY":
		return KeyTypeRSA, nil
	case "EC PRIVATE KEY":
		return KeyTypeEC, nil
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", core.ErrAPIError.WithDetails("failed to parse private key").WithCause(err)
		}
		switch key.(type) {
		case *rsa.PrivateKey:
			return KeyTypeRSA, nil
		case *ecdsa.PrivateKey:
			return KeyTypeEC, nil
		default:
			return "", core.ErrAPIError.WithDetails("encountered unknown private key type")
		}
	default:
		return "", core.ErrAPIError.WithDetails("encountered unknown pem block type")
	}
}
