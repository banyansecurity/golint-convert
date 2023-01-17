SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

#Constants
ARTIFACT_NAME = golint-convert

HOSTNAME=github.com
NAMESPACE=banyansecurity
NAME=banyan
OS_ARCH=darwin_amd64
VERSION=1.0.0

# ifeq ($(origin .RECIPEPREFIX), undefined)
#   $(error This Make does not support .RECIPEPREFIX. Please use GNU Make 4.0 or later)
# endif
# .RECIPEPREFIX = >

default: build

build:
	go build -o $(ARTIFACT_NAME)
.PHONY: build
