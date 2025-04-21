KUBECTL_CONTEXT ?= colima-k3s
KUBECONTEXT ?= colima-k3s
# Image URL to use all building/pushing image targets
TAG ?= v0.0.1
IMG ?= k8s.tochka.com/multidc-pdb-operator:${TAG}
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:ignoreUnexportedFields=true"

# Get the currentIMGly used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: manifests generate fmt vet ## Run tests.
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

##@ Build

build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go --cert-dir=/tmp/pki

docker-build: test ## Build docker image with the manager.
	docker build -t ${IMG} .

docker-push: ## Push docker image with the manager.
	docker push ${IMG}

chart-build: manifests kustomize
	helm dependency update $(shell pwd)/chart
	$(KUSTOMIZE) build config/crd > $(shell pwd)/chart/templates/crd.yaml
	$(KUSTOMIZE) build config/rbac > $(shell pwd)/chart/templates/rbac.yaml

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl --context $(KUBECTL_CONTEXT) apply -f -

install-samples: manifests ## Install Instances of Custom Resources
	kubectl --context $(KUBECTL_CONTEXT) apply -f $(shell pwd)/config/samples/

install-certificate: ## Install admission
	kubectl --context $(KUBECTL_CONTEXT) apply -f $(shell pwd)/chart/templates/certificate.yaml

install-admission: ## Install admission
	kubectl --context $(KUBECTL_CONTEXT) apply -f $(shell pwd)/deploy/admission.yaml

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl --context $(KUBECTL_CONTEXT) delete -f -

#deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
#	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
#	$(KUSTOMIZE) build config/default | kubectl --context $(KUBECTL_CONTEXT) apply -f -

deploy-rbac: manifests kustomize
	cd config/rbac && $(KUSTOMIZE) edit set namespace default
	$(KUSTOMIZE) build config/rbac | kubectl --context $(KUBECTL_CONTEXT) -n default apply -f -

get-rbac:
#	kubectl --context $(KUBECTL_CONTEXT) -n default get secret $(shell kubectl --context $(KUBECTL_CONTEXT) -n default get secrets  | ggrep -o -P "k8s-multidc-pdb-controller-manager-token-\w+") -o json | jq -r .data.token | base64 -dD && echo
	kubectl --context $(KUBECTL_CONTEXT) -n default get secret $(shell kubectl --context $(KUBECTL_CONTEXT) -n default get secrets  | ggrep -o -P "k8s-multidc-pdb-controller-manager-token-\w+") --template='{{.data.token | base64decode}}' && echo

#undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
#	$(KUSTOMIZE) build config/default | kubectl --context $(KUBECTL_CONTEXT) delete -f -

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	GOBIN=$(PROJECT_DIR)/bin GO111MODULE=on go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	GOBIN=$(PROJECT_DIR)/bin GO111MODULE=on go install sigs.k8s.io/kustomize/kustomize/v4@latest

repo_update:
	helm repo update

deploy: repo_update
	helm --kube-context $(KUBECTL_CONTEXT) -n default upgrade -i -f "deploy/values.yaml,deploy/values.$(KUBECTL_CONTEXT).yaml" \
		--set "applications.k8s-multidc-pdb-operator.containers.k8s-multidc-pdb-operator.imageTag=${TAG}" \
		k8s-multidc-pdb-operator devexp/hell

deploy-testapp: repo_update
	helm --kube-context $(KUBECTL_CONTEXT) -n default upgrade -i -f deploy/test/values.yaml multidcdpb-testz devexp/hell

undeploy:
	helm --kube-context $(KUBECTL_CONTEXT) -n default uninstall k8s-multidc-pdb-operator

undeploy-testapp:
	helm --kube-context $(KUBECTL_CONTEXT) -n default uninstall multidcdpb-testz

helm_meta_patch = '{"metadata":{"labels":{"app.kubernetes.io/managed-by": "Helm"},"annotations":{"meta.helm.sh/release-name": "k8s-multidc-pdb-operator", "meta.helm.sh/release-namespace": "default"}}}'
patch-helm:
	kubectl --context $(KUBECTL_CONTEXT) -n default patch serviceaccounts k8s-multidc-pdb-controller-manager -p $(helm_meta_patch)
	kubectl --context $(KUBECTL_CONTEXT) -n default patch clusterrole k8s-multidc-pdb-manager-role -p $(helm_meta_patch)
	kubectl --context $(KUBECTL_CONTEXT) -n default patch clusterrolebindings k8s-multidc-pdb-manager-rolebinding -p $(helm_meta_patch)
	kubectl --context $(KUBECTL_CONTEXT) -n default patch ValidatingWebhookConfiguration multidc-pdb.admission.tochka.com -p $(helm_meta_patch)
	#kubectl --context $(KUBECTL_CONTEXT) -n default delete certificate k8s-multidc-pdb-operator.default.gateway

# export helm_meta_patch='{"metadata":{"labels":{"app.kubernetes.io/managed-by": "Helm"},"annotations":{"meta.helm.sh/release-name": "k8s-multidc-pdb-operator", "meta.helm.sh/release-namespace": "default"}}}'
# kubectl -n default patch CustomResourceDefinition multidcpoddisruptionbudgets.k8s.tochka.com -p "${helm_meta_patch}"

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef