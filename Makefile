.ONESHELL:

include make.d/macros.mk
include make.d/os.mk

controller-gen.bin := $(shell which controller-gen)
controller-gen.bin := $(if $(controller-gen.bin),$(controller-gen.bin),$(GOPATH)/bin/controller-gen)

make.d make.d/os.mk make.d/macros.mk:
	@: $(info loading git sub modules)
	git submodule init
	git submodule update

BIN := k8s-policies-controller
CRD_OPTIONS ?= "crd:trivialVersions=true"
PKG := github.com/nuxeo/k8s-policies-controller
ARCH ?= amd64
APP ?= k8s-policies-controller
NAMESPACE ?= default
RELEASE_NAME ?= k8s-policies-controller
KO_DOCKER_REPO = registry.softonic.io/k8s-policies-controller
REPOSITORY ?= gcr.io/build-jx-prod/library
VERSION ?= 0.0.0-SNAPHOT
VERSION_PKG ?= $(PKG)/pkg/version
VERSION_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LD_FLAGS := -X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).buildDate=$(VERSION_DATE)

IMAGE := $(BIN)

BUILD_IMAGE ?= golang:1.15-buster


.PHONY: all
all: dev

.PHONY: generate
generate: controller-gen
generate: manifest.yaml

.PHONY: build
build: generate compile 

.PHONY: compile
compile:
	go mod download
	GOARCH=${ARCH} go build -ldflags "$(LD_FLAGS)"

.PHONY: test
test:
	GOARCH=${ARCH} go test -v -ldflags "$(LD_FLAGS)" ./...

.PHONY: image
image:
	docker build -t $(IMAGE):latest -f Dockerfile .

.PHONY: dev
dev: image
	kind load docker-image $(IMAGE):$(VERSION)

.PHONY: unkustomizes
unkustomizes: manifest.yaml
	kubectl delete -f manifest.yaml || true

.PHONY: kustomizes
kustomizes: manifest.yaml
	kubectl apply -f manifest.yaml


ifdef DOCKER_REGISTRY
manifest.yaml: image:=$(DOCKER_REGISTRY)/$(DOCKER_REGISTRY_ORG)/$(IMAGE)
else
manifest.yaml: image:=$(IMAGE)
endif

manifest.yaml: tag:=$(VERSION) 

manifest.yaml: $(wildcard kustomizes/*.yaml) $(wildcard kustomizes/*/*.yaml)

manifest.yaml: | $(kustomize.bin) $(jx-cli.bin)
	@: $(info generating manifest for $(image):$(tag))
	(cd kustomizes/controller && kustomize edit set image k8s-policies-controller:latest=$(image):$(tag))
	$(kustomize.bin) cfg fmt kustomizes
	$(kustomize.bin) build kustomizes -o manifest.yaml
	
controller-gen: packages := gcpauth gcpworkload node meta
controller-gen: script=$(controller-gen.script)
controller-gen: | $(controller-gen.bin) $(jx-cli.bin) $(kustomize.bin)

.PHONY: controller-gen
controller-gen:
	@: $(info generating controller descriptors)
	$(foreach package,$(packages),$(script))
	$(jx-cli.bin) gitops rename --dir=kustomizes

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

docker-%: tags := $(REPOSITORY)/$(IMAGE):latest $(REPOSITORY)/$(IMAGE):$(VERSION)

.PHONY: docker-tag
docker-tag: script=$(docker-tag.script)
docker-tag:
	$(foreach tag,$(tags),$(script))

define docker-tag.script =
docker tag $(IMAGE):latest $(tag)
$(newline)
endef

.PHONY: docker-push
docker-push: script=$(docker-push.script)
docker-push:
	$(foreach tag,$(tags),$(script))

define docker-push.script =
docker push $(tag)
$(newline)
endef

.PHONY: version
version:
	@echo $(VERSION)

VERSION:
	@echo v$(VERSION) > $(@)

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
