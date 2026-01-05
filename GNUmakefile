# Copyright 2016-2025 terraform-provider-sakura authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#====================
AUTHOR          ?= The sacloud/terraform-provider-sakura Authors
COPYRIGHT_YEAR  ?= 2016-2025

BIN            ?= terraform-provider-sakura
BUILD_LDFLAGS   ?= "-s -w -X github.com/sacloud/terraform-provider-sakura/sakura.Revision=`git rev-parse --short HEAD`"

include includes/go/common.mk
include includes/go/single.mk
#====================
export GOPROXY=https://proxy.golang.org
TF_DOCS_EXAMPLE_FILES ?= $(shell find ./examples -name '*.tf' -o -name '*.sh')

default: generate-docs fmt set-license go-licenses-check goimports lint test build

UNIT_TEST_UA ?= (Unit Test)
ACC_TEST_UA ?= (Acceptance Test)

.PHONY: tools
tools: dev-tools

.PHONY: generate-docs
generate-docs: docs/index.md
docs/index.md: $(GO_FILES) $(TF_DOCS_EXAMPLE_FILES)
	cd tools; go generate ./... ; ruby ./update_subcategories.rb
