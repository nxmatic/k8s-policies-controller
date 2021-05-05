.ONESHELL:
.SHELLFLAGS = -e -c

include make.d/macros.mk
include make.d/os.mk

controller-gen.bin := $(shell which controller-gen)
controller-gen.bin := $(if $(controller-gen.bin),$(controller-gen.bin),$(GOPATH)/bin/controller-gen)

make.d make.d/os.mk make.d/macros.mk:
	@: $(info loading git sub modules)
	git submodule init
	git submodule update

ifndef
NAMESPACE := jx
endif
ifndef VERSION
VERSION := $(shell git describe --tags| sed -r 's/^v//')
endif
VERSION_PKG := $(PKG)/pkg/version
VERSION_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
VERSION_TAG := v$(VERSION)

APP := $(if $(APP_NAME),$(APP_NAME),k8s-policies-controller)
BIN := $(APP)
PKG := github.com/nuxeo/$(APP)
OS := $(os)
ARCH := amd64
LD_FLAGS := -X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).buildDate=$(VERSION_DATE)
IMAGE := $(BIN)
BUILD_IMAGE := golang:1.15-buster


.PHONY: all
all: dev


.PHONY: release
release:
	@:

release: generate
release: release~kustomizes
release: release~binaries


.PHONY: release~binaries
release~binaries:
	@: $(info releasing the binaries for $(VERSION_TAG))
	git tag -a -m 'chore: release $(VERSION_TAG)' -f $(VERSION_TAG)
	goreleaser release --config=.goreleaser.yml --rm-dist

.PHONY: release~kustomizes
release~kustomizes:
	@: $(info versioning kustomizes@$(VERSION_TAG))
	[ -z "$$(git status -s)" ] || git commit -m 'chore: versioning kustomizes $(VERSION_TAG)' kustomizes
	git tag -f $(VERSION_TAG) && git push -f origin $(VERSION_TAG)

release~binaries: export GITHUB_TOKEN=$(GIT_TOKEN)
release~binaries: export REV=$(shell git rev-parse --short HEAD 2> /dev/null || echo 'unknown')
release~binaries: export BRANCH=$(BRANCH_NAME)
release~binaries: export BUILDDATE=$(VERSION_DATE)
release~binaries: export GOVERSION=$(shell go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
release~binaries: export export VERSION=$(VERSION_TAG)
release~binaries: export export ROOTPACKAGE = $(PKG)

.PHONY: generate
generate: controller-gen
generate: manifest.yaml

.PHONY: build
build:: generate compile

.PHONY: compile
compile:
	@: $(info compiling)
	go mod download
	GOOS=$(GOOS) GOARCH=${ARCH} go build -ldflags "$(LD_FLAGS)"

.PHONY: test
test:
	GOARCH=${ARCH} go test -v -ldflags "$(LD_FLAGS)" ./...

.PHONY: image
image:
	docker build -t $(IMAGE):latest -f Dockerfile .

image: $(BIN)

.PHONY: unkustomizes
unkustomizes: manifest.yaml
	kubectl delete -f manifest.yaml || true

.PHONY: kustomizes
kustomizes: manifest.yaml
	kubectl apply -f manifest.yaml


manifest.yaml: $(wildcard kustomizes/*.yaml) $(wildcard kustomizes/*/*.yaml)

manifest.yaml: | $(kustomize.bin) $(jx-cli.bin)
	@: $(info generating manifest for $(image):$(tag))
	(cd kustomizes/controller && kustomize edit set image k8s-policies-controller:latest=$(image):$(tag))
	$(kustomize.bin) cfg fmt kustomizes
	$(kustomize.bin) build kustomizes -o manifest.yaml

manifest.yaml: image:=$(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/$(DOCKER_REGISTRY_ORG)/$(IMAGE),$(IMAGE))
manifest.yaml: tag:=$(VERSION)


controller-gen: | $(controller-gen.bin) $(jx-cli.bin) $(kustomize.bin)

.PHONY: controller-gen
controller-gen:
	@: $(info generating controller descriptors)
	$(foreach package,$(packages),$(script))
	$(jx-cli.bin) gitops rename --dir=kustomizes

controller-gen: packages := gcpauth gcpworkload node meta
controller-gen: script=$(controller-gen.script)

define controller-gen.script =
	$(controller-gen.bin) object paths=./pkg/apis/$(package)/...
	$(controller-gen.bin) crd paths=./pkg/apis/$(package)/... output:crd:artifacts:config=kustomizes/$(package)
	$(controller-gen.bin) rbac:roleName=$(package)-controller paths=./pkg/apis/$(package)/... output:rbac:artifacts:config=kustomizes/$(package)
$(newline)
endef

# Run go fmt against code
fmt:
	go fmt ./...


.PHONY: up
up: image unkustomizes kustomizes

.PHONY: version
version:
	@echo $(VERSION_TAG)

VERSION::
	@echo v$(VERSION_TAG) > $(@)

null  :=
space := $(null) #
comma := ,

define newline :=

endef

# Run go vet against code
vet:
	go vet ./...

$(GOPATH)/bin/controller-gen:
	@: $(info building controller-gen)
	tmpdir=$$(mktemp -d)
	cd $$tmpdir
	go mod init tmp
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5
	rm -rf $$tmpdir
