# ====================================================================================
# Setup Project

PROJECT_NAME := provider-alibaba
PROJECT_REPO := github.com/crossplane-contrib/$(PROJECT_NAME)
IMG ?= crossplane/provider-alibaba:v1

# kind-related versions
KIND_VERSION ?= v0.11.1
KIND_NODE_IMAGE_TAG ?= v1.19.11

PLATFORMS ?= linux_amd64 linux_arm64
# -include will silently skip missing files, which allows us
# to load those files with a target in the Makefile. If only
# "include" was used, the make command would fail and refuse
# to run a target until the include commands succeeded.
-include build/makelib/common.mk

# ====================================================================================
# Setup Output

-include build/makelib/output.mk

# ====================================================================================
# Setup Go

# TODO(jastang): Upgrade Go version to be in-line with build submodule.
GO_REQUIRED_VERSION = 1.17

# Set a sane default so that the nprocs calculation below is less noisy on the initial
# loading of this file
NPROCS ?= 1

# each of our test suites starts a kube-apiserver and running many test suites in
# parallel can lead to high CPU utilization. by default we reduce the parallelism
# to half the number of CPU cores.
GO_TEST_PARALLEL := $(shell echo $$(( $(NPROCS) / 2 )))

GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/provider
GO_LDFLAGS += -X $(GO_PROJECT)/pkg/version.Version=$(VERSION)
GO_SUBDIRS += cmd pkg apis
GO111MODULE = on
-include build/makelib/golang.mk

# ====================================================================================
# Setup Kubernetes tools

UP_VERSION = v0.13.0
UP_CHANNEL = stable
-include build/makelib/k8s_tools.mk

# ====================================================================================
# Setup Images

IMAGES = provider-alibaba
-include build/makelib/imagelight.mk

# ====================================================================================
# Setup XPKG

XPKG_REG_ORGS ?= xpkg.upbound.io/crossplan-contrib index.docker.io/crossplanecontrib
# NOTE(hasheddan): skip promoting on xpkg.upbound.io as channel tags are
# inferred.
XPKG_REG_ORGS_NO_PROMOTE ?= xpkg.upbound.io/crossplane-contrib
XPKGS = provider-alibaba
-include build/makelib/xpkg.mk

# NOTE(hasheddan): we force image building to happen prior to xpkg build so that
# we ensure image is present in daemon.
xpkg.build.provider-alibaba: do.build.images

# ====================================================================================
# Targets

# run `make help` to see the targets and options

# We want submodules to be set up the first time `make` is run.
# We manage the build/ folder and its Makefiles as a submodule.
# The first time `make` is run, the includes of build/*.mk files will
# all fail, and this target will be run. The next time, the default as defined
# by the includes will be run instead.
fallthrough: submodules
	@echo Initial setup complete. Running make again . . .
	@make


# Generate a coverage report for cobertura applying exclusions on
# - generated file
cobertura:
	@cat $(GO_TEST_OUTPUT)/coverage.txt | \
		grep -v zz_generated.deepcopy | \
		$(GOCOVER_COBERTURA) > $(GO_TEST_OUTPUT)/cobertura-coverage.xml

crds.clean:
	@$(INFO) cleaning generated CRDs
	@find package/crds -name *.yaml -exec sed -i.sed -e '1,2d' {} \; || $(FAIL)
	@find package/crds -name *.yaml.sed -delete || $(FAIL)
	@$(OK) cleaned generated CRDs

generate: crds.clean

# integration tests
e2e.run: test-integration

# Run integration tests.
test-integration: $(KIND) $(KUBECTL) $(UP) $(HELM3)
	@$(INFO) running integration tests using kind $(KIND_VERSION)
	KIND_NODE_IMAGE_TAG=${KIND_NODE_IMAGE_TAG} $(ROOT_DIR)/cluster/local/integration_tests.sh || $(FAIL)
	@$(OK) integration tests passed

# Update the submodules, such as the common build scripts.
submodules:
	@git submodule sync
	@git submodule update --init --recursive

# NOTE(hasheddan): we must ensure up is installed in tool cache prior to build
# as including the k8s_tools machinery prior to the xpkg machinery sets UP to
# point to tool cache.
build.init: $(UP)

# This is for running out-of-cluster locally, and is for convenience. Running
# this make target will print out the command which was used. For more control,
# try running the binary directly with different arguments.
run: go.build
	@$(INFO) Running Crossplane locally out-of-cluster . . .
	@# To see other arguments that can be provided, run the command with --help instead
	$(GO_OUT_DIR)/$(PROJECT_NAME) --debug

manifests:
	@$(INFO) Deprecated. Run make generate instead.

.PHONY: cobertura submodules fallthrough run crds.clean manifests

# ====================================================================================
# Special Targets

define CROSSPLANE_MAKE_HELP
Crossplane Targets:
    cobertura             Generate a coverage report for cobertura applying exclusions on generated files.
    submodules            Update the submodules, such as the common build scripts.
    run                   Run crossplane locally, out-of-cluster. Useful for development.

endef
# The reason CROSSPLANE_MAKE_HELP is used instead of CROSSPLANE_HELP is because the crossplane
# binary will try to use CROSSPLANE_HELP if it is set, and this is for something different.
export CROSSPLANE_MAKE_HELP

crossplane.help:
	@echo "$$CROSSPLANE_MAKE_HELP"

help-special: crossplane.help

.PHONY: crossplane.help help-special

demo:
	docker build . -t ${IMG} -f ./hack/demo/Dockerfile
	kind load docker-image $(IMG) || { echo >&2 "kind not installed or error loading image: $(IMG)"; exit 1; }
	kubectl apply -f ./package/crds
	./hack/demo/prepare-alibaba-credentials.sh
	kubectl apply -f ./examples/provider.yaml
	./hack/demo/helm_install_crossplane_master.sh
	kubectl apply -f ./examples/database/rds.yaml
	kubectl apply -f ./hack/demo/deploy/
