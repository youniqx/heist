package vaultsyncsecret

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/managed"
	"github.com/youniqx/heist/pkg/vault"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/pki"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	renewDisabled      time.Duration = 0
	certRenewFrequency time.Duration = 2
)

var (
	ErrInvalidField     = errors.New("invalid-field")
	ErrTemplateNotFound = errors.New("template-not-found")
	ErrAlreadyOwned     = errors.New("already-owned")
)

type dataFetcher struct {
	Context       context.Context
	Client        client.Client
	VaultAPI      vault.API
	Spec          *heistv1alpha1.VaultSyncSecretSpec
	Data          map[string][]byte
	RenewInterval time.Duration
	CertMap       map[int]*pki.Certificate
	SyncSecret    *heistv1alpha1.VaultSyncSecret
}

func (r *Reconciler) FetchData(ctx context.Context, sync *heistv1alpha1.VaultSyncSecret) (time.Duration, map[string][]byte, error) {
	fetcher := &dataFetcher{
		Context:       ctx,
		Client:        r.Client,
		VaultAPI:      r.VaultAPI,
		SyncSecret:    sync,
		Spec:          &sync.Spec,
		Data:          make(map[string][]byte),
		RenewInterval: renewDisabled,
		CertMap:       make(map[int]*pki.Certificate),
	}

	if err := fetcher.FetchData(); err != nil {
		return renewDisabled, nil, err
	}

	return fetcher.RenewInterval, fetcher.Data, nil
}

func (d *dataFetcher) FetchData() error {
	var err error
	for key, source := range d.Spec.Data {
		switch {
		case source.CipherText != "":
			if d.Data[key], err = d.VaultAPI.TransitDecrypt(managed.TransitEngine, managed.TransitKey, string(source.CipherText)); err != nil {
				return err
			}
		case source.CertificateAuthority != nil:
			if d.Data[key], err = d.FetchCertificateAuthority(source.CertificateAuthority); err != nil {
				return err
			}
		case source.Certificate != nil:
			if d.Data[key], err = d.FetchCertificate(source.Certificate); err != nil {
				return err
			}
		case source.KVSecret != nil:
			if d.Data[key], err = d.FetchKvSecret(source.KVSecret); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *dataFetcher) FetchCertificateAuthority(authority *heistv1alpha1.VaultSyncCertificateAuthoritySource) ([]byte, error) {
	ca := &heistv1alpha1.VaultCertificateAuthority{
		ObjectMeta: metav1.ObjectMeta{
			Name:      authority.Name,
			Namespace: d.SyncSecret.Namespace,
		},
	}
	if err := d.Client.Get(d.Context, client.ObjectKeyFromObject(ca), ca); err != nil {
		return nil, err
	}

	if meta.IsStatusConditionFalse(ca.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		return nil, fmt.Errorf("CertificateAuthority not provisioned yet")
	}

	publicSecret, err := d.VaultAPI.ReadKvSecret(common.InternalKvEngine, core.SecretPath(common.GetCAInfoSecretPath(ca)))
	if err != nil {
		return nil, err
	}
	privateSecret, err := d.VaultAPI.ReadKvSecret(common.InternalKvEngine, core.SecretPath(common.GetCAPrivateKeySecretPath(ca)))
	if err != nil {
		return nil, err
	}

	var value string
	switch authority.Field {
	case heistv1alpha1.VaultBindingCertificateFieldTypeCertChain:
		value = publicSecret.Fields[common.CACertificateChainField]
	case heistv1alpha1.VaultBindingCertificateFieldTypeFullCertChain:
		value = publicSecret.Fields[common.CACertificateFullChainField]
	case heistv1alpha1.VaultBindingCertificateFieldTypePrivateKey:
		value = privateSecret.Fields[common.CAPrivateKeyField]
	case heistv1alpha1.VaultBindingCertificateFieldTypeCertificate:
		value = publicSecret.Fields[common.CACertificateField]
	default:
		return nil, ErrInvalidField
	}

	return []byte(value), nil
}

//nolint:cyclop
func (d *dataFetcher) FetchCertificate(certificate *heistv1alpha1.VaultSyncCertificateSource) ([]byte, error) {
	_, tpl, err := d.findTemplateAndIndex(certificate)
	if err != nil {
		return nil, err
	}

	cert := &heistv1alpha1.VaultCertificateRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tpl.CertificateRole,
			Namespace: d.SyncSecret.Namespace,
		},
	}
	if err := d.Client.Get(d.Context, client.ObjectKeyFromObject(cert), cert); err != nil {
		return nil, err
	}

	if meta.IsStatusConditionFalse(cert.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		return nil, fmt.Errorf("certificate not provisioned yet")
	}

	ca := &heistv1alpha1.VaultCertificateAuthority{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cert.Spec.Issuer,
			Namespace: d.SyncSecret.Namespace,
		},
	}
	if err := d.Client.Get(d.Context, client.ObjectKeyFromObject(ca), ca); err != nil {
		return nil, err
	}

	if meta.IsStatusConditionFalse(ca.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		return nil, fmt.Errorf("issuer not provisioned yet")
	}

	rootCA, err := common.FindRootCA(d.Context, d.Client, ca)
	if err != nil {
		return nil, err
	}

	if meta.IsStatusConditionFalse(rootCA.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		return nil, fmt.Errorf("root authority not provisioned yet")
	}

	rootPEM, err := d.VaultAPI.ReadCACertificatePEM(rootCA)
	if err != nil {
		return nil, err
	}

	roleName, err := cert.GetRoleName()
	if err != nil {
		return nil, err
	}

	issuedCert, err := d.getIssuedCert(certificate, ca, roleName)
	if err != nil {
		return nil, err
	}

	certRenewalInterval := cert.Spec.Settings.TTL.Duration / certRenewFrequency
	if d.RenewInterval == renewDisabled || d.RenewInterval > certRenewalInterval {
		d.RenewInterval = certRenewalInterval
	}

	chain := strings.Join(issuedCert.CAChain, "\n")

	var value string
	switch certificate.Field {
	case heistv1alpha1.VaultBindingCertificateFieldTypeCertChain:
		value = strings.TrimSpace(chain)
	case heistv1alpha1.VaultBindingCertificateFieldTypeFullCertChain:
		value = strings.TrimSpace(chain + "\n" + rootPEM)
	case heistv1alpha1.VaultBindingCertificateFieldTypePrivateKey:
		value = issuedCert.PrivateKey
	case heistv1alpha1.VaultBindingCertificateFieldTypeCertificate:
		value = issuedCert.Certificate
	default:
		return nil, ErrInvalidField
	}

	return []byte(value), nil
}

