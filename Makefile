# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= latest

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "preview,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=preview,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="preview,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# BUNDLE_IMAGE defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMAGE=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMAGE ?= controller-bundle:$(VERSION)

BINARY_PATH ?= bin/heist

# Image URL to use all building/pushing image targets
IMAGE_REPO ?= youniqx/heist
IMAGE ?= $(IMAGE_REPO):$(VERSION)

GIT_COMMIT=$(shell git rev-list -1 HEAD)
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
MODULE_NAME = github.com/youniqx/heist
LDFLAGS = -X '$(MODULE_NAME)/cmd.commit=$(GIT_COMMIT)' -X '$(MODULE_NAME)/cmd.version=$(VERSION)' -X '$(MODULE_NAME)/cmd.buildTime=$(BUILD_TIME)'

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd"

all: generate fix test kind release_dry_run

KIND = $(shell pwd)/bin/kind
bin/kind:
	$(call go-get-tool,$(KIND),sigs.k8s.io/kind)

kind/cluster: bin/kind
	$(KIND) create cluster --config=kind_config.yaml
	mkdir -p kind
	touch kind/cluster

KIND_EXTERNAL_REGISTRY = localhost:32000
KIND_INTERNAL_REGISTRY = kind.registry.svc.cluster.local:5000
kind/registry: kind/cluster
	kubectl --context kind-kind apply -f kind_registry.yaml
	kubectl --context kind-kind --namespace registry wait --for=condition=Available --timeout=120s deployment -l app.kubernetes.io/instance=kind
	touch kind/registry

kind/vault: kind/cluster
	helm repo add hashicorp https://helm.releases.hashicorp.com
	helm repo update
	helm upgrade vault hashicorp/vault \
		--kube-context kind-kind \
		--namespace vault \
		--create-namespace \
		--install --atomic \
		--set server.dev.enabled=true \
		--set server.logLevel=trace
	touch kind/vault

kind/cert-manager: kind/cluster
	helm repo add jetstack https://charts.jetstack.io
	helm repo update
	helm upgrade cert-manager jetstack/cert-manager \
	  --kube-context kind-kind \
	  --namespace cert-manager \
	  --create-namespace \
	  --version v1.3.1 \
	  --atomic --install \
	  --set installCRDs=true
	touch kind/cert-manager

KIND_IMAGE_TAG := $(shell date +%s)
KIND_HEIST_IMAGE = $(KIND_INTERNAL_REGISTRY)/heist:$(KIND_IMAGE_TAG)
kind/heist_image: kind/registry
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags '-w -extldflags "-static"' -o tmp/heist main.go
	docker build -t "$(KIND_EXTERNAL_REGISTRY)/heist:$(KIND_IMAGE_TAG)" --build-arg "BINARY_PATH=tmp/heist" .
	docker tag "$(KIND_EXTERNAL_REGISTRY)/heist:$(KIND_IMAGE_TAG)" "$(KIND_EXTERNAL_REGISTRY)/heist:latest"
	docker push "$(KIND_EXTERNAL_REGISTRY)/heist:$(KIND_IMAGE_TAG)"
	docker push "$(KIND_EXTERNAL_REGISTRY)/heist:latest"

kind/heist: kind/heist_image kind/vault kind/cert-manager install_manifests
	export KUBECONFIG=$$(echo "$${KUBECONFIG}" | awk -F ':' '{print $$1}') && go run . setup k8s \
		--vault-namespace vault \
		--vault-service vault \
		--vault-port 8200 \
		--vault-token root \
		--heist-service-account heist-operator \
		--kubernetes-jwt-issuer https://kubernetes.default.svc.cluster.local
	cd config/manager && $(KUSTOMIZE) edit set image "controller=$(KIND_HEIST_IMAGE)"
	$(KUSTOMIZE) build config/default | kubectl --context kind-kind --namespace heist-system apply -f -
	cd config/manager && $(KUSTOMIZE) edit set image "controller=youniqx/heist:latest"
	kubectl --context kind-kind --namespace heist-system wait --for=condition=Available --timeout=120s deployment -l control-plane=controller-manager

kind/test_namespace: kind/cluster
	kubectl --context kind-kind create namespace test || :
	kubectl --context kind-kind label namespace test heist.youniqx.com/inject-agent=true --overwrite
	kubectl --context kind-kind config set-context kind-kind --namespace test

kind/example: kind/test_namespace kind/heist
	kubectl --context kind-kind -n test apply -f demo

kind: kind/heist

clean:
	$(KIND) delete cluster || :
	rm -rf bin testbin dist kind

