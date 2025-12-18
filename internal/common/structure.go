// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/mitchellh/go-homedir"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/plans"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

func SakuraCloudID(id string) iaastypes.ID {
	return iaastypes.StringID(id)
}

func ExpandSakuraCloudID(d basetypes.StringValue) iaastypes.ID {
	if d.IsNull() || d.IsUnknown() {
		return iaastypes.ID(0)
	}

	return SakuraCloudID(d.ValueString())
}

func ExpandSakuraCloudIDs(d basetypes.SetValue) []iaastypes.ID {
	strIDs := TsetToStrings(d)
	if len(strIDs) == 0 {
		return nil
	}

	var ids []iaastypes.ID
	for _, strID := range strIDs {
		if strID == "" {
			continue
		}
		ids = append(ids, SakuraCloudID(strID))
	}

	return ids
}

func ExpandSakuraCloudIDsFromList(d basetypes.ListValue) []iaastypes.ID {
	strIDs := TlistToStrings(d)
	if len(strIDs) == 0 {
		return nil
	}

	var ids []iaastypes.ID
	for _, strID := range strIDs {
		if strID == "" {
			continue
		}
		ids = append(ids, SakuraCloudID(strID))
	}

	return ids
}

func GetZone(zone basetypes.StringValue, client *APIClient, diags *diag.Diagnostics) string {
	if zone.IsNull() || zone.IsUnknown() {
		return client.defaultZone
	}

	z := zone.ValueString()
	if err := utils.StringInSlice(client.zones, "zone", z, false); err != nil {
		diags.AddError("Get zone error", err.Error())
		return ""
	}

	return z
}

func GetApiClientFromProvider(providerData any, diags *diag.Diagnostics) *APIClient {
	if providerData == nil {
		return nil
	}

	apiclient, ok := providerData.(*APIClient)
	if !ok {
		diags.AddError("Unexpected ProviderData type", "Expected *APIClient.")
		return nil
	}

	return apiclient
}

func TlistToStrings(d types.List) []string {
	if d.IsNull() || d.IsUnknown() {
		return nil
	}

	var tags []string
	for _, v := range d.Elements() {
		if vStr, ok := v.(types.String); ok && !vStr.IsNull() && !vStr.IsUnknown() {
			tags = append(tags, vStr.ValueString())
		}
	}
	return tags
}

func TsetToStrings(d types.Set) []string {
	if d.IsNull() || d.IsUnknown() {
		return nil
	}

	var tags []string
	for _, v := range d.Elements() {
		if vStr, ok := v.(types.String); ok && !vStr.IsNull() && !vStr.IsUnknown() {
			tags = append(tags, vStr.ValueString())
		}
	}
	return tags
}

func TlistToStringsOrDefault(d types.List) []string {
	list := TlistToStrings(d)
	if list == nil {
		return []string{}
	}
	return list
}

func TsetToStringsOrDefault(d types.Set) []string {
	set := TsetToStrings(d)
	if set == nil {
		return []string{}
	}
	return set
}

func StringsToTset(values []string) types.Set {
	// types.SetValueでは内部でcontext.Background()を呼び出しているため、同じアプローチを採用
	setValue, _ := types.SetValueFrom(context.Background(), types.StringType, values)
	return setValue
}

func StringsToTlist(values []string) types.List {
	// types.SetValueでは内部でcontext.Background()を呼び出しているため、同じアプローチを採用
	listValue, _ := types.ListValueFrom(context.Background(), types.StringType, values)
	return listValue
}

func StrMapToTmap(values map[string]string) types.Map {
	mapValue, _ := types.MapValueFrom(context.Background(), types.StringType, values)
	return mapValue
}

func TmapToStrMap(values types.Map) map[string]string {
	if values.IsNull() || values.IsUnknown() {
		return nil
	}

	result := make(map[string]string)
	for k, v := range values.Elements() {
		if vStr, ok := v.(types.String); ok && !vStr.IsNull() && !vStr.IsUnknown() {
			result[k] = vStr.ValueString()
		}
	}
	return result
}

func IntToInt32(i int) int32 {
	return int32(i)
}

func IntToInt64(i int) int64 {
	return int64(i)
}

func ToString[S ~string](s S) string {
	return string(s)
}

func MapTo[S any, T any](s []S, cast func(S) T) []T {
	if len(s) == 0 {
		return nil
	}

	t := make([]T, 0, len(s))
	for _, v := range s {
		t = append(t, cast(v))
	}
	return t
}