func (d *dataFetcher) findTemplateAndIndex(certificate *heistv1alpha1.VaultSyncCertificateSource) (int, *heistv1alpha1.VaultCertificateTemplate, error) {
	for index, template := range d.Spec.CertificateTemplates {
		if template.Alias != "" {
			if template.Alias == certificate.Name {
				return index, template.DeepCopy(), nil
			}
		} else if template.CertificateRole == certificate.Name {
			return index, template.DeepCopy(), nil
		}
	}

	return -1, nil, ErrTemplateNotFound
}

func (d *dataFetcher) getIssuedCert(certificate *heistv1alpha1.VaultSyncCertificateSource, ca *heistv1alpha1.VaultCertificateAuthority, roleName string) (*pki.Certificate, error) {
	index, template, err := d.findTemplateAndIndex(certificate)
	if err != nil {
		return nil, err
	}

	existingCertificate := d.CertMap[index]
	if existingCertificate != nil {
		return existingCertificate, nil
	}

	issuedCert, err := d.VaultAPI.IssueCertificate(ca, core.RoleName(roleName), &pki.IssueCertOptions{
		CommonName:        template.CommonName,
		DNSSans:           template.DNSSans,
		OtherSans:         template.OtherSans,
		IPSans:            template.IPSans,
		URISans:           template.URISans,
		TTL:               template.TTL.Duration,
		ExcludeCNFromSans: template.ExcludeCNFromSans,
	})
	if err != nil {
		return nil, err
	}

	d.CertMap[index] = issuedCert

	return issuedCert, nil
}

func (d *dataFetcher) FetchKvSecret(secret *heistv1alpha1.VaultSyncKVSecretSource) ([]byte, error) {
	kvSecret := &heistv1alpha1.VaultKVSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: d.SyncSecret.Namespace,
		},
	}
	if err := d.Client.Get(d.Context, client.ObjectKeyFromObject(kvSecret), kvSecret); err != nil {
		return nil, err
	}

	if meta.IsStatusConditionFalse(kvSecret.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		return nil, fmt.Errorf("kv secret not provisioned yet")
	}

	kvEngine := &heistv1alpha1.VaultKVSecretEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kvSecret.Spec.Engine,
			Namespace: d.SyncSecret.Namespace,
		},
	}
	if err := d.Client.Get(d.Context, client.ObjectKeyFromObject(kvEngine), kvEngine); err != nil {
		return nil, err
	}

	if meta.IsStatusConditionFalse(kvEngine.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		return nil, fmt.Errorf("kv engine not provisioned yet")
	}

	secretData, err := d.VaultAPI.ReadKvSecret(kvEngine, kvSecret)
	if err != nil {
		return nil, err
	}

	return []byte(secretData.Fields[secret.Field]), nil
}