include mk/*.mk

GINKGO = $(shell pwd)/bin/ginkgo
bin/ginkgo:
	$(call go-get-tool,$(GINKGO),github.com/onsi/ginkgo/ginkgo)

test_setup: vault_test_setup kubernetes_test_setup bin/ginkgo

test_injector: test_setup
	$(GINKGO) -v ./pkg/injector -coverprofile cover-injector.out

test_agent: test_setup
	$(GINKGO) -v ./pkg/agent -coverprofile cover-agent.out

test_agentserver: test_setup
	$(GINKGO) -v ./pkg/agentserver -coverprofile cover-agentserver.out

test_webhooks: test_setup
	$(GINKGO) -v ./pkg/apis/heist.youniqx.com/v1alpha1 -coverprofile cover-webhooks.out

test_controllers: test_controllers_vaultbinding test_controllers_vaultcertificaterole test_controllers_vaultcertificateauthority test_controllers_vaultkvengine test_controllers_vaultkvsecret test_controllers_vaultsyncsecret test_controllers_vaulttransitengine test_controllers_vaulttransitkey

test_controllers_vaultbinding: test_setup
	$(GINKGO) -v ./pkg/controllers/e2e_test/vaultbinding -coverprofile cover-controllers-vaultbinding.out

test_controllers_vaultcertificaterole: test_setup
	$(GINKGO) -v ./pkg/controllers/e2e_test/vaultcertificaterole -coverprofile cover-controllers-vaultcertificaterole.out

test_controllers_vaultcertificateauthority: test_setup
	$(GINKGO) -v ./pkg/controllers/e2e_test/vaultcertificateauthority -coverprofile cover-controllers-vaultcertificateauthority.out

test_controllers_vaultkvengine: test_setup
	$(GINKGO) -v ./pkg/controllers/e2e_test/vaultkvengine -coverprofile cover-controllers-vaultkvengine.out

test_controllers_vaultkvsecret: test_setup
	$(GINKGO) -v ./pkg/controllers/e2e_test/vaultkvsecret -coverprofile cover-controllers-vaultkvsecret.out

test_controllers_vaultsyncsecret: test_setup
	$(GINKGO) -v ./pkg/controllers/e2e_test/vaultsyncsecret -coverprofile cover-controllers-vaultsyncsecret.out

test_controllers_vaulttransitengine: test_setup
	$(GINKGO) -v ./pkg/controllers/e2e_test/vaulttransitengine -coverprofile cover-controllers-vaulttransitengine.out

test_controllers_vaulttransitkey: test_setup
	$(GINKGO) -v ./pkg/controllers/e2e_test/vaulttransitkey -coverprofile cover-controllers-vaulttransitkey.out

test_vault: test_setup
	$(GINKGO) -v ./pkg/vault/e2e_test -coverprofile cover-vault.out

test: generate test_vault test_injector test_agent test_agentserver test_webhooks test_controllers

# Build manager binary
manager: generate
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate
	go run ./main.go

KUSTOMIZE = $(shell pwd)/bin/kustomize
download_kustomize:
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3)

# Install CRDs into a cluster
install_manifests: generate_manifests download_kustomize
	$(KUSTOMIZE) build config/crd | kubectl --context kind-kind apply -f -
	$(KUSTOMIZE) build config/crd | kubectl --context kind-kind apply -f -

# Uninstall CRDs from a cluster
uninstall_manifests: generate_manifests download_kustomize
	$(KUSTOMIZE) build config/crd | kubectl --context kind-kind delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: generate_manifests download_kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMAGE)
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
print_manifests: generate_manifests download_kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMAGE)
	$(KUSTOMIZE) build config/default

# UnDeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy: download_kustomize
	$(KUSTOMIZE) build config/default | kubectl delete -f -

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
download_controller-gen:
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen)

# Generate manifests e.g. CRD, RBAC etc.
generate_manifests: download_controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Generate code
generate_controller: download_controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

generate: generate_controller generate_manifests generate_client format

CLIENT_GEN = $(shell pwd)/bin/client-gen
download_client_gen:
	$(call go-get-tool,$(CLIENT_GEN),k8s.io/code-generator/cmd/client-gen)

LISTER_GEN = $(shell pwd)/bin/lister-gen
download_lister_gen:
	$(call go-get-tool,$(LISTER_GEN),k8s.io/code-generator/cmd/lister-gen)

generate_client: download_client_gen download_lister_gen
	./hack/update-clientset.sh v1alpha1

GORELEASER = $(shell pwd)/bin/goreleaser
bin/goreleaser:
	$(call go-get-tool,$(GORELEASER),github.com/goreleaser/goreleaser)

build: release_dry_run

release_dry_run: bin/goreleaser
	$(GORELEASER) release --rm-dist --snapshot --skip-publish

install:
	go install -ldflags "$(LDFLAGS)"

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags '-w -extldflags "-static"' -o tmp/heist main.go
	docker build -t "$(IMAGE)" --build-arg "BINARY_PATH=tmp/heist" .

# Push the docker image
docker-push:
	docker push "$(IMAGE)"

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: generate_manifests download_kustomize
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMAGE)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMAGE) .
