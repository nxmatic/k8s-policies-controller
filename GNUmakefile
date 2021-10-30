.ONESHELL:

SHELLFLAGS=-e -cs -o pipelinefail -x
# MAKEFLAGS += --no-builtin-rules
# MAKEFLAGS += --no-builtin-variables
MAKEFLAGS += --no-print-directory

.is-verbose := $(if $(filter-out disabled,$(make.trace)),$(true),$(false))
.if-verbose = $(if $(.is-verbose),$(1),$(2))

@ = $(call .if-verbose,,@)

# ensure we have the cache generated and the required cli

goals := $(MAKECMDGOALS)
goal= := $(filter-out .local,$(goals))
goals := $(filter-out make.d/make.mk,$(goals))
goals := $(filter-out $(git-credentials),$(goals))


.PHONY: $(goals)

$(goals): %:
	@: $(info executing $(@))
	exec $(MAKE) -f Makefile $(*)

$(goals): | .local

.local:
	mkdir -p .local/tmp .local/etc .local/bin .local/var/cache
	$(MAKE) -f make.d/make.mk .local/bin/jq

.local: | make.d/make.mk

make.d/make.mk:
	git submodule init
	git submodule update

.cluster.is-inside := $(if $(wildcard /var/run/secrets/kubernetes.io),true,false)

ifeq (true,$(.cluster.is-inside))

git-credentials := $(if $(XDG_CONFIG_HOME),$(XDG_CONFIG_HOME),$(HOME))/git/credentials

make.d/make.mk:  | $(git-credentials)

$(git-credentials): | $(dir $(git-credentials))
	: $(file >$(@),$(git-credentials.template))
	git config --global user.username $(username)
	git config --global credential.helper store

$(git-credentials): secret   = /var/run/secrets/git-credentials
$(git-credentials): username = $(file <$(secret)/username)
$(git-credentials): password = $(file <$(secret)/password)

$(dir $(git-credentials)):
	@:
	mkdir -p $(@)

define git-credentials.template =
https://$(username):$(password)@github.com
$()
endef

endif
