package testenv

import (
	"time"

	"github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/vault"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/kvsecret"
	"github.com/youniqx/heist/pkg/vault/transit"
)

const (
	timeout      = 30 * time.Second
	pollInterval = 200 * time.Millisecond
)

func (e *testEnv) fetch(fetcher func(api vault.API) interface{}) gomega.AsyncAssertion {
	return gomega.Eventually(func() interface{} {
		return fetcher(e.RootAPI)
	}, timeout, pollInterval)
}

type TestSecret interface {
	kvsecret.Entity
	Engine() core.MountPathEntity
	API() vault.API
}

type testSecret struct {
	kvsecret.Entity
	EngineRef core.MountPathEntity
	VaultAPI  vault.API
}

func (t *testSecret) Engine() core.MountPathEntity {
	return t.EngineRef
}

func (t *testSecret) API() vault.API {
	return t.VaultAPI
}

func (e *testEnv) KvSecret(engine core.MountPathEntity, secret core.SecretPathEntity) gomega.AsyncAssertion {
	return e.fetch(func(api vault.API) interface{} {
		secret, _ := api.ReadKvSecret(engine, secret)
		if secret == nil {
			return nil
		}
		return &testSecret{
			Entity:    secret,
			EngineRef: engine,
			VaultAPI:  api,
		}
	})
}

func (e *testEnv) KvEngine(engine core.MountPathEntity) gomega.AsyncAssertion {
	return e.fetch(func(api vault.API) interface{} {
		engine, _ := api.ReadKvEngine(engine)
		return engine
	})
}

func (e *testEnv) TransitEngine(engine core.MountPathEntity) gomega.AsyncAssertion {
	return e.fetch(func(api vault.API) interface{} {
		engine, _ := api.ReadTransitEngine(engine)
		return engine
	})
}

func (e *testEnv) TransitKey(engine core.MountPathEntity, key transit.KeyNameEntity) gomega.AsyncAssertion {
	return e.fetch(func(api vault.API) interface{} {
		engine, _ := api.ReadTransitKey(engine, key)
		return engine
	})
}

func (e *testEnv) Policy(policy core.PolicyNameEntity) gomega.AsyncAssertion {
	return e.fetch(func(api vault.API) interface{} {
		engine, _ := api.ReadPolicy(policy)
		return engine
	})
}

func (e *testEnv) KubernetesAuthRole(auth core.MountPathEntity, role core.RoleNameEntity) gomega.AsyncAssertion {
	return e.fetch(func(api vault.API) interface{} {
		role, _ := api.ReadKubernetesAuthRole(auth, role)
		return role
	})
}

func (e *testEnv) AuthMethod(auth core.MountPathEntity) gomega.AsyncAssertion {
	return e.fetch(func(api vault.API) interface{} {
		auth, _ := api.ReadAuthMethod(auth)
		return auth
	})
}

func (e *testEnv) Mount(mount core.MountPathEntity) gomega.AsyncAssertion {
	return e.fetch(func(api vault.API) interface{} {
		mount, _ := api.ReadMount(mount)
		return mount
	})
}

func (e *testEnv) CA(mount core.MountPathEntity) gomega.AsyncAssertion {
	return e.fetch(func(api vault.API) interface{} {
		ca, _ := api.ReadCA(mount)
		return ca
	})
}

func (e *testEnv) TuneConfig(engine core.MountPathEntity) gomega.AsyncAssertion {
	return e.fetch(func(api vault.API) interface{} {
		config, _ := api.ReadTuneConfig(engine)
		return config
	})
}

func (e *testEnv) CertificateRole(ca core.MountPathEntity, role core.RoleNameEntity) gomega.AsyncAssertion {
	return e.fetch(func(api vault.API) interface{} {
		certRole, _ := api.ReadCertificateRole(ca, role)
		return certRole
	})
}
