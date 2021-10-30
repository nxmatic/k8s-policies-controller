include $(top-dir)/make.d/os.mk
include $(top-dir)/make.d/docker.mk
include $(top-dir)/make.d/docker-compose.mk

include $(cache-dir)/devcontainer.mk

define devcontainer.cache.mk =
.devcontainer.names := $(shell $(jq.bin) -M -r '.dockerComposeFile[]' .devcontainer/devcontainer.json)
endef

.devcontainer.files := $(addprefix .devcontainer/,$(.devcontainer.names))

.devcontainer/devcontainer.yaml: .devcontainer/devcontainer.json

.devcontainer/docker-compose~%: .devcontainer/devcontainer.json

.devcontainer/docker-compose~%: project-directory s= $(top-dir)/.devcontainer
.devcontainer/docker-compose~%: project-name := $(top-dir.name)_devcontainer
.devcontainer/docker-compose~%: files := $(.devcontainer.files)

.devcontainer/docker-compose~%: export IMAGE_TYPE ?= full
.devcontainer/docker-compose~%: export PREVIEW_PACKAGES_VERSION ?= $(version)

define .devcontainer.docker-compose.rule =
.devcontainer/docker-compose~$(command): docker-compose~$(command)
.devcontainer/docker-compose~$(command): docker-compose~$(command)*options=$$(.devcontainer/docker-compose~$(command)*options)
endef

$(foreach command,$(docker-compose.commands),$(eval $(call .devcontainer.docker-compose.rule)))

$(call make.to-options.with-pattern,.devcontainer/docker-compose~%)

#.devcontainer/devcontainer.json:
#	@:$(info converting devcontainer back to json)
#	yq --output-format=json eval . $(<) > $(@)
#	touch -r $(@) $(<)
#
#.devcontainer/devcontainer.json: .devcontainer/devcontainer.yaml

.devcontainer/devcontainer.yaml:
	@:$(info converting devcontainer to yaml)
	yq -P eval $(@:.yaml=.json) > $(@)
	touch -r $(@) $(<)
