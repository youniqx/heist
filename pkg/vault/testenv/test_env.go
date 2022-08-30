package testenv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/vault"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/transit"
)

type TestEnv interface {
	ControlPlane
	Assertions
}

type ControlPlane interface {
	GetAddress() string
	GetRootToken() string
	WritePolicy(name string, policyFilePath string) error
	CreateToken(policies ...string) (string, error)
	Stop() error
	GetAPI() (vault.API, error)
}

type Assertions interface {
	KvSecret(engine core.MountPathEntity, secret core.SecretPathEntity) gomega.AsyncAssertion
	KvEngine(engine core.MountPathEntity) gomega.AsyncAssertion
	TransitEngine(engine core.MountPathEntity) gomega.AsyncAssertion
	TransitKey(engine core.MountPathEntity, key transit.KeyNameEntity) gomega.AsyncAssertion
	Policy(policy core.PolicyNameEntity) gomega.AsyncAssertion
	KubernetesAuthRole(auth core.MountPathEntity, role core.RoleNameEntity) gomega.AsyncAssertion
	AuthMethod(auth core.MountPathEntity) gomega.AsyncAssertion
	Mount(mount core.MountPathEntity) gomega.AsyncAssertion
	CA(mount core.MountPathEntity) gomega.AsyncAssertion
	TuneConfig(engine core.MountPathEntity) gomega.AsyncAssertion
	CertificateRole(ca core.MountPathEntity, role core.RoleNameEntity) gomega.AsyncAssertion
}

type testEnv struct {
	Address           string
	Port              int
	RootToken         string
	BinaryPath        string
	VaultServeCommand *exec.Cmd
	RootAPI           vault.API
}

const (
	maxStartAttempts            = 2
	maxStatusPollAttempts       = 40
	httpTimeout                 = 10 * time.Second
	afterStopSleepDuration      = 500 * time.Millisecond
	statusPollInterval          = 250 * time.Millisecond
	vaultProcessCleanupInterval = 200 * time.Millisecond
	vaultProcessCleanupTimeout  = 60 * time.Second
	handleFolderPerm            = 0o700
	handleFilePerm              = 0o600
)

func StartTestEnv(port int) (TestEnv, error) {
	binaryPath, err := findVaultBinary()
	if err != nil {
		return nil, err
	}

	env := &testEnv{
		Address:    fmt.Sprintf("127.0.0.1:%d", port),
		Port:       port,
		RootToken:  "root",
		BinaryPath: binaryPath,
	}

	if err := env.StopPreviousInstanceHandle(); err != nil {
		return nil, err
	}

	if err := env.Start(); err != nil {
		return nil, err
	}

	env.RootAPI, err = vault.NewAPI().
		WithAddressFrom(core.Value(env.GetAddress())).
		WithTokenFrom(core.Value(env.GetRootToken())).
		Complete()
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (e *testEnv) WritePolicy(name string, policyFilePath string) error {
	_, _, err := e.RunVaultCommand("policy", "write", name, policyFilePath)
	return err
}

type authResponse struct {
	Auth authResponseData `json:"auth"`
}

type authResponseData struct {
	ClientToken string `json:"client_token"`
}

func (e *testEnv) CreateToken(policies ...string) (string, error) {
	args := make([]string, 0, len(policies)+1)
	args = append(args, "token", "create", "-format=json")

	for _, policy := range policies {
		args = append(args, fmt.Sprintf("-policy=%s", policy))
	}

	out, _, err := e.RunVaultCommand(args...)
	if err != nil {
		return "", err
	}

	response := &authResponse{}
	if err := json.Unmarshal(out, response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Auth.ClientToken, nil
}

func (e *testEnv) EnableK8sAuth() error {
	address := fmt.Sprintf("http://%s/v1/sys/auth/kubernetes", e.Address)

	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, address, strings.NewReader("{\"path\":\"kubernetes\",\"type\":\"kubernetes\",\"config\":{}}"))
	if err != nil {
		return fmt.Errorf("failed to create k8s auth request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("X-Vault-Token", e.RootToken)

	c := http.Client{}

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to enable k8s auth: %w", err)
	}

	defer resp.Body.Close()

	return nil
}

func (e *testEnv) RunVaultCommand(args ...string) (stdOut []byte, stdErr []byte, err error) {
	// #nosec G204
	cmd := exec.Command(e.BinaryPath, args...)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("VAULT_ADDR=%s", e.GetAddress()),
		fmt.Sprintf("VAULT_TOKEN=%s", e.RootToken),
	)
	stdOutWriter := &splitWriter{
		target: os.Stdout,
	}
	stdErrWriter := &splitWriter{
		target: os.Stderr,
	}
	cmd.Stdout = stdOutWriter
	cmd.Stderr = stdErrWriter

	err = cmd.Run()

	return stdOutWriter.buffer.Bytes(), stdErrWriter.buffer.Bytes(), err
}

type splitWriter struct {
	target io.Writer
	buffer bytes.Buffer
}

func (s *splitWriter) Write(p []byte) (n int, err error) {
	s.buffer.Write(p)
	return s.target.Write(p)
}

func findVaultBinary() (string, error) {
	testVaultPath := os.Getenv("TEST_VAULT_BINARY_PATH")
	if testVaultPath != "" {
		info, err := os.Stat(testVaultPath)
		if err != nil {
			return "", fmt.Errorf("failed to check if vault binary exists at path %s: %w", testVaultPath, err)
		}

		if info.IsDir() {
			return "", core.ErrAPIError.WithDetails(fmt.Sprintf("found directory at expected test vault path: %s", testVaultPath))
		}

		return testVaultPath, nil
	}

	path, err := exec.LookPath("vault")
	if err != nil {
		return "", fmt.Errorf("failed to lookup vault in path: %w", err)
	}

	return path, nil
}

func (e *testEnv) StopPreviousInstanceHandle() error {
	handlePath := e.getHandlePath()

	data, err := os.ReadFile(handlePath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to check if handle file exists: %w", err)
	}

	processID, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("parsing handle data failed: %w", err)
	}

	deadline := time.Now().Add(vaultProcessCleanupTimeout)
	if err := killProcessAndWaitForItToExit(deadline, processID); err != nil {
		return fmt.Errorf("failed to kill old vault processes: %w", err)
	}

	return os.Remove(handlePath)
}

func (e *testEnv) getHandlePath() string {
	homeDir := os.Getenv("HOME")
	handlePath := filepath.Join(homeDir, ".vault-test-env", "handles", strconv.Itoa(e.Port))

	return handlePath
}

var errCleanupTimeout = errors.New("cleanup took too long")

func killProcessAndWaitForItToExit(deadline time.Time, id int) error {
	for time.Now().Before(deadline) {
		process, err := os.FindProcess(id)
		if err != nil {
			//nolint:nilerr
			return nil
		}

		if err := process.Kill(); err != nil {
			//nolint:nilerr
			return nil
		}

		time.Sleep(vaultProcessCleanupInterval)
	}

	return errCleanupTimeout
}

func (e *testEnv) Start() error {
	handlePath := e.getHandlePath()

	tryStartVault := true
	for loopCount := 0; tryStartVault; loopCount++ {
		// #nosec G204
		e.VaultServeCommand = exec.Command(e.BinaryPath, "server", "-dev",
			fmt.Sprintf("-dev-root-token-id=%s", e.RootToken),
			fmt.Sprintf("-dev-listen-address=%s", e.Address),
		)
		e.VaultServeCommand.Stdout = os.Stdout
		e.VaultServeCommand.Stderr = os.Stderr

		if err := e.VaultServeCommand.Start(); err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(handlePath), handleFolderPerm); err != nil {
			return err
		}

		processID := strconv.Itoa(e.VaultServeCommand.Process.Pid)
		if err := ioutil.WriteFile(handlePath, []byte(processID), handleFilePerm); err != nil {
			return err
		}

		for i := 0; true; i++ {
			time.Sleep(statusPollInterval)

			_, _, err := e.RunVaultCommand("status")
			if err == nil {
				tryStartVault = false
				break
			}

			if i >= maxStatusPollAttempts {
				if loopCount >= maxStartAttempts-1 {
					return core.ErrAPIError.WithDetails("vault is not starting properly")
				}

				break
			}
		}
	}

	return e.EnableK8sAuth()
}

func (e *testEnv) GetAddress() string {
	return fmt.Sprintf("http://%s", e.Address)
}

func (e *testEnv) GetRootToken() string {
	return e.RootToken
}

func (e *testEnv) Stop() error {
	defer time.Sleep(afterStopSleepDuration)
	return e.VaultServeCommand.Process.Kill()
}

func (e *testEnv) GetAPI() (vault.API, error) {
	return e.RootAPI, nil
}
