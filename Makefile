# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: tidy
tidy: ## Run go vet against code.
	go mod tidy

##@ Build

.PHONY: build
build: fmt vet ## Build manager binary.
	go build -o bin/argo-cd-toolkit main.go

.PHONY: run
run: fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: generate
generate: protoc
	# protoc --proto_path=./pkg/config/proto --go_out=./pkg/config/v1alpha1 --openapiv2_out=logtostderr=true:./schema/swagger ./pkg/config/proto/cluster-config.proto
	protoc --proto_path=./pkg/config/proto --go_out=./pkg/config/v1alpha1 ./pkg/config/proto/cluster-config.proto
	go run schema/main.go

.PHONY: protoc
protoc: tidy
	test -s $(GOBIN)/protoc-gen-go || go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
        google.golang.org/protobuf/cmd/protoc-gen-go \
        google.golang.org/grpc/cmd/protoc-gen-go-grpc