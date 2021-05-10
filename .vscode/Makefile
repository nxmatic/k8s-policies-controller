include ../make.d/make.mk
include ../make.d/os.mk


kustomizes.files := $(shell find kustomizes -name '*.yaml')
manifest.files := manifest §$(wildcard manifest/*.yaml)

$(manifest.files)&: $(kustomizes.files) | $(jx.bin) $(kustomize.bin)
	@: $(building devpod manifests)
	kustomize build kustomizes -o manifest
	jx cli gitops rename manifest

apply: manifest
	kpt live apply manifest $(if $(force-conflicts),--force-conflicts) $(if $(adopt),--inventory-policy=adopt)
