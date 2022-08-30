package operator

import (
	"context"
	"fmt"
	"sync"

	"github.com/youniqx/heist/pkg/connector"
	"github.com/youniqx/heist/pkg/managed"
	"github.com/youniqx/heist/pkg/vault"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/kubernetesauth"
	"github.com/youniqx/heist/pkg/vault/policy"
	"golang.org/x/term"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	vaultForwardedPort = 27374
	setupTaskCount     = 2
)

type SetupConfig struct {
	VaultNamespace       string
	VaultServiceName     string
	VaultPort            string
	VaultURL             string
	VaultCAs             []string
	VaultToken           string
	KubernetesHost       string
	KubernetesJWTIssuer  string
	KubernetesJWTCACert  string
	KubernetesJWTPemKeys []string
	PolicyName           string
	RoleName             string
	HeistNamespace       string
	HeistServiceAccount  string
	RESTConfig           *rest.Config
	Quiet                bool
	VaultScheme          string
}

type setupManager struct {
	StepManager  *stepManager
	WaitGroup    sync.WaitGroup
	Config       *SetupConfig
	StepChannel  chan *Step
	ErrorChannel chan error
}

func SetupOperator(config *SetupConfig) error {
	stepChannel := make(chan *Step)

	setup := &setupManager{
		StepChannel:  stepChannel,
		ErrorChannel: make(chan error, 1),
		StepManager: &stepManager{
			CurrentStepName: "",
			Channel:         stepChannel,
		},
		Config:    config,
		WaitGroup: sync.WaitGroup{},
	}

	return setup.Run()
}

func (s *setupManager) Run() error {
	s.WaitGroup.Add(setupTaskCount)
	go s.runSetup()
	go s.printStepResults()
	s.WaitGroup.Wait()

	return <-s.ErrorChannel
}

func (s *setupManager) runSetup() {
	defer s.WaitGroup.Done()
	s.ErrorChannel <- s.performSetup()
}

func (s *setupManager) performSetup() error {
	defer s.StepManager.Complete()

	var vaultURL string
	if s.Config.VaultURL == "" {
		s.StepManager.NextStep("Fetch Vault Service")

		clientSet, err := kubernetes.NewForConfig(s.Config.RESTConfig)
		if err != nil {
			s.StepManager.StepFailed()
			return err
		}

		service, err := clientSet.CoreV1().Services(s.Config.VaultNamespace).Get(context.TODO(), s.Config.VaultServiceName, metav1.GetOptions{})
		if err != nil {
			s.StepManager.StepFailed()
			return err
		}

		s.StepManager.NextStep("Forward Port on Vault Pod")

		forwarder, err := connector.PortForward(s.Config.RESTConfig, service, []string{fmt.Sprintf("%d:%s", vaultForwardedPort, s.Config.VaultPort)})
		if err != nil {
			s.StepManager.StepFailed()
			return err
		}
		defer forwarder.Stop()

		vaultURL = fmt.Sprintf("%s://localhost:%d", s.Config.VaultScheme, vaultForwardedPort)
	} else {
		vaultURL = s.Config.VaultURL
	}

	s.StepManager.NextStep("Connect to Vault instance")

	cas := make([]core.StringSource, 0, len(s.Config.VaultCAs))
	for _, cert := range s.Config.VaultCAs {
		cas = append(cas, core.Value(cert))
	}

	api, err := vault.NewAPI().
		WithAddressFrom(core.Value(vaultURL)).
		WithCAsFrom(cas...).
		WithTokenFrom(core.Value(s.Config.VaultToken)).
		Complete()
	if err != nil {
		s.StepManager.StepFailed()
		return err
	}

	s.StepManager.NextStep("Write Heist Operator Policy")

	err = api.UpdatePolicy(&policy.Policy{
		Name:  s.Config.PolicyName,
		Rules: operatorPolicyRules,
	})
	if err != nil {
		s.StepManager.StepFailed()
		return err
	}

	s.StepManager.NextStep("Configure Kubernetes Auth Method")

	authMethodPath, err := managed.KubernetesAuth.GetMountPath()
	if err != nil {
		s.StepManager.StepFailed()
		return err
	}

	err = api.UpdateKubernetesAuthMethod(&kubernetesauth.Method{
		Path: authMethodPath,
		Config: &kubernetesauth.Config{
			KubernetesHost:       s.Config.KubernetesHost,
			Issuer:               s.Config.KubernetesJWTIssuer,
			PemKeys:              s.Config.KubernetesJWTPemKeys,
			KubernetesCACert:     s.Config.KubernetesJWTCACert,
			TokenReviewerJWT:     "",
			DisableISSValidation: false,
			DisableLocalCAJWT:    false,
		},
	})
	if err != nil {
		s.StepManager.StepFailed()
		return err
	}

	s.StepManager.NextStep("Write Heist Operator Kubernetes Auth Role")

	err = api.UpdateKubernetesAuthRole(managed.KubernetesAuth, &kubernetesauth.Role{
		Name:                 s.Config.RoleName,
		Policies:             []core.PolicyName{core.PolicyName(s.Config.PolicyName)},
		BoundNamespaces:      []string{s.Config.HeistNamespace},
		BoundServiceAccounts: []string{s.Config.HeistServiceAccount},
	})
	if err != nil {
		s.StepManager.StepFailed()
		return err
	}

	return nil
}

func (s *setupManager) printStepResults() {
	defer s.WaitGroup.Done()
	for step := range s.StepChannel {
		if s.Config.Quiet {
			continue
		}

		width, _, err := term.GetSize(0)
		if err != nil {
			continue
		}

		var text string
		switch step.Status {
		case StepStatusSuccess:
			text = fmt.Sprintf("\r[ DONE ] %s", step.Name)
		case StepStatusFailed:
			text = fmt.Sprintf("\r[FAILED] %s", step.Name)
		case StepStatusInProgress:
			text = fmt.Sprintf("\r[      ] %s", step.Name)
		}
		for i := 0; i < width-len(text); i++ {
			text += " "
		}

		switch step.Status {
		case StepStatusSuccess, StepStatusFailed:
			//nolint:forbidigo
			fmt.Println(text)
		case StepStatusInProgress:
			//nolint:forbidigo
			fmt.Print(text)
		}
	}
}
