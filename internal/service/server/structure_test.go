// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/sacloud-sdk-go/api/iaas"
	iaastypes "github.com/sacloud/sacloud-sdk-go/api/iaas/types"
	serverBuilder "github.com/sacloud/sacloud-sdk-go/service/iaas/server/builder"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sctypes "github.com/sacloud/terraform-provider-sakura/internal/types"
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
			msg: "updated: disk_edit_parameter with write-only",
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
						PasswordWO:        types.StringValue("password-wo"),
						PasswordWOVersion: types.Int32Value(1),
						SSHKeys:           types.SetNull(types.StringType),
						SSHKeyIDs:         types.SetNull(types.StringType),
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
						PasswordWO:        types.StringValue("password-wo-upd"),
						PasswordWOVersion: types.Int32Value(2),
						SSHKeys:           types.SetNull(types.StringType),
						SSHKeyIDs:         types.SetNull(types.StringType),
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

func TestFlattenServerNICsSRN(t *testing.T) {
	cases := []struct {
		name     string
		zone     string
		switchID iaastypes.ID
		scope    iaastypes.EScope
		upstream string
	}{
		{name: "disconnected", zone: "is1c", upstream: "disconnect"},
		{name: "shared", zone: "tk1a", switchID: iaastypes.ID(2), scope: iaastypes.Scopes.Shared, upstream: "shared"},
		{name: "switch", zone: "is1b", switchID: iaastypes.ID(3), upstream: "3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nic := &iaas.InterfaceView{}
			nic.ID = iaastypes.ID(100000000001)
			nic.SwitchID = tc.switchID
			nic.SwitchScope = tc.scope
			server := &iaas.Server{Interfaces: []*iaas.InterfaceView{nic}}
			got := flattenServerNICs(server, tc.zone)
			if len(got) != 1 {
				t.Fatalf("flattenServerNICs() returned %d NICs, want 1", len(got))
			}
			wantSRN := sctypes.SRNValue(fmt.Sprintf("srnv1:sakura-%s:sakura.iaas.interface:100000000001", tc.zone))
			if !got[0].SRN.Equal(wantSRN) {
				t.Fatalf("SRN = %s, want %s", got[0].SRN, wantSRN)
			}
			if got[0].Upstream.ValueString() != tc.upstream {
				t.Fatalf("Upstream = %s, want %s", got[0].Upstream.ValueString(), tc.upstream)
			}
		})
	}
}

func TestPreserveNetworkingSuiteNICSRN(t *testing.T) {
	interfaceSRN := sctypes.SRNValue("srnv1:sakura-is1c:sakura.iaas.interface:100000000001")
	connectionSRN := sctypes.SRNValue("srnv1:sakura-is1c:sakura.networking-suite.interface-connection:100000000002")
	subnetSRN := sctypes.SRNValue("srnv1:sakura-is1c:sakura.networking-suite.subnet:100000000003")
	cases := []struct {
		name     string
		upstream string
		wantSRN  sctypes.SRN
	}{
		{name: "networking suite connection", upstream: networkingSuiteSubnetUpstream, wantSRN: connectionSRN},
		{name: "ordinary interface", upstream: "shared", wantSRN: interfaceSRN},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			target := []serverNetworkInterfaceModel{{
				SRN:      interfaceSRN,
				Upstream: types.StringValue("disconnect"),
			}}
			source := []serverNetworkInterfaceModel{{
				SRN:       connectionSRN,
				Upstream:  types.StringValue(tc.upstream),
				SubnetSRN: subnetSRN,
			}}
			got := mergeNetworkingSuiteNICState(target, source)
			if !got[0].SRN.Equal(tc.wantSRN) {
				t.Fatalf("SRN = %s, want %s", got[0].SRN, tc.wantSRN)
			}
			if tc.upstream == networkingSuiteSubnetUpstream && (!got[0].SubnetSRN.Equal(subnetSRN) || !got[0].Upstream.Equal(source[0].Upstream)) {
				t.Fatal("Networking Suite upstream and subnet SRN were not preserved")
			}
		})
	}
}

func TestValidateNetworkingSuiteNICs(t *testing.T) {
	subnetSRN := sctypes.SRNValue("srnv1:sakura-is1a:sakura.networking-suite.subnet:subnet-id")
	cases := []struct {
		name    string
		nic     serverNetworkInterfaceModel
		wantErr bool
	}{
		{
			name: "networking suite requires subnet srn",
			nic: serverNetworkInterfaceModel{
				Upstream: types.StringValue(networkingSuiteSubnetUpstream),
			},
			wantErr: true,
		},
		{
			name: "networking suite accepts subnet srn",
			nic: serverNetworkInterfaceModel{
				Upstream:  types.StringValue(networkingSuiteSubnetUpstream),
				SubnetSRN: subnetSRN,
			},
		},
		{
			name: "other upstream rejects subnet srn",
			nic: serverNetworkInterfaceModel{
				Upstream:  types.StringValue("shared"),
				SubnetSRN: subnetSRN,
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateNetworkingSuiteNICs([]serverNetworkInterfaceModel{tc.nic})
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateNetworkingSuiteNICs() error = %v, wantErr %t", err, tc.wantErr)
			}
		})
	}
}

func TestExpandNetworkingSuiteNICAsDisconnected(t *testing.T) {
	model := &serverResourceModel{
		serverBaseModel: serverBaseModel{
			NetworkInterface: []serverNetworkInterfaceModel{{
				Upstream: types.StringValue(networkingSuiteSubnetUpstream),
			}},
		},
	}
	if _, ok := expandServerNIC(model).(*serverBuilder.DisconnectedNICSetting); !ok {
		t.Fatal("expected a disconnected NIC setting for a Networking Suite interface")
	}
}

func TestNetworkingSuiteConfigChanged(t *testing.T) {
	subnetSRN := sctypes.SRNValue("srnv1:sakura-is1a:sakura.networking-suite.subnet:subnet-id")
	managedNIC := serverNetworkInterfaceModel{
		Upstream:  types.StringValue(networkingSuiteSubnetUpstream),
		SubnetSRN: subnetSRN,
	}
	tests := []struct {
		name string
		plan []serverNetworkInterfaceModel
		old  []serverNetworkInterfaceModel
		want bool
	}{
		{
			name: "unchanged",
			plan: []serverNetworkInterfaceModel{managedNIC},
			old:  []serverNetworkInterfaceModel{managedNIC},
		},
		{
			name: "other nic upstream changed",
			plan: []serverNetworkInterfaceModel{
				managedNIC,
				{Upstream: types.StringValue("100000000000")},
			},
			old: []serverNetworkInterfaceModel{
				managedNIC,
				{Upstream: types.StringValue("shared")},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := networkingSuiteConfigChanged(tc.plan, tc.old); got != tc.want {
				t.Fatalf("networkingSuiteConfigChanged() = %t, want %t", got, tc.want)
			}
		})
	}
}
