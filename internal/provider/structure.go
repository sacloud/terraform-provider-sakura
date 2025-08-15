// Copyright 2016-2025 terraform-provider-sakuracloud authors
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

package sakura

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/mitchellh/go-homedir"
	iaastypes "github.com/sacloud/iaas-api-go/types"
)

func sakuraCloudID(id string) iaastypes.ID {
	return iaastypes.StringID(id)
}

func expandSakuraCloudID(d basetypes.StringValue) iaastypes.ID {
	if d.IsNull() || d.IsUnknown() {
		return iaastypes.ID(0)
	}

	return sakuraCloudID(d.ValueString())
}

// SDK v2のHasChangeの代替
func hasChange(x, y any) bool {
	return !cmp.Equal(x, y)
}

func getZone(zone basetypes.StringValue, client *APIClient, diags *diag.Diagnostics) string {
	if zone.IsNull() || zone.IsUnknown() {
		return client.defaultZone
	}

	z := zone.ValueString()
	if err := StringInSlice(client.zones, "zone", z, false); err != nil {
		diags.AddError("Get zone error", err.Error())
		return ""
	}

	return z
}

func getApiClientFromProvider(providerData any, diags *diag.Diagnostics) *APIClient {
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

func tlistToStrings(d types.List) []string {
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

func tsetToStrings(d types.Set) []string {
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

func stringsToTset(tags []string) types.Set {
	// types.SetValueでは内部でcontext.Background()を呼び出しているため、同じアプローチを採用
	setValue, _ := types.SetValueFrom(context.Background(), types.StringType, tags)
	return setValue
}

func intToInt32(i int) int32 {
	return int32(i)
}

func intToInt64(i int) int64 {
	return int64(i)
}

func mapTo[S any, T any](s []S, cast func(S) T) []T {
	if len(s) == 0 {
		return nil
	}

	t := make([]T, 0, len(s))
	for _, v := range s {
		t = append(t, cast(v))
	}
	return t
}

func StringInSlice(validList []string, k string, v string, ignoreCase bool) error {
	for _, valid := range validList {
		if v == valid {
			return nil
		}
		if ignoreCase && strings.EqualFold(v, valid) {
			return nil
		}
	}

	return fmt.Errorf("invalid %s value: %s. valid values are %s", k, v, validList)
}

func expandHomeDir(path string) (string, error) {
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

func md5CheckSumFromFile(path string) (string, error) {
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
