# Directory to store executables
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN): ## Create local bin directory if necessary.
	mkdir -p $(LOCALBIN)

# KWOK repository
KWOK_REPO=kubernetes-sigs/kwok
# Get latest
KWOK_RELEASE=v0.4.0

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ kwok

.PHONY: kwok-run
kwok-run: kwok ## run kwok
	$(KWOK) 						   \
      --kubeconfig=~/.kube/config 	   \
      --manage-all-nodes=false 	  	   \
      --manage-nodes-with-annotation-selector=kwok.x-k8s.io/node=fake 		  \
      --disregard-status-with-annotation-selector=kwok.x-k8s.io/status=custom \
      --cidr=10.0.0.1/24 			   \
      --node-ip=10.0.0.1 			   \
      --node-lease-duration-seconds=40

.PHONY: kwok-install-operator
kwok-install-operator: ## install kwok operator
	@echo "Installing kwok version $(KWOK_RELEASE)..."
	@kubectl apply -f "https://github.com/$(KWOK_REPO)/releases/download/$(KWOK_RELEASE)/kwok.yaml"

.PHONY: kwok-uninstall-operator
kwok-uninstall-operator: ## uninstall kwok operator
	@echo "Uninstalling kwok operator..."
	@kubectl delete -f "https://github.com/$(KWOK_REPO)/releases/download/$(KWOK_RELEASE)/kwok.yaml"

.PHONY: kwok-install-stages
kwok-install-stages: ## install kwok stages
	@echo "Setting up kwok stages..."
	@kubectl apply -f internal/simulator/data/stages.yaml

.PHONY: kwok-uninstall-stages
kwok-uninstall-stages: ## uninstall kwok stages
	@echo "Uninstalling kwok stages..."
	@kubectl delete -f internal/simulator/data/stages.yaml

##@ Simulation

.PHONY: create-fake-node
create-fake-node: ## create fake node managed by kwok
	@echo "Creating fake node..."
	@kubectl apply -f hack/fake-node.yaml

.PHONY: delete-fake-node
delete-fake-node: ## delete fake node
	@echo "Deleting fake node..."
	@kubectl delete -f hack/fake-node.yaml

.PHONY: create-fake-pod
create-fake-pod: ## create fake pod managed by kwok
	@echo "Creating fake pod..."
	@kubectl apply -f hack/fake-pod.yaml

.PHONY: delete-fake-pod
delete-fake-pod: ## delete fake pod
	@echo "Deleting fake pod..."
	@kubectl delete -f hack/fake-pod.yaml

##@ Build

.PHONY: build
build: ## build simulator binary
	@echo "Building binary..."
	@go build -o bin/batchsim cmd/simulator/main.go

##@ Lint

.PHONY: lint
lint: golangci-lint ## lint code using golangci-lint
	@echo "Running golangci-lint..."
	@$(GOLANGCI_LINT) run --timeout=5m

.PHONY: lint-fix
lint-fix: golangci-lint ## lint code and fix issues using golangci-lint
	@echo "Running golangci-lint with fix enabled..."
	@$(GOLANGCI_LINT) run --fix --timeout=5m

##@ Test

.PHONY: test
test: ## run all tests
	@echo "Running all tests..."
	@$(MAKE) test-unit
	@$(MAKE) test-integration

.PHONY: test-unit
test-unit: gotestsum ## run unit tests
	@echo "Running unit tests..."
	@$(GOTESTSUM) 				   			   \
		--format short-verbose 	   			   \
		--junitfile test-output/unit-tests.xml \
		--jsonfile test-output/unit-tests.json \
		-- -coverprofile=test-output/coverage.out -covermode=atomic ./...

test-integration: gotestsum ## run integration tests
	@echo "Running integration tests..."
	@INTEGRATION_TEST=true $(GOTESTSUM)   			  \
		--format short-verbose		      			  \
		--junitfile test-output/integration-tests.xml \
		--jsonfile test-output/integration-tests.json \
		-- -run _Integration -coverprofile=test-output/coverage.out -covermode=atomic ./...

##@ Documentation

.PHONY: docs
docs: ## generate documentation
	@echo "Generating documentation..."
	@go run cmd/simulator/main.go docgen

##@ Metrics

.PHONY: metrics-server-install
metrics-server-install: ## install metrics server using helm
	@echo "Installing metrics server..."
	@kubectl apply --filename https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

.PHONY: metrics-server-uninstall
metrics-server-uninstall: ## uninstall metrics server using helm
	@echo "Uninstalling metrics server..."
	@kubectl delete --filename https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

##@ Dashboard

.PHONY: dashboard-install
dashboard-install: ## install dashboard
	@echo "Installing dashboard..."
	@kubectl apply -filename https://raw.githubusercontent.com/kubernetes/dashboard/v2.3.1/aio/deploy/recommended.yaml

.PHONY: dashboard-uninstall
dashboard-uninstall: ## uninstall dashboard
	@echo "Uninstalling dashboard..."
	@kubectl delete -filename https://raw.githubusercontent.com/kubernetes/dashboard/v2.3.1/aio/deploy/recommended.yaml

.PHONY: dashboard-create-admin-user
dashboard-create-admin-user: ## create admin user for dashboard
	@echo "Creating admin user..."
	@kubectl apply -filename hack/admin-user.yaml

.PHONY: delete-admin-user
dashboard-delete-admin-user: ## delete admin user
	@echo "Deleting admin user..."
	@kubectl delete -f hack/admin-user.yaml

dashboard-create-token: ## create dashboard login token for admin user
	@echo "Creating login token..."
	@kubectl --namespace kubernetes-dashboard create token admin-user

.PHONY: get-token
dashboard-get-token: dashboard-create-admin-user ## get dashboard login token for admin user
	@echo "Getting login token..."
	@kubectl get secret admin-user --namespace kubernetes-dashboard -o jsonpath={".data.token"} | base64 -d

.PHONY: dashboard
dashboard: ## open dashboard
	@echo "Opening dashboard..."
	@open http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/

##@ Misc

.PHONY: kube-proxy
kube-proxy: ## run kube proxy
	@echo "Running kube proxy..."
	@kubectl proxy

##@ External Dependencies

GOTESTSUM ?= $(LOCALBIN)/gotestsum
.PHONY: gotestsum
gotestsum: $(GOTESTSUM) ## download gotestsum locally if necessary.
$(GOTESTSUM): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install gotest.tools/gotestsum@v1.11.0

KWOK ?= $(LOCALBIN)/kwok
.PHONY: kwok
kwok: $(KWOK) ## download kwok locally if necessary.
$(KWOK): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kwok/cmd/kwok@$(KWOK_RELEASE)

GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2

.PHONY: install-cobra-cli
cobra-cli: ## install cobra cli globally
	@echo "Installing cobra cli..."
	@go install github.com/spf13/cobra/cobra-cli@latest

.PHONY: install-kube-state-metrics
install-kube-state-metrics: ## install kube-state-metrics using helm
	@echo "Installing kube-state-metrics..."
	@kubectl apply -f
	@helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	@helm repo update
	@helm install kube-state-metrics prometheus-community/kube-state-metrics --create-namespace --namespace monitoring

uninstall-kube-state-metrics: ## uninstall kube-state-metrics using helm
	@echo "Uninstalling kube-state-metrics..."
	@helm uninstall kube-state-metrics --namespace monitoring
