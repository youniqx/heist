
KUBERNETES_ENV_VERSION = v0.7.0
KUBERNETES_ENV_CHECKSUM = 42caee86cd41ef08dac1d4e1464d7b6503444e4376934eb20894bb1758251c67

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
export KUBEBUILDER_ASSETS = $(ENVTEST_ASSETS_DIR)/bin
KUBERNETES_ENV_SCRIPT_DIRECTORY = $(ENVTEST_ASSETS_DIR)/scripts
KUBERNETES_ENV_SCRIPT_PATH = $(KUBERNETES_ENV_SCRIPT_DIRECTORY)/kube-env-$(KUBERNETES_ENV_VERSION)-$(KUBERNETES_ENV_CHECKSUM)

kubernetes_test_setup:
	@mkdir -p "$(KUBERNETES_ENV_SCRIPT_DIRECTORY)"
	@test -f "$(KUBERNETES_ENV_SCRIPT_PATH)" || curl -fLo "$(KUBERNETES_ENV_SCRIPT_PATH)" https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/$(KUBERNETES_ENV_VERSION)/hack/setup-envtest.sh
	@echo "$(KUBERNETES_ENV_CHECKSUM)  $(KUBERNETES_ENV_SCRIPT_PATH)" | sha256sum -c
	@bash -c "source \"$(KUBERNETES_ENV_SCRIPT_PATH)\"; export ENVTEST_K8S_VERSION=1.24.1; fetch_envtest_tools \"$(ENVTEST_ASSETS_DIR)\";"