func MustAtoI(target string) int {
	v, _ := strconv.Atoi(target)
	return v
}

func MustAtoInt64(target string) int64 {
	v, _ := strconv.ParseInt(target, 10, 64)
	return v
}

func ExpandHomeDir(path string) (string, error) {
	expanded, err := homedir.Expand(path)
	if err != nil {
		return "", fmt.Errorf("expanding homedir in path[%s] is failed: %s", expanded, err)
	}
	// file exists?
	if _, err := os.Stat(expanded); err != nil {
		return "", fmt.Errorf("opening file[%s] is failed: %s", expanded, err)
	}

	return expanded, nil
}

func Md5CheckSumFromFile(path string) (string, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("opening file[%s] is failed: %s", path, err)
	}
	defer f.Close() //nolint

	b := base64.NewEncoder(base64.StdEncoding, f)
	defer b.Close() //nolint

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, f); err != nil {
		return "", fmt.Errorf("encoding to base64 from file[%s] is failed: %s", path, err)
	}

	h := md5.New() //nolint:gosec
	if _, err := io.Copy(h, &buf); err != nil {
		return "", fmt.Errorf("calculating md5 from file[%s] is failed: %s", path, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func ExpandBackupWeekdays(d types.Set) []iaastypes.EDayOfTheWeek {
	var vs []iaastypes.EDayOfTheWeek

	for _, w := range TsetToStrings(d) {
		vs = append(vs, iaastypes.EDayOfTheWeek(w))
	}
	iaastypes.SortDayOfTheWeekList(vs)
	return vs
}

func FlattenBackupWeekdays(weekdays []iaastypes.EDayOfTheWeek) types.Set {
	set := make([]string, 0, len(weekdays))
	for _, w := range weekdays {
		set = append(set, w.String())
	}
	return StringsToTset(set)
}

func FlattenTags(tags iaastypes.Tags) types.Set {
	filtered := iaastypes.Tags{}
	for _, t := range tags {
		if !strings.HasPrefix(t, plans.PreviousIDTagName) && !strings.HasPrefix(t, "@appliance-") {
			filtered = append(filtered, t)
		}
	}
	return StringsToTset(filtered)
}

// TODO: Interface/Genericsを使ってXXXMonitoringSuite/XXXMonitoringSuiteLogを一つにすることを検討する。
// IaaS API側で統合される可能性もあるので、現状はとりあえず構造体ごとに処理を分けておく。
func FlattenMonitoringSuiteLog(ms *iaas.MonitoringSuiteLog) types.Object {
	enabled := false
	if ms != nil {
		enabled = ms.Enabled
	}
	model := SakuraMonitoringSuiteModel{
		Enabled: types.BoolValue(enabled),
	}
	value, diags := types.ObjectValueFrom(context.Background(), model.AttributeTypes(), model)
	if diags.HasError() {
		return types.ObjectNull(SakuraMonitoringSuiteModel{}.AttributeTypes())
	}
	return value
}

func ExpandMonitoringSuiteLog(ms types.Object) *iaas.MonitoringSuiteLog {
	enabled := false
	if !ms.IsNull() && !ms.IsUnknown() {
		var model SakuraMonitoringSuiteModel
		diags := ms.As(context.Background(), &model, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil
		}
		enabled = model.Enabled.ValueBool()
	}
	return &iaas.MonitoringSuiteLog{Enabled: enabled}
}

func FlattenMonitoringSuite(ms *iaas.MonitoringSuite) types.Object {
	enabled := false
	if ms != nil {
		enabled = ms.Enabled
	}
	model := SakuraMonitoringSuiteModel{
		Enabled: types.BoolValue(enabled),
	}
	value, diags := types.ObjectValueFrom(context.Background(), model.AttributeTypes(), model)
	if diags.HasError() {
		return types.ObjectNull(SakuraMonitoringSuiteModel{}.AttributeTypes())
	}
	return value
}

func ExpandMonitoringSuite(ms types.Object) *iaas.MonitoringSuite {
	enabled := false
	if !ms.IsNull() && !ms.IsUnknown() {
		var model SakuraMonitoringSuiteModel
		diags := ms.As(context.Background(), &model, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil
		}
		enabled = model.Enabled.ValueBool()
	}
	return &iaas.MonitoringSuite{Enabled: enabled}
}

// Use saclient-go's version in the future
func Ptr[T any](t T) *T { return &t }
