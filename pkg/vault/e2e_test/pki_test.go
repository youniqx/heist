package e2e_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/mount"
	"github.com/youniqx/heist/pkg/vault/pki"
)

var (
	rootPrivateKey  = "-----BEGIN PRIVATE KEY-----\n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\nUhpoCUNxZwiOrQqP+6Z1ZqVA9Tiorw==\n-----END PRIVATE KEY-----"
	rootCertificate = "-----BEGIN CERTIFICATE-----\nMIIFkzCCA3ugAwIBAgIUOa+M/6+UJ19Jf5GZzK0uzKO/FdcwDQYJKoZIhvcNAQEL\nBQAwWTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxHDAaBgNVBAoM\nE3lvdW5pcXggSWRlbnRpdHkgQUcxFzAVBgNVBAMMDlVuaXQgVGVzdCBSb290MB4X\nDTIxMDgxODA5MzI0OVoXDTMxMDgxNjA5MzI0OVowWTELMAkGA1UEBhMCQVUxEzAR\nBgNVBAgMClNvbWUtU3RhdGUxHDAaBgNVBAoME3lvdW5pcXggSWRlbnRpdHkgQUcx\nFzAVBgNVBAMMDlVuaXQgVGVzdCBSb290MIICIjANBgkqhkiG9w0BAQEFAAOCAg8A\nMIICCgKCAgEAzgzPL5x6N3+VD7G0oqub9l7I8dkKcX4SyH7TturDWbSHwRufN8wN\n9mZMDcLGh4xV/i6WiYste16qAnes4CC1cBX+seJiO3H4BzLtixXiNbNd0uzvFSIB\nAONxPuHWNW/LD+A7kbKGUEgVe6JmLH0XkkpzsAOwlbl6Abi0q6+8kfE35HffZce0\nYc5Vi5nIMFLurRRiO0PyE7SfLBqGzhI3oZqyHs+ce/nXxL5lUMImWNGoUHjjtoY5\nNeVWXQMk1NMF/KQGqtU2mIRq38w5lADZ80u6JCqMxxjJCwvLgNvvNZVXs0Xu1KD2\nFhlXXwxzgdkyDzHjGptIMkUr27Sakj3L2civ/gPAQTT/zhMnEnOB6hyHpJ8dehpB\nqF07vSrfFAH2OE8p31yTqIfh9U3SRtybOmgZtLWzyDaAHDtbqRc2+DJ/eIbgR+vk\nFOonUU5HGANu0mBFbPoRsJF4M7J4v7+bReexBYPlxQrdb75XWoOSU67GkuByDy7w\nUBxioBLFcT7lVPx46E3E8be5B+OAARYw3BgzNYfCdkAQi5a4EbYT7yGUy5s+AETl\nhgICicxyS9HkvBkUZwZNPjO9rEbGdBRv7FMpi5KuoTXRUcoPUh7Bx8A7BQItrMv+\n/Am32WnYMhbp5RLT8zi2qPIDoe1qOzETsYLIHoUOzXC05czdjQRiXgkCAwEAAaNT\nMFEwHQYDVR0OBBYEFHyfSI5wGHiw2dyat3GsW6v0pOLnMB8GA1UdIwQYMBaAFHyf\nSI5wGHiw2dyat3GsW6v0pOLnMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQEL\nBQADggIBAAkETo1D0GTw21Dl1xSN1eZpT+S3cK2oPgdJtFinnT36B94VutwxNHjv\nueq8Y92EOTUztkD513bZonWPuBn7oOZMe03UF62hkRUoClGgqHYO7EcPB8YV1vfF\nokD4AJyPAiyYO0/Yv3WNpQ0su1hKvjPyp7FQJTsOtMXAPeI3M/kmc98pqE2gAjkZ\nsVnncseBNB6av8+m8j3dDjc5Zzj5jtG1eDNv1JaO8RDfktBj8M/ezqR3c1H1c/uj\nANdZUAhNDWqSEiYU3qV4g65lYbPHiPr917yrIJrOFBz2Uen9C2UE9saDTG71Jmkc\nf+i1ZR1mia9OtZhpET5EKk6fuegOyzcWpMQTWGYuX4XIL+UtoLExRUrqjbhgN2Ya\n+lsX5XjyAF5dkqWpc/i4DPBNRMBXoDnW/vwO898UJFRHRq/hCB641dF9Q6Pnk8lM\nzCpudrcTRGO/e/v/rCRgCKKwaL5FyQjrha81eMdyV8Y61tR4A+StkSBOWFd3uWZ5\n+zm6zXjskb0R/tR/aszBPfQyRsOAXJsGzMry8+8i+3b1xGWxQKW6v9j2Ou9oAGkw\nhmyCY+bgPZQRqKO+zhsCTtg59p2Rj63u17w33NAokOU6riUM4fY/lmp6MPuSlElj\n2PdCP9aczzXKmZoRym1cL0YTfvse+v9K3Jt2XMFW7ddSZAoqjWMx\n-----END CERTIFICATE-----"

	intermediatePrivateKey  = "-----BEGIN PRIVATE KEY-----\n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\n4ktbfSBxycIVmpMMzJhNsRCrdt3lGcU=\n-----END PRIVATE KEY-----"
	intermediateCertificate = "-----BEGIN CERTIFICATE-----\nMIIFazCCA1OgAwIBAgIJAML9xGazwOBhMA0GCSqGSIb3DQEBCwUAMFkxCzAJBgNV\nBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMRwwGgYDVQQKDBN5b3VuaXF4IElk\nZW50aXR5IEFHMRcwFQYDVQQDDA5Vbml0IFRlc3QgUm9vdDAeFw0yMTA4MTgxMjAx\nMDBaFw0yNjA4MTcxMjAxMDBaMD8xHDAaBgNVBAoME3lvdW5pcXggSWRlbnRpdHkg\nQUcxHzAdBgNVBAMMFlVuaXQgVGVzdCBJbnRlcm1lZGlhdGUwggIiMA0GCSqGSIb3\nDQEBAQUAA4ICDwAwggIKAoICAQDQnTeJJgC+nkbJ1dLOaUcXfcVj0LLETioaEPls\nw9F4w4K3Na2zCnHBjUtWKuSkkZSYS3PR08BWL3qi0rNGFSun2IspFPpzrgql8lYv\ndiSJ+VItt2zPl4IlJWCD9T5ssvN6mwgMHuJR/4FRHMpAlrSXkIctvOyMUp2+svDn\neCLJEcqWIzTOiUA9avQQq7lRB/xvJxmW3HKKEuiE5Flx5Gdu13vvsIhHy5jG+C/6\nQ36zKnK8qGgkiNeeaxP1WjZxYhBWNMskqR6MO1LDlBkstMCGyfZltW3m8r+8Ttsm\ntv+a4JLQO1nIxzJcX9W3b7XdAbgdGiCk453biUYgrw6aNjNIB1t0/qySh/BsfXDS\nLqD2mWtvWqzOxwr0d55xtlG+I4e19krVdexWQFGJzx5LzIXXf9no1Ar1a5+822x8\n/TExpi+w/h/yihLusRDERnuSrwe6brtoc6Hf4d1b55XFBPQqVTxCaW+IgMqxJA6y\nuNZlAwLC/MwTDTHpL+yMy5ZcsznLi2A+h7XQy/HNP8M6aCDWimLPQBbypodrSt7n\nupXO+PXV46Vp8Ir2SwSVoUCU6lNg6+l06dmELguWjg7faw8ESdeIKt0vN+Cozb23\nj/JWW7ZS6xL5CU0VC4M++Kt0aU69fdJ0vzwUehbFFZkt+sVFryGwr/EoEfEQtTEt\nlqyi/wIDAQABo1AwTjAdBgNVHQ4EFgQUYoYP9Dbb0rvCx32ROm5jQqY9VSowHwYD\nVR0jBBgwFoAUfJ9IjnAYeLDZ3Jq3caxbq/Sk4ucwDAYDVR0TBAUwAwEB/zANBgkq\nhkiG9w0BAQsFAAOCAgEAdm1Fk0p4qVC2m5jW8qYLZZJ2EFSRnsP8KgHq4hopx5zr\neicBBxo+2tdl3SUvtJClywNlpsnmUSLnajCpKBeg/84EYWuAsXcnqtTnftFz9H+Q\nftY1XAsTUEK4Ps4KbM1XXAzkYQdtozpdOkvfEL1JKeynygvlYfQTnILUhGXb+JjU\nGus+ajcIoOEcHybs5CLwX+in77VioyV6K2haKZ04olT9JFPRsZV0tnjdSl6PU2Nz\ndrt+QeukW+KEoTE8SiMoU+cBaHrGbZEIU5ZzLnwlVxXitutPCDxYPZD+K/CKM+O5\nA3g5BSIVcpwAzS1gF72adUDWZqmdAQce974Nma/3Xs/LfQIHEV+GtRnMlo6SMNJM\nZ4cwRL4kr+xmDBgGdK0HsYdNemK+HqgItQwjJVSgr+6jL12FBZlXvoDHZ35oOlUw\nOkqXT8Pe7ihFm3ksUnbAebXaCGCw9bb0oG2gX2M857MkrA3gf/5hjtf9mcob3Axn\nS58arvUps4pGOg18FsuKX31ZhEZnmEammN98oXctcWBCWIOWpf0bxlRDmE71c1ah\n/jVziXFPY20DvCKMi1B4iSGtbOtE4S+dffg35FRVReki2OcwSSm+kpBQeg4VR0Ys\nswhK4VX9/1pCp1QTQe4mF0rHFWLvA4L3dVrkXzR2ZSdyG8/tFqrjedMm4yX+1X4=\n-----END CERTIFICATE-----"
)

