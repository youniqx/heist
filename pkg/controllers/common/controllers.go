package common

import (
	"context"
	"time"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/testhelper"
	"github.com/youniqx/heist/pkg/vault"
	"github.com/youniqx/heist/pkg/vault/testenv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

type TestDataWrapper struct {
	K8sClient         client.Client
	TestEnv           *envtest.Environment
	VaultEnv          testenv.TestEnv
	RootAPI           vault.API
	K8sEnv            testhelper.KubernetesTestHelper
	DefaultCipherText string
}

const (
	requeueAfter = 60 * time.Second
	// YouniqxFinalizer defines the finalizer value used by all CRDs.
	YouniqxFinalizer = "youniqx.com/finalizer"
)

// Requeue unifies the returned controller result when an error occurs.
// By default it the handling of the CRD is queued to run in 5 seconds.
var Requeue = ctrl.Result{
	Requeue:      true,
	RequeueAfter: requeueAfter,
}

func FindRootCA(ctx context.Context, k8s client.Client, ca *heistv1alpha1.VaultCertificateAuthority) (*heistv1alpha1.VaultCertificateAuthority, error) {
	potentialRoot := ca
	for potentialRoot.Spec.Issuer != "" {
		nextCA := &heistv1alpha1.VaultCertificateAuthority{
			ObjectMeta: metav1.ObjectMeta{
				Name:      potentialRoot.Spec.Issuer,
				Namespace: potentialRoot.Namespace,
			},
		}
		if err := k8s.Get(ctx, client.ObjectKeyFromObject(nextCA), nextCA); err != nil {
			return nil, err
		}
		potentialRoot = nextCA
	}
	return potentialRoot, nil
}
