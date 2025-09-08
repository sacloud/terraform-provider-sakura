// Copyright 2016-2025 terraform-provider-sakura authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"errors"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"

	"github.com/sacloud/iaas-api-go/search"
	"github.com/sacloud/iaas-api-go/search/keys"
)

const (
	filterAttrName                   = "filter"
	filteringOperatorPartialMatchAnd = "partial_match_and"
	filteringOperatorExactMatchOr    = "exact_match_or"
)

type FilterSchemaOption struct {
	excludeTags bool
}

var (
	filterConfigKeys = []string{
		"id",
		"names",
		"condition",
	}
	filterConfigKeysWithTags = append(filterConfigKeys, "tags")
	filteringOperators       = []string{
		filteringOperatorPartialMatchAnd,
		filteringOperatorExactMatchOr,
	}
)

type FilterConditionBlockModel struct {
	Name     types.String `tfsdk:"name"`
	Values   types.List   `tfsdk:"values"`
	Operator types.String `tfsdk:"operator"`
}

type FilterBlockModel struct {
	ID        types.String                `tfsdk:"id"`
	Names     types.List                  `tfsdk:"names"`
	Tags      types.Set                   `tfsdk:"tags"`
	Condition []FilterConditionBlockModel `tfsdk:"condition"`
}

func FilterSchema(opt *FilterSchemaOption) map[string]schema.Block {
	if opt == nil {
		opt = &FilterSchemaOption{}
	}
	keys := filterConfigKeysWithTags
	if opt.excludeTags {
		keys = filterConfigKeys
	}
	pathExpressions := make([]path.Expression, 0, len(keys))
	for _, key := range keys {
		pathExpressions = append(pathExpressions, path.MatchRoot("filter").AtName(key))
	}

	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Optional:    true,
			Description: "The resource id on SakuraCloud used for filtering",
			Validators: []validator.String{
				stringvalidator.ExactlyOneOf(pathExpressions...),
			},
		},
		"names": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "The resource names on SakuraCloud used for filtering. If multiple values are specified, they combined as AND condition",
			Validators: []validator.List{
				listvalidator.ExactlyOneOf(pathExpressions...),
			},
		},
		"tags": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "The resource tags on SakuraCloud used for filtering. If multiple values are specified, they combined as AND condition",
			Validators: []validator.Set{
				setvalidator.ExactlyOneOf(pathExpressions...),
			},
		},
	}
	if opt.excludeTags {
		delete(attrs, "tags")
	}

	return map[string]schema.Block{
		"filter": schema.SingleNestedBlock{
			Description: "One or more values used for filtering, as defined below",
			Attributes:  attrs,
			Blocks: map[string]schema.Block{
				"condition": schema.ListNestedBlock{
					Description: "One or more name/values pairs used for filtering. There are several valid keys, for a full reference, check out finding section in the [SakuraCloud API reference](https://developer.sakura.ad.jp/cloud/api/1.1/)",
					Validators: []validator.List{
						listvalidator.ExactlyOneOf(pathExpressions...),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								//Required:    true,　　// Blockの中に一つでもRequiredがあると、Block自体がRequiredになってしまうため、ここではOptionalにする
								Optional:    true,
								Description: "The name of the target field. This value is case-sensitive",
							},
							"values": schema.ListAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "The values of the condition. If multiple values are specified, they combined as AND condition",
							},
							"operator": schema.StringAttribute{
								Optional: true,
								Description: desc.Sprintf(
									"The filtering operator. This must be one of following:  \n%s", filteringOperators,
								),
								Validators: []validator.String{
									stringvalidator.OneOfCaseInsensitive(filteringOperators...),
								},
							},
						},
					},
				},
			},
		},
	}
}

var ErrFilterNoResult = errors.New("Your query returned no results. Please change your filter or selectors and try again")

func FilterNoResultErr(diag *diag.Diagnostics) {
	if os.Getenv(resource.EnvTfAcc) != "" {
		return
	}

	diag.AddError("Filter No Result", ErrFilterNoResult.Error())
}

func CreateFindCondition(id types.String, name types.String, tags types.Set) *iaas.FindCondition {
	condition := &iaas.FindCondition{}

	var names types.List
	if !name.IsNull() && !name.IsUnknown() && name.ValueString() != "" {
		elements := []attr.Value{name}
		names, _ = types.ListValue(types.StringType, elements)
	}
	condition.Filter = ExpandSearchFilter(&FilterBlockModel{ID: id, Names: names, Tags: tags})

	return condition
}

func ExpandSearchFilter(filters *FilterBlockModel) search.Filter {
	ret := search.Filter{}
	// ID
	if !filters.ID.IsNull() && !filters.ID.IsUnknown() {
		id := filters.ID.ValueString()
		if id != "" {
			ret[search.Key(keys.ID)] = search.AndEqual(id)
		}
	}
	// Names
	if !filters.Names.IsNull() && !filters.Names.IsUnknown() {
		names := TlistToStrings(filters.Names)
		if len(names) > 0 {
			ret[search.Key(keys.Name)] = search.AndEqual(names...)
		}
	}
	// Tags
	if !filters.Tags.IsNull() && !filters.Tags.IsUnknown() {
		tags := TsetToStrings(filters.Tags)
		if len(tags) > 0 {
			ret[search.Key(keys.Tags)] = search.TagsAndEqual(tags...)
		}
	}

	// others
	if len(filters.Condition) != 0 {
		for _, cond := range filters.Condition {
			keyName := cond.Name.ValueString()
			values := TlistToStrings(cond.Values)
			operator := cond.Operator.ValueString()
			if operator == "" {
				// operatorのスキーマ定義でDefaultを設定できないため、ここでデフォルト値を設定
				operator = filteringOperatorPartialMatchAnd
			}

			var conditions []string
			for _, value := range values {
				if value != "" {
					conditions = append(conditions, value)
				}
			}
			if len(conditions) > 0 {
				if operator == filteringOperatorExactMatchOr {
					var vs []interface{}
					for _, p := range conditions {
						vs = append(vs, p)
					}
					ret[search.Key(keyName)] = search.OrEqual(vs...)
				} else {
					ret[search.Key(keyName)] = search.AndEqual(conditions...)
				}
			}
		}
	}

	return ret
}

type nameFilterable interface {
	GetName() string
}

func hasNames(target interface{}, cond []string) bool {
	t, ok := target.(nameFilterable)
	if !ok {
		return false
	}
	name := t.GetName()
	for _, c := range cond {
		if !strings.Contains(name, c) {
			return false
		}
	}
	return true
}

type tagFilterable interface {
	HasTag(string) bool
}

func hasTags(target interface{}, cond []string) bool {
	t, ok := target.(tagFilterable)
	if !ok {
		return false
	}
	for _, c := range cond {
		if !t.HasTag(c) {
			return false
		}
	}
	return true
}
