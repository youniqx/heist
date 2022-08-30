VAULT_CHECKSUM_1_9_3_darwin_amd64 = 32614b830aae76cd42d6142cca86d6bc88d4951d505e2e2b39c2e03396e04e4c
VAULT_CHECKSUM_1_9_3_darwin_arm64 = 36ee0bb83b29842100960fd6a89b826bea5d09d272f4bdeeddae11fcbbf2643e
VAULT_CHECKSUM_1_9_3_freebsd_386 = 98dbb43a9836a035745d8baf8d62553069d928d3f080dbe5fbffe1a53e2b8cd6
VAULT_CHECKSUM_1_9_3_freebsd_amd64 = d1175ea83732ceea90bd1e225c7a93efb44adbaef91f2dad1af786c80e89545a
VAULT_CHECKSUM_1_9_3_freebsd_arm = fa1be3b562ce30bfb21e17e9835d1b3c0789aec0a8baf7e2aa2ed898a0477541
VAULT_CHECKSUM_1_9_3_linux_386 = 8b21befe285cd4a60d6e6c1d2a4c44d0d4e7ca5cfbab2b1f579734c57dd078eb
VAULT_CHECKSUM_1_9_3_linux_amd64 = 16059f245fb4df2800fe6ba320ea66aba9c2615348e37bcfd42754591a528639
VAULT_CHECKSUM_1_9_3_linux_arm = babbedfa3f134fbd68a4e1bb48e7302b9e40e8eae393f5f3688e295a7ea37f4f
VAULT_CHECKSUM_1_9_3_linux_arm64 = c420f14b8b712197c8c47852ea3d1a5976e9ceaf5bb8e6a0b311624111aa14d4
VAULT_CHECKSUM_1_9_3_netbsd_386 = 1327148eeb927ebd1cf4063fef63408943285f8767bd81cef65d3b8b04a7bf01
VAULT_CHECKSUM_1_9_3_netbsd_amd64 = b11fa765038adb0534e9edb1547fd379e7ca3aa53727b32b82346c1374a68625
VAULT_CHECKSUM_1_9_3_netbsd_arm = a38af7217c602a27691c56bca43709a8ae3c92c97db1a72e660ff2fb68954595
VAULT_CHECKSUM_1_9_3_openbsd_386 = 7db17153e11b8db7c1abf550a6e14d42076c56b91fd64d1219b7f37f461ba34b
VAULT_CHECKSUM_1_9_3_openbsd_amd64 = 3f3cfee19e827582a74f9b301562cf447bcdd2ae4cbdfda1a2b2dbb79de64eb8
VAULT_CHECKSUM_1_9_3_openbsd_arm = 7560688b8e5875a09ba074b561af279c16d785a698995f1f3f9eb6ba29510c4b
VAULT_CHECKSUM_1_9_3_solaris_amd64 = d279262a644921fa1f4e68d3ca07941154ec64fdc062a625d2162fcfc8f2dd0f
VAULT_CHECKSUM_1_9_3_windows_386 = 9267d9ddf9462fd46a9c45fda4d9ddd57736b762e44f1e54888aeedafe0b18d5
VAULT_CHECKSUM_1_9_3_windows_amd64 = 0c43af1330a1df11811eea36d9c0bc8ae9ae03d3ac98baa263a14b6fc8eb0cfe

VAULT_VERSION = 1.9.3
VAULT_ARCH = $(shell  go env GOARCH)
VAULT_OS = $(shell go env GOOS)

ifeq ($(OS),Windows_NT)
	VAULT_BINARY=vault.exe
else
	VAULT_BINARY=vault
endif

ENVTEST_ASSETS_DIR = $(shell pwd)/testbin
VAULT_CHECKSUM := $(VAULT_CHECKSUM_$(shell echo '$(VAULT_VERSION)' | sed 's/\./_/g')_$(VAULT_OS)_$(VAULT_ARCH))
VAULT_ARTIFACT_ID = $(VAULT_VERSION)-$(VAULT_OS)-$(VAULT_ARCH)-$(VAULT_CHECKSUM)
export TEST_VAULT_BINARY_PATH = $(ENVTEST_ASSETS_DIR)/bin/$(VAULT_BINARY)

vault_test_setup:
	@mkdir -p "$(ENVTEST_ASSETS_DIR)/archives"
	@mkdir -p "$(ENVTEST_ASSETS_DIR)/versions"
	@mkdir -p "$(ENVTEST_ASSETS_DIR)/bin"
	@test -f "$(ENVTEST_ASSETS_DIR)/archives/vault-$(VAULT_ARTIFACT_ID).zip" || curl -fLo "$(ENVTEST_ASSETS_DIR)/archives/vault-$(VAULT_ARTIFACT_ID).zip" "https://releases.hashicorp.com/vault/$(VAULT_VERSION)/vault_$(VAULT_VERSION)_$(VAULT_OS)_$(VAULT_ARCH).zip"
	@echo "$(VAULT_CHECKSUM)  $(ENVTEST_ASSETS_DIR)/archives/vault-$(VAULT_ARTIFACT_ID).zip" | sha256sum -c
	@test -f "$(ENVTEST_ASSETS_DIR)/versions/vault-$(VAULT_ARTIFACT_ID)" || bash -c 'temp="$$(mktemp -d)" && unzip "$(ENVTEST_ASSETS_DIR)/archives/vault-$(VAULT_ARTIFACT_ID).zip" -d "$${temp}" && mv "$${temp}/$(VAULT_BINARY)" "$(ENVTEST_ASSETS_DIR)/versions/vault-$(VAULT_ARTIFACT_ID)"'
	@ln -f -s "$(ENVTEST_ASSETS_DIR)/versions/vault-$(VAULT_ARTIFACT_ID)" "$(ENVTEST_ASSETS_DIR)/bin/$(VAULT_BINARY)"
	@echo "# Vault Binary (Version $(VAULT_VERSION)) ready at path: $(ENVTEST_ASSETS_DIR)/bin/$(VAULT_BINARY)"
