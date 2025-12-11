// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type dummyResourceValueChangeHandler struct {
	oldState *serverResourceModel
	newState *serverResourceModel
}

func TestStructureServer_isDiskEditParameterChanged(t *testing.T) {
	cases := []struct {
		msg    string
		in     *dummyResourceValueChangeHandler
		expect bool
	}{
		{
			msg: "nil",
			in: &dummyResourceValueChangeHandler{
				oldState: &serverResourceModel{},
				newState: &serverResourceModel{},
			},
			expect: false,
		},
		{
			msg: "added: disks",
			in: &dummyResourceValueChangeHandler{
				oldState: &serverResourceModel{},
				newState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
					},
				},
			},
			expect: true,
		},
		{
			msg: "added: disk_edit_parameter",
			in: &dummyResourceValueChangeHandler{
				oldState: &serverResourceModel{},
				newState: &serverResourceModel{
					DiskEdit: &serverDiskEditModel{},
				},
			},
			expect: true,
		},
		/* v2のmap[string]anyだから意味のあるテストなのでコメントアウト
		{
			msg: "added: network_interface",
			in: &dummyResourceValueChangeHandler{
				oldState: &serverResourceModel{},
				newState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						NetworkInterface: []serverNetworkInterfaceModel{},
					},
				},
			},
			expect: true,
		},
		*/
		{
			msg: "updated: no changes",
			in: &dummyResourceValueChangeHandler{
				oldState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream: types.StringValue("shared"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						Password: types.StringValue("password"),
						// 本来ではreq.Plan.Getにより型がついた値が入るためEqualがtrueになるが、直接生成ではEqualがfalseになるため、ここではSetNullで合わせている
						// falseになるのは、types.Set等のコレクションタイプはEqualの実装においてElementTypeがnilだと常にfalseを返すため
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
				newState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream: types.StringValue("shared"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						Password:  types.StringValue("password"),
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
			},
			expect: false,
		},
		{
			msg: "updated: disks",
			in: &dummyResourceValueChangeHandler{
				oldState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream: types.StringValue("shared"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						Password:  types.StringValue("password"),
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
				newState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"2"}),
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream: types.StringValue("shared"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						Password:  types.StringValue("password"),
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
			},
			expect: true,
		},
		{
			msg: "updated: disk_edit_parameter",
			in: &dummyResourceValueChangeHandler{
				oldState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream: types.StringValue("shared"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						Password:  types.StringValue("password"),
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
				newState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream: types.StringValue("shared"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						Password:  types.StringValue("password-upd"),
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
			},
			expect: true,
		},
		{
			msg: "updated: network_interface.upstream",
			in: &dummyResourceValueChangeHandler{
				oldState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream: types.StringValue("1"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						Password:  types.StringValue("password"),
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
				newState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream: types.StringValue("2"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						Password:  types.StringValue("password"),
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
			},
			expect: true,
		},
		{
			msg: "updated: network_interface.other",
			in: &dummyResourceValueChangeHandler{
				oldState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream:       types.StringValue("1"),
								UserIPAddress:  types.StringValue("192.168.0.1"),
								PacketFilterID: types.StringValue("1"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						Password:  types.StringValue("password"),
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
				newState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream:       types.StringValue("1"),
								UserIPAddress:  types.StringValue("192.168.0.2"),
								PacketFilterID: types.StringValue("2"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						Password:  types.StringValue("password"),
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
			},
			expect: false,
		},
		{
			msg: "deleted: disks",
			in: &dummyResourceValueChangeHandler{
				oldState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream: types.StringValue("shared"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
				newState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream: types.StringValue("shared"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
			},
			expect: true,
		},
		{
			msg: "deleted: network_interface",
			in: &dummyResourceValueChangeHandler{
				oldState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
						NetworkInterface: []serverNetworkInterfaceModel{
							{
								Upstream: types.StringValue("shared"),
							},
						},
					},
					DiskEdit: &serverDiskEditModel{
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
				newState: &serverResourceModel{
					serverBaseModel: serverBaseModel{
						Disks: common.StringsToTlist([]string{"1"}),
					},
					DiskEdit: &serverDiskEditModel{
						SSHKeys:   types.SetNull(types.StringType),
						SSHKeyIDs: types.SetNull(types.StringType),
					},
				},
			},
			expect: true,
		},
	}

	for _, tc := range cases {
		got := isDiskEditParameterChanged(tc.in.newState, tc.in.oldState)
		if got != tc.expect {
			fmt.Printf("diff: %s", cmp.Diff(tc.in.oldState, tc.in.newState, cmp.AllowUnexported(serverResourceModel{}, serverDiskEditModel{}, serverDiskEditScriptModel{}, serverNetworkInterfaceModel{})))
			t.Fatalf("got unexpected state: pattern: %s expected: %t actual: %t", tc.msg, tc.expect, got)
		}
	}
}
