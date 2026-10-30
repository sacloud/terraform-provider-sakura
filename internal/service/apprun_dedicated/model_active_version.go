// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ver "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/version"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type activeVerResourceModel struct{ verResourceModel }

// the ID is the application ID, not the synthetic "<application>/<version>" of a version
func (v *activeVerResourceModel) updateState(ctx context.Context, d *ver.VersionDetail, aid appID) diag.Diagnostics {
	before := v.verModel
	ret := v.verResourceModel.updateState(ctx, d, aid)
	v.ID = uuid2StringValue(aid)
	v.keepEmpty(&before)

	return ret
}

// updateState stores an empty cmd, exposed_ports or host as null, which Terraform would
// then diff against the empty value of the plan or the state; keep that value
func (v *activeVerResourceModel) keepEmpty(before *verModel) {
	if before.ExposedPorts != nil && len(before.ExposedPorts) == 0 && v.ExposedPorts == nil {
		v.ExposedPorts = []exposedPortModel{}
	}

	emptyHosts := make(map[int32]types.Set, len(before.ExposedPorts))

	for _, p := range before.ExposedPorts {
		if utils.IsKnown(p.Host) && len(p.Host.Elements()) == 0 {
			emptyHosts[p.TargetPort.ValueInt32()] = p.Host
		}
	}

	for i, p := range v.ExposedPorts {
		if host, ok := emptyHosts[p.TargetPort.ValueInt32()]; ok && p.Host.IsNull() {
			v.ExposedPorts[i].Host = host
		}
	}

	if utils.IsKnown(before.Cmd) && len(before.Cmd.Elements()) == 0 && v.Cmd.IsNull() {
		v.Cmd = before.Cmd
	}
}

// drops the variables the active version no longer has, or has with the other secrecy
// (updateState only merges by key); a deprecated secret = true entry stays in env_vars
func (v *activeVerResourceModel) dropMissingVars(d *ver.VersionDetail) {
	secret := make(map[string]bool, len(d.EnvVars))

	for _, e := range d.EnvVars {
		secret[e.Key] = e.Secret
	}

	v.EnvVars = slices.DeleteFunc(v.EnvVars, func(e envVarModel) bool {
		s, ok := secret[e.Key.ValueString()]
		return !ok || (s && !e.Secret.ValueBool())
	})

	v.SecretVars = slices.DeleteFunc(v.SecretVars, func(s secretVarModel) bool {
		isSecret, ok := secret[s.Key.ValueString()]
		return !ok || !isSecret
	})
}