var _ = Describe("PKI API", func() {
	When("creating a new PKI Engine", func() {
		engine := &pki.Engine{
			Path: "managed/pki/some-engine",
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(10 * core.Year),
			},
		}

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(engine)).Should(Succeed())
		})

		It("Should not exist before creating it", func() {
			vaultEnv.Mount(engine).Should(BeNil())
		})

		It("Should be able to create a pki engine", func() {
			Expect(vaultAPI.UpdatePKIEngine(engine)).To(Succeed())
			vaultEnv.Mount(engine).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-engine",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(10 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))
		})
	})

	When("creating a new root ca", func() {
		root := &pki.CA{
			Path: "managed/pki/some-root",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(10 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(10 * core.Year),
			},
		}

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(root)).Should(Succeed())
		})

		It("Should not exist before creating it", func() {
			vaultEnv.Mount(root).Should(BeNil())
		})

		It("Should be able to be created implicitly with an update", func() {
			_, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).To(HaveOccurred())

			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			vaultEnv.Mount(root).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-root",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(10 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).NotTo(BeEmpty())
		})

		It("Should be able to be created explicitly in internal mode", func() {
			info, err := vaultAPI.CreateRootCA(pki.ModeInternal, root)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Path).To(Equal("managed/pki/some-root"))
			Expect(info.SerialNumber).NotTo(BeEmpty())
			Expect(info.PrivateKey).To(BeEmpty())
			Expect(info.PrivateKeyType).To(BeEmpty())
			Expect(info.Certificate).NotTo(BeEmpty())
			Expect(info.CertificateChain).NotTo(BeEmpty())
			Expect(info.IssuingCertificateAuthority).To(Equal(info.Certificate))

			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			vaultEnv.Mount(root).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-root",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(10 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(info.Certificate))
		})

		It("Should be able to be created explicitly in exported mode", func() {
			info, err := vaultAPI.CreateRootCA(pki.ModeExported, root)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Path).To(Equal("managed/pki/some-root"))
			Expect(info.SerialNumber).NotTo(BeEmpty())
			Expect(info.PrivateKey).NotTo(BeEmpty())
			Expect(info.PrivateKeyType).NotTo(BeEmpty())
			Expect(info.Certificate).NotTo(BeEmpty())
			Expect(info.CertificateChain).NotTo(BeEmpty())
			Expect(info.IssuingCertificateAuthority).To(Equal(info.Certificate))

			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			vaultEnv.Mount(root).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-root",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(10 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(info.Certificate))
		})

		It("Should not modify an existing ca", func() {
			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			certA, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(certA).NotTo(BeEmpty())

			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			certB, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(certB).NotTo(BeEmpty())

			Expect(certA).To(Equal(certB))
		})
	})

	When("managing a root ca", func() {
		root := &pki.CA{
			Path: "managed/pki/some-root",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(10 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(10 * core.Year),
			},
		}

		BeforeEach(func() {
			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(root)).Should(Succeed())
		})

		It("Should exist", func() {
			vaultEnv.Mount(root).ShouldNot(BeNil())
		})

		It("Should throw an error if I tried to create it again", func() {
			info, err := vaultAPI.CreateRootCA(pki.ModeInternal, root)
			Expect(err).To(HaveOccurred())
			Expect(info).To(BeNil())
		})

		It("Should be readable", func() {
			ca, err := vaultAPI.ReadCA(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(ca).To(Equal(&pki.CA{
				Path:     "managed/pki/some-root",
				Settings: nil,
				Subject: &pki.Subject{
					CommonName:      "example.com",
					SubjectSettings: &pki.SubjectSettings{},
				},
				Config: &mount.TuneConfig{
					DefaultLeaseTTL: core.NewTTL(32 * core.Day),
					MaxLeaseTTL:     core.NewTTL(10 * core.Year),
				},
			}))
		})

		It("Should be able to run a tidy operation", func() {
			settings := &pki.TidySettings{
				TidyCertStore:    true,
				TidyRevokedCerts: true,
				SafetyBuffer:     core.NewTTL(3 * core.Day),
			}
			Expect(vaultAPI.Tidy(root, settings)).To(Succeed())
		})

		It("Should be possible to rotate CRLs", func() {
			Expect(vaultAPI.RotateCRLs(root)).To(Succeed())
		})
	})

	When("importing a root ca during create", func() {
		root := &pki.CA{
			Path: "managed/pki/some-root",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(10 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(10 * core.Year),
			},
			ImportedCert: &pki.ImportedCert{
				PrivateKey:  rootPrivateKey,
				Certificate: rootCertificate,
			},
		}

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(root)).Should(Succeed())
		})

		It("Should not exist before creating it", func() {
			vaultEnv.Mount(root).Should(BeNil())
		})

		It("Should be able to be created implicitly with an update", func() {
			_, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).To(HaveOccurred())

			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			vaultEnv.Mount(root).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-root",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(10 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(rootCertificate))
		})

		It("Should be able to be created explicitly in internal mode", func() {
			info, err := vaultAPI.CreateRootCA(pki.ModeInternal, root)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Path).To(Equal("managed/pki/some-root"))
			Expect(info.SerialNumber).To(Equal("39:af:8c:ff:af:94:27:5f:49:7f:91:99:cc:ad:2e:cc:a3:bf:15:d7"))
			Expect(info.PrivateKey).To(BeEmpty())
			Expect(info.PrivateKeyType).To(BeEmpty())
			Expect(info.Certificate).To(Equal(rootCertificate))
			Expect(info.CertificateChain).To(Equal(rootCertificate))
			Expect(info.IssuingCertificateAuthority).To(Equal(rootCertificate))

			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			vaultEnv.Mount(root).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-root",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(10 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(rootCertificate))
		})

		It("Should be able to be created explicitly in exported mode", func() {
			info, err := vaultAPI.CreateRootCA(pki.ModeExported, root)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Path).To(Equal("managed/pki/some-root"))
			Expect(info.SerialNumber).To(Equal("39:af:8c:ff:af:94:27:5f:49:7f:91:99:cc:ad:2e:cc:a3:bf:15:d7"))
			Expect(info.PrivateKey).To(Equal(rootPrivateKey))
			Expect(info.PrivateKeyType).To(Equal(pki.KeyTypeRSA))
			Expect(info.Certificate).To(Equal(rootCertificate))
			Expect(info.CertificateChain).To(Equal(rootCertificate))
			Expect(info.IssuingCertificateAuthority).To(Equal(rootCertificate))

			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			vaultEnv.Mount(root).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-root",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(10 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(info.Certificate))
		})

		It("Should not modify an existing ca", func() {
			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			certA, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(certA).To(Equal(rootCertificate))

			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			certB, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(certB).To(Equal(rootCertificate))
		})
	})

	When("importing a root ca during update", func() {
		root := &pki.CA{
			Path: "managed/pki/some-root",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(10 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA4096,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(10 * core.Year),
			},
			ImportedCert: &pki.ImportedCert{
				PrivateKey:  rootPrivateKey,
				Certificate: rootCertificate,
			},
		}

		BeforeEach(func() {
			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(root)).Should(Succeed())
		})

		It("Should exist", func() {
			vaultEnv.Mount(root).ShouldNot(BeNil())
		})

		It("Should throw an error if I tried to create it again", func() {
			info, err := vaultAPI.CreateRootCA(pki.ModeInternal, root)
			Expect(err).To(HaveOccurred())
			Expect(info).To(BeNil())
		})

		It("Should have the imported cert", func() {
			pem, err := vaultAPI.ReadCACertificatePEM(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(pem).To(Equal(rootCertificate))
		})

		It("Should be readable", func() {
			ca, err := vaultAPI.ReadCA(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(ca).To(Equal(&pki.CA{
				Path:     "managed/pki/some-root",
				Settings: nil,
				Subject: &pki.Subject{
					CommonName: "Unit Test Root",
					SubjectSettings: &pki.SubjectSettings{
						Organization: pki.StringArray{"youniqx Identity AG"},
						Country:      pki.StringArray{"AU"},
						Province:     pki.StringArray{"Some-State"},
					},
				},
				Config: &mount.TuneConfig{
					DefaultLeaseTTL: core.NewTTL(32 * core.Day),
					MaxLeaseTTL:     core.NewTTL(10 * core.Year),
				},
			}))
		})

		It("Should be able to run a tidy operation", func() {
			settings := &pki.TidySettings{
				TidyCertStore:    true,
				TidyRevokedCerts: true,
				SafetyBuffer:     core.NewTTL(3 * core.Day),
			}
			Expect(vaultAPI.Tidy(root, settings)).To(Succeed())
		})

		It("Should be possible to rotate CRLs", func() {
			Expect(vaultAPI.RotateCRLs(root)).To(Succeed())
		})
	})

	When("creating a new intermediate ca", func() {
		root := &pki.CA{
			Path: "managed/pki/some-root",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(10 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(10 * core.Year),
			},
		}

		intermediate := &pki.CA{
			Path: "managed/pki/some-intermediate",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(5 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(5 * core.Year),
			},
		}

		intermediate2 := &pki.CA{
			Path: "managed/pki/some-intermediate-2",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(5 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(5 * core.Year),
			},
		}

		var rootCACert string

		BeforeEach(func() {
			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			var err error
			rootCACert, err = vaultAPI.ReadCACertificatePEM(root)
			Expect(err).NotTo(HaveOccurred())
			Expect(rootCACert).NotTo(BeEmpty())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(intermediate)).Should(Succeed())
			Expect(vaultAPI.DeleteEngine(intermediate2)).Should(Succeed())
			Expect(vaultAPI.DeleteEngine(root)).Should(Succeed())
		})

		It("Should not exist before creating it", func() {
			vaultEnv.CA(intermediate).Should(BeNil())
		})

		It("Should be able to be created implicitly with an update", func() {
			Expect(vaultAPI.UpdateIntermediateCA(root, intermediate)).To(Succeed())
			vaultEnv.Mount(intermediate).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-intermediate",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(5 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).NotTo(BeEmpty())
		})

		It("Should be able to be created explicitly in internal mode", func() {
			info, err := vaultAPI.CreateIntermediateCA(pki.ModeInternal, root, intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Path).To(Equal("managed/pki/some-intermediate"))
			Expect(info.SerialNumber).NotTo(BeEmpty())
			Expect(info.PrivateKey).To(BeEmpty())
			Expect(info.PrivateKeyType).To(BeEmpty())
			Expect(info.Certificate).NotTo(BeEmpty())
			expectedChain := fmt.Sprintf("%s\n%s", info.Certificate, rootCACert)
			Expect(info.CertificateChain).To(Equal(expectedChain))
			Expect(info.IssuingCertificateAuthority).NotTo(BeEmpty())
			Expect(info.IssuingCertificateAuthority).NotTo(Equal(info.Certificate))
			Expect(info.IssuingCertificateAuthority).To(Equal(rootCACert))

			vaultEnv.Mount(intermediate).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-intermediate",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(5 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(info.Certificate))
		})

		It("Should be able to be a longer chain in internal mode", func() {
			info, err := vaultAPI.CreateIntermediateCA(pki.ModeInternal, root, intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Path).To(Equal("managed/pki/some-intermediate"))
			Expect(info.SerialNumber).NotTo(BeEmpty())
			Expect(info.PrivateKey).To(BeEmpty())
			Expect(info.PrivateKeyType).To(BeEmpty())
			Expect(info.Certificate).NotTo(BeEmpty())
			expectedChain := fmt.Sprintf("%s\n%s", info.Certificate, rootCACert)
			Expect(info.CertificateChain).To(Equal(expectedChain))
			Expect(info.IssuingCertificateAuthority).NotTo(BeEmpty())
			Expect(info.IssuingCertificateAuthority).NotTo(Equal(info.Certificate))
			Expect(info.IssuingCertificateAuthority).To(Equal(rootCACert))

			vaultEnv.Mount(intermediate).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-intermediate",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(5 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(info.Certificate))

			info2, err := vaultAPI.CreateIntermediateCA(pki.ModeInternal, intermediate, intermediate2)
			Expect(err).NotTo(HaveOccurred())
			Expect(info2).NotTo(BeNil())
			Expect(info2.Path).To(Equal("managed/pki/some-intermediate-2"))
			Expect(info2.SerialNumber).NotTo(BeEmpty())
			Expect(info2.PrivateKey).To(BeEmpty())
			Expect(info2.PrivateKeyType).To(BeEmpty())
			Expect(info2.Certificate).NotTo(BeEmpty())
			Expect(info2.CertificateChain).To(Equal(fmt.Sprintf("%s\n%s\n%s", info2.Certificate, info.Certificate, rootCACert)))
			Expect(info2.IssuingCertificateAuthority).NotTo(Equal(info2.Certificate))
			Expect(info2.IssuingCertificateAuthority).To(Equal(info.Certificate))

			vaultEnv.Mount(intermediate2).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-intermediate-2",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(5 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err = vaultAPI.ReadCACertificatePEM(intermediate2)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(info2.Certificate))
		})

		It("Should be able to be created explicitly in exported mode", func() {
			info, err := vaultAPI.CreateIntermediateCA(pki.ModeExported, root, intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Path).To(Equal("managed/pki/some-intermediate"))
			Expect(info.SerialNumber).NotTo(BeEmpty())
			Expect(info.PrivateKey).NotTo(BeEmpty())
			Expect(info.PrivateKeyType).NotTo(BeEmpty())
			Expect(info.Certificate).NotTo(BeEmpty())
			expectedChain := fmt.Sprintf("%s\n%s", info.Certificate, rootCACert)
			Expect(info.CertificateChain).To(Equal(expectedChain))
			Expect(info.IssuingCertificateAuthority).NotTo(BeEmpty())
			Expect(info.IssuingCertificateAuthority).NotTo(Equal(info.Certificate))
			Expect(info.IssuingCertificateAuthority).To(Equal(rootCACert))

			vaultEnv.Mount(intermediate).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-intermediate",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(5 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(info.Certificate))
		})

		It("Should be able to be a longer chain in exported mode", func() {
			info, err := vaultAPI.CreateIntermediateCA(pki.ModeExported, root, intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Path).To(Equal("managed/pki/some-intermediate"))
			Expect(info.SerialNumber).NotTo(BeEmpty())
			Expect(info.PrivateKey).NotTo(BeEmpty())
			Expect(info.PrivateKeyType).NotTo(BeEmpty())
			Expect(info.Certificate).NotTo(BeEmpty())
			expectedChain := fmt.Sprintf("%s\n%s", info.Certificate, rootCACert)
			Expect(info.CertificateChain).To(Equal(expectedChain))
			Expect(info.IssuingCertificateAuthority).NotTo(BeEmpty())
			Expect(info.IssuingCertificateAuthority).NotTo(Equal(info.Certificate))
			Expect(info.IssuingCertificateAuthority).To(Equal(rootCACert))

			vaultEnv.Mount(intermediate).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-intermediate",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(5 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(info.Certificate))

			info2, err := vaultAPI.CreateIntermediateCA(pki.ModeExported, intermediate, intermediate2)
			Expect(err).NotTo(HaveOccurred())
			Expect(info2).NotTo(BeNil())
			Expect(info2.Path).To(Equal("managed/pki/some-intermediate-2"))
			Expect(info2.SerialNumber).NotTo(BeEmpty())
			Expect(info2.PrivateKey).NotTo(BeEmpty())
			Expect(info2.PrivateKeyType).NotTo(BeEmpty())
			Expect(info2.Certificate).NotTo(BeEmpty())
			Expect(info2.CertificateChain).To(Equal(fmt.Sprintf("%s\n%s\n%s", info2.Certificate, info.Certificate, rootCACert)))
			Expect(info2.IssuingCertificateAuthority).NotTo(Equal(info2.Certificate))
			Expect(info2.IssuingCertificateAuthority).To(Equal(info.Certificate))

			vaultEnv.Mount(intermediate2).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-intermediate-2",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(5 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err = vaultAPI.ReadCACertificatePEM(intermediate2)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(info2.Certificate))
		})

		It("Should not modify an already existing CA", func() {
			Expect(vaultAPI.UpdateIntermediateCA(root, intermediate)).To(Succeed())
			certA, err := vaultAPI.ReadCACertificatePEM(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(certA).NotTo(BeEmpty())

			Expect(vaultAPI.UpdateIntermediateCA(root, intermediate)).To(Succeed())
			certB, err := vaultAPI.ReadCACertificatePEM(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(certB).NotTo(BeEmpty())

			Expect(certA).To(Equal(certB))
		})
	})

	When("managing an intermediate ca", func() {
		root := &pki.CA{
			Path: "managed/pki/some-root",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(10 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(10 * core.Year),
			},
		}

		intermediate := &pki.CA{
			Path: "managed/pki/some-intermediate",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(5 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(5 * core.Year),
			},
		}

		BeforeEach(func() {
			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			Expect(vaultAPI.UpdateIntermediateCA(root, intermediate)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(intermediate)).Should(Succeed())
			Expect(vaultAPI.DeleteEngine(root)).Should(Succeed())
		})

		It("Should exist", func() {
			vaultEnv.CA(intermediate).ShouldNot(BeNil())
		})

		It("Should throw an error if I tried to create it again", func() {
			info, err := vaultAPI.CreateIntermediateCA(pki.ModeInternal, root, intermediate)
			Expect(err).To(HaveOccurred())
			Expect(info).To(BeNil())
		})

		It("Should be readable", func() {
			ca, err := vaultAPI.ReadCA(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(ca).To(Equal(&pki.CA{
				Path:     "managed/pki/some-intermediate",
				Settings: nil,
				Subject: &pki.Subject{
					CommonName:      "example.com",
					SubjectSettings: &pki.SubjectSettings{},
				},
				Config: &mount.TuneConfig{
					DefaultLeaseTTL: core.NewTTL(32 * core.Day),
					MaxLeaseTTL:     core.NewTTL(5 * core.Year),
				},
			}))
		})

		It("Should be able to run a tidy operation", func() {
			settings := &pki.TidySettings{
				TidyCertStore:    true,
				TidyRevokedCerts: true,
				SafetyBuffer:     core.NewTTL(3 * core.Day),
			}
			Expect(vaultAPI.Tidy(intermediate, settings)).To(Succeed())
		})

		It("Should be possible to rotate CRLs", func() {
			Expect(vaultAPI.RotateCRLs(intermediate)).To(Succeed())
		})
	})

	When("importing an intermediate ca during create", func() {
		root := &pki.CA{
			Path: "managed/pki/some-root",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(10 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(10 * core.Year),
			},
			ImportedCert: &pki.ImportedCert{
				PrivateKey:  rootPrivateKey,
				Certificate: rootCertificate,
			},
		}

		intermediate := &pki.CA{
			Path: "managed/pki/some-intermediate",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(5 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(5 * core.Year),
			},
			ImportedCert: &pki.ImportedCert{
				PrivateKey:  intermediatePrivateKey,
				Certificate: intermediateCertificate,
			},
		}

		BeforeEach(func() {
			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(intermediate)).Should(Succeed())
			Expect(vaultAPI.DeleteEngine(root)).Should(Succeed())
		})

		It("Should not exist before creating it", func() {
			vaultEnv.CA(intermediate).Should(BeNil())
		})

		It("Should be able to be created explicitly in internal mode", func() {
			info, err := vaultAPI.CreateIntermediateCA(pki.ModeInternal, root, intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Path).To(Equal("managed/pki/some-intermediate"))
			Expect(info.SerialNumber).To(Equal("c2:fd:c4:66:b3:c0:e0:61"))
			Expect(info.PrivateKey).To(BeEmpty())
			Expect(info.PrivateKeyType).To(BeEmpty())
			Expect(info.Certificate).To(Equal(intermediateCertificate))
			Expect(info.CertificateChain).To(Equal(intermediateCertificate))
			Expect(info.IssuingCertificateAuthority).To(Equal(rootCertificate))

			vaultEnv.Mount(intermediate).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-intermediate",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(5 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(intermediateCertificate))
		})

		It("Should be able to be created explicitly in exported mode", func() {
			info, err := vaultAPI.CreateIntermediateCA(pki.ModeExported, root, intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Path).To(Equal("managed/pki/some-intermediate"))
			Expect(info.SerialNumber).To(Equal("c2:fd:c4:66:b3:c0:e0:61"))
			Expect(info.PrivateKey).To(Equal(intermediatePrivateKey))
			Expect(info.PrivateKeyType).To(Equal(pki.KeyTypeRSA))
			Expect(info.Certificate).To(Equal(intermediateCertificate))
			Expect(info.CertificateChain).To(Equal(intermediateCertificate))
			Expect(info.IssuingCertificateAuthority).To(Equal(rootCertificate))

			vaultEnv.Mount(intermediate).Should(Equal(&mount.Mount{
				Path: "managed/pki/some-intermediate",
				Type: "pki",
				Config: &mount.TuneConfig{
					MaxLeaseTTL:     core.NewTTL(5 * core.Year),
					DefaultLeaseTTL: core.NewTTL(0),
				},
			}))

			cert, err := vaultAPI.ReadCACertificatePEM(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal(intermediateCertificate))
		})

		It("Should not modify an already existing CA", func() {
			Expect(vaultAPI.UpdateIntermediateCA(root, intermediate)).To(Succeed())
			certA, err := vaultAPI.ReadCACertificatePEM(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(certA).To(Equal(intermediateCertificate))

			Expect(vaultAPI.UpdateIntermediateCA(root, intermediate)).To(Succeed())
			certB, err := vaultAPI.ReadCACertificatePEM(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(certB).To(Equal(intermediateCertificate))
		})
	})

	When("importing an intermediate ca during update", func() {
		root := &pki.CA{
			Path: "managed/pki/some-root",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(10 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA4096,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(10 * core.Year),
			},
			ImportedCert: &pki.ImportedCert{
				PrivateKey:  rootPrivateKey,
				Certificate: rootCertificate,
			},
		}

		intermediate := &pki.CA{
			Path: "managed/pki/some-intermediate",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(5 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA4096,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(5 * core.Year),
			},
			ImportedCert: &pki.ImportedCert{
				PrivateKey:  intermediatePrivateKey,
				Certificate: intermediateCertificate,
			},
		}

		BeforeEach(func() {
			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			Expect(vaultAPI.UpdateIntermediateCA(root, intermediate)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(intermediate)).Should(Succeed())
			Expect(vaultAPI.DeleteEngine(root)).Should(Succeed())
		})

		It("Should exist", func() {
			vaultEnv.CA(intermediate).ShouldNot(BeNil())
		})

		It("Should throw an error if I tried to create it again", func() {
			info, err := vaultAPI.CreateIntermediateCA(pki.ModeInternal, root, intermediate)
			Expect(err).To(HaveOccurred())
			Expect(info).To(BeNil())
		})

		It("should have its imported cert", func() {
			pem, err := vaultAPI.ReadCACertificatePEM(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(pem).To(Equal(intermediateCertificate))
		})

		It("Should be readable", func() {
			ca, err := vaultAPI.ReadCA(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(ca).To(Equal(&pki.CA{
				Path:     "managed/pki/some-intermediate",
				Settings: nil,
				Subject: &pki.Subject{
					CommonName: "Unit Test Intermediate",
					SubjectSettings: &pki.SubjectSettings{
						Organization: pki.StringArray{"youniqx Identity AG"},
					},
				},
				Config: &mount.TuneConfig{
					DefaultLeaseTTL: core.NewTTL(32 * core.Day),
					MaxLeaseTTL:     core.NewTTL(5 * core.Year),
				},
			}))
		})

		It("Should be able to run a tidy operation", func() {
			settings := &pki.TidySettings{
				TidyCertStore:    true,
				TidyRevokedCerts: true,
				SafetyBuffer:     core.NewTTL(3 * core.Day),
			}
			Expect(vaultAPI.Tidy(intermediate, settings)).To(Succeed())
		})

		It("Should be possible to rotate CRLs", func() {
			Expect(vaultAPI.RotateCRLs(intermediate)).To(Succeed())
		})
	})

	When("creating a certificate role", func() {
		root := &pki.CA{
			Path: "managed/pki/some-root",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(10 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(10 * core.Year),
			},
		}

		intermediate := &pki.CA{
			Path: "managed/pki/some-intermediate",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(5 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(5 * core.Year),
			},
		}

		role := &pki.CertificateRole{
			Name: "some-role",
			Settings: &pki.RoleSettings{
				TTL:     core.NewTTL(core.Week),
				MaxTTL:  core.NewTTL(core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.SubjectSettings{
				Organization: []string{"youniqx Identity AG"},
			},
		}

		BeforeEach(func() {
			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			Expect(vaultAPI.UpdateIntermediateCA(root, intermediate)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(intermediate)).Should(Succeed())
			Expect(vaultAPI.DeleteEngine(root)).Should(Succeed())
		})

		It("Should be able to create a certificate role", func() {
			Expect(vaultAPI.UpdateCertificateRole(intermediate, role)).To(Succeed())
		})
	})

	When("managing a certificate role", func() {
		root := &pki.CA{
			Path: "managed/pki/some-root",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(10 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(10 * core.Year),
			},
		}

		intermediate := &pki.CA{
			Path: "managed/pki/some-intermediate",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(5 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "example.com",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(5 * core.Year),
			},
		}

		role := &pki.CertificateRole{
			Name: "some-role",
			Settings: &pki.RoleSettings{
				TTL:              core.NewTTL(core.Week),
				MaxTTL:           core.NewTTL(core.Year),
				KeyType:          pki.KeyTypeRSA,
				KeyBits:          pki.KeyBitsRSA2048,
				AllowedDomains:   []string{"example.com"},
				AllowBareDomains: true,
				AllowSubdomains:  true,
			},
			Subject: &pki.SubjectSettings{
				Organization: []string{"youniqx Identity AG"},
			},
		}

		BeforeEach(func() {
			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			Expect(vaultAPI.UpdateIntermediateCA(root, intermediate)).To(Succeed())
			Expect(vaultAPI.UpdateCertificateRole(intermediate, role)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(intermediate)).Should(Succeed())
			Expect(vaultAPI.DeleteEngine(root)).Should(Succeed())
		})

		It("Should be able to delete non-existing certificate roles", func() {
			Expect(vaultAPI.DeleteCertificateRole(intermediate, core.RoleName("some-nonexisting-role"))).To(Succeed())
		})

		It("Should exist", func() {
			vaultEnv.CertificateRole(intermediate, role).ShouldNot(BeNil())
		})

		It("Should be able to issue a certificate", func() {
			cert, err := vaultAPI.IssueCertificate(intermediate, role, &pki.IssueCertOptions{
				CommonName:        "example.com",
				DNSSans:           nil,
				ExcludeCNFromSans: false,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).NotTo(BeNil())
		})

		It("Should be able to sign a csr", func() {
			keyBytes, _ := rsa.GenerateKey(rand.Reader, 2048)

			emailAddress := "test@example.com"
			subj := pkix.Name{
				CommonName:         "example.com",
				Country:            []string{"AU"},
				Province:           []string{"Some-State"},
				Locality:           []string{"MyCity"},
				Organization:       []string{"Company Ltd"},
				OrganizationalUnit: []string{"IT"},
			}
			rawSubj := subj.ToRDNSequence()
			rawSubj = append(rawSubj, []pkix.AttributeTypeAndValue{
				{Type: asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}, Value: emailAddress},
			})

			asn1Subj, err := asn1.Marshal(rawSubj)
			Expect(err).NotTo(HaveOccurred())
			template := x509.CertificateRequest{
				RawSubject:         asn1Subj,
				EmailAddresses:     []string{emailAddress},
				SignatureAlgorithm: x509.SHA256WithRSA,
			}

			csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, keyBytes)
			Expect(err).NotTo(HaveOccurred())

			csr := pem.EncodeToMemory(&pem.Block{
				Type: "CERTIFICATE REQUEST", Bytes: csrBytes,
			})

			cert, err := vaultAPI.SignCertificateSigningRequest(intermediate, role, &pki.SignCsr{
				CSR:        string(csr),
				CommonName: "test.example.com",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).NotTo(BeNil())
		})

		It("Should be able to be deleted", func() {
			vaultEnv.CertificateRole(intermediate, role).ShouldNot(BeNil())
			Expect(vaultAPI.DeleteCertificateRole(intermediate, role)).To(Succeed())
			vaultEnv.CertificateRole(intermediate, role).Should(BeNil())
		})

		It("Should be readable", func() {
			certRole, err := vaultAPI.ReadCertificateRole(intermediate, role)
			Expect(err).NotTo(HaveOccurred())
			Expect(certRole).To(Equal(&pki.CertificateRole{
				Name: "some-role",
				Settings: &pki.RoleSettings{
					TTL:                           core.NewTTL(core.Week),
					MaxTTL:                        core.NewTTL(core.Year),
					AllowLocalhost:                false,
					AllowedDomains:                []string{"example.com"},
					AllowedDomainsTemplate:        false,
					AllowBareDomains:              true,
					AllowSubdomains:               true,
					AllowGlobDomains:              false,
					AllowAnyName:                  false,
					EnforceHostNames:              false,
					AllowIPSans:                   false,
					AllowedURISans:                []string{},
					AllowedOtherSans:              []string{},
					ServerFlag:                    false,
					ClientFlag:                    false,
					CodeSigningFlag:               false,
					EmailProtectionFlag:           false,
					KeyType:                       pki.KeyTypeRSA,
					KeyBits:                       pki.KeyBitsRSA2048,
					KeyUsage:                      []pki.KeyUsage{},
					ExtendedKeyUsage:              []pki.ExtendedKeyUsage{},
					ExtendedKeyUsageOids:          []string{},
					UseCSRCommonName:              false,
					UseCSRSans:                    false,
					GenerateLease:                 false,
					NoStore:                       false,
					RequireCommonName:             false,
					PolicyIdentifiers:             []string{},
					BasicConstraintsValidForNonCA: false,
					NotBeforeDuration:             core.NewTTL(30 * time.Second),
				},
				Subject: &pki.SubjectSettings{
					Organization:       []string{"youniqx Identity AG"},
					OrganizationalUnit: []string{},
					Country:            []string{},
					Locality:           []string{},
					Province:           []string{},
					StreetAddress:      []string{},
					PostalCode:         []string{},
				},
			}))
		})
	})

	When("manging a certificate from a role", func() {
		root := &pki.CA{
			Path: "managed/pki/some-root",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(10 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "youniqx Root CA",
				SubjectSettings: &pki.SubjectSettings{
					Organization: []string{"youniqx Identity AG"},
				},
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(10 * core.Year),
			},
		}

		intermediate := &pki.CA{
			Path: "managed/pki/some-intermediate",
			Settings: &pki.CASettings{
				TTL:     core.NewTTL(5 * core.Year),
				KeyType: pki.KeyTypeRSA,
				KeyBits: pki.KeyBitsRSA2048,
			},
			Subject: &pki.Subject{
				CommonName: "youniqx Intermediate CA",
				SubjectSettings: &pki.SubjectSettings{
					Organization: []string{"youniqx Identity AG"},
				},
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL: core.NewTTL(5 * core.Year),
			},
		}

		role := &pki.CertificateRole{
			Name: "some-role",
			Settings: &pki.RoleSettings{
				TTL:              core.NewTTL(core.Week),
				MaxTTL:           core.NewTTL(core.Year),
				KeyType:          pki.KeyTypeRSA,
				KeyBits:          pki.KeyBitsRSA2048,
				AllowedDomains:   []string{"example.com"},
				AllowBareDomains: true,
			},
			Subject: &pki.SubjectSettings{
				Organization: []string{"youniqx Identity AG"},
			},
		}

		var certificate *pki.Certificate

		BeforeEach(func() {
			Expect(vaultAPI.UpdateRootCA(root)).To(Succeed())
			Expect(vaultAPI.UpdateIntermediateCA(root, intermediate)).To(Succeed())
			Expect(vaultAPI.UpdateCertificateRole(intermediate, role)).To(Succeed())
			var err error
			certificate, err = vaultAPI.IssueCertificate(intermediate, role, &pki.IssueCertOptions{
				CommonName:        "example.com",
				DNSSans:           nil,
				ExcludeCNFromSans: false,
			})
			Expect(certificate).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(intermediate)).Should(Succeed())
			Expect(vaultAPI.DeleteEngine(root)).Should(Succeed())
		})

		It("Should be properly configured", func() {
			block, extra := pem.Decode([]byte(certificate.Certificate))
			Expect(block).NotTo(BeNil())
			Expect(extra).To(BeEmpty())

			x509Cert, err := x509.ParseCertificate(block.Bytes)
			Expect(err).NotTo(HaveOccurred())

			Expect(x509Cert.Issuer.Organization).To(ContainElement("youniqx Identity AG"))
			Expect(x509Cert.Issuer.Organization).To(HaveLen(1))
			Expect(x509Cert.Issuer.CommonName).To(Equal("youniqx Intermediate CA"))

			Expect(x509Cert.Subject.Organization).To(ContainElement("youniqx Identity AG"))
			Expect(x509Cert.Subject.Organization).To(HaveLen(1))

			Expect(x509Cert.PublicKeyAlgorithm).To(Equal(x509.RSA))

			Expect(x509Cert.CRLDistributionPoints).To(ContainElement("http://127.0.0.1:8100/v1/managed/pki/some-intermediate/crl"))
			Expect(x509Cert.CRLDistributionPoints).To(HaveLen(1))

			Expect(x509Cert.IssuingCertificateURL).To(ContainElement("http://127.0.0.1:8100/v1/managed/pki/some-intermediate/ca"))
			Expect(x509Cert.IssuingCertificateURL).To(HaveLen(1))
		})

		It("Should be able to be revoked", func() {
			Expect(vaultAPI.RevokeCertificate(intermediate, certificate)).To(Succeed())
		})

		It("Should appear in the the list of certificates of the CA", func() {
			certs, err := vaultAPI.ListCerts(intermediate)
			Expect(err).NotTo(HaveOccurred())
			Expect(certs).To(ContainElement(certificate.SerialNumber))
		})
	})
})
