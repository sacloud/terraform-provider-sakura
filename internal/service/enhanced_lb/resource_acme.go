// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package enhanced_lb

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type enhancedLBACMEResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &enhancedLBACMEResource{}
	_ resource.ResourceWithConfigure   = &enhancedLBACMEResource{}
	_ resource.ResourceWithImportState = &enhancedLBACMEResource{}
)

func NewEnhancedLBACMEResource() resource.Resource {
	return &enhancedLBACMEResource{}
}

func (r *enhancedLBACMEResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_enhanced_lb_acme"
}

func (r *enhancedLBACMEResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type enhancedLBACMEResourceModel struct {
	ID                        types.String   `tfsdk:"id"`
	EnhancedLBID              types.String   `tfsdk:"enhanced_lb_id"`
	AcceptTOS                 types.Bool     `tfsdk:"accept_tos"`
	CommonName                types.String   `tfsdk:"common_name"`
	SubjectAltNames           types.Set      `tfsdk:"subject_alt_names"`
	UpdateDelaySec            types.Int32    `tfsdk:"update_delay_sec"`
	GetCertificatesTimeoutSec types.Int32    `tfsdk:"get_certificates_timeout_sec"`
	Certificate               types.Object   `tfsdk:"certificate"`
	Timeouts                  timeouts.Value `tfsdk:"timeouts"`
}

func (r *enhancedLBACMEResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("Enhanced LB ACME"),
			"enhanced_lb_id": schema.StringAttribute{
				Required:    true,
				Description: "The id of the Enhanced LB that set ACME settings to",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"accept_tos": schema.BoolAttribute{
				Required:    true,
				Description: "The flag to accept the current Let's Encrypt terms of service(see: https://letsencrypt.org/repository/). This must be set `true` explicitly",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"common_name": schema.StringAttribute{
				Required:    true,
				Description: "The FQDN used by ACME. This must set resolvable value",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subject_alt_names": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "The Subject alternative names used by ACME",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"update_delay_sec": schema.Int32Attribute{
				Optional:    true,
				Description: "The wait time in seconds. This typically used for waiting for a DNS propagation",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"get_certificates_timeout_sec": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The timeout in seconds for the certificate acquisition to complete",
				Default:     int32default.StaticInt32(120),
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"certificate": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The certificate information",
				Attributes: map[string]schema.Attribute{
					"server_cert": schema.StringAttribute{
						Computed:    true,
						Description: "The certificate for a server",
					},
					"intermediate_cert": schema.StringAttribute{
						Computed:    true,
						Description: "The intermediate certificate for a server",
					},
					"private_key": schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "The private key for a server",
					},
					"common_name": schema.StringAttribute{
						Computed:    true,
						Description: "The common name of the certificate",
					},
					"subject_alt_names": schema.StringAttribute{
						Computed:    true,
						Description: "The subject alternative names of the certificate",
					},
					"additional_certificate": schema.ListNestedAttribute{
						Computed:    true,
						Description: "The additional certificates",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"server_cert": schema.StringAttribute{
									Computed:    true,
									Description: "The certificate for a server",
								},
								"intermediate_cert": schema.StringAttribute{
									Computed:    true,
									Description: "The intermediate certificate for a server",
								},
								"private_key": schema.StringAttribute{
									Computed:    true,
									Sensitive:   true,
									Description: "The private key for a server",
								},
							},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Enhanced Load Balancer's ACME",
	}
}

func (r *enhancedLBACMEResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *enhancedLBACMEResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan enhancedLBACMEResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout20min)
	defer cancel()

	elbOp := iaas.NewProxyLBOp(r.client)
	elbID := plan.EnhancedLBID.ValueString()

	common.SakuraMutexKV.Lock(elbID)
	defer common.SakuraMutexKV.Unlock(elbID)

	elb := getELB(ctx, r.client, elbID, &resp.State, &resp.Diagnostics)
	if elb == nil {
		return
	}

	le := &iaas.ProxyLBACMESetting{Enabled: false}
	tos := plan.AcceptTOS.ValueBool()
	commonName := plan.CommonName.ValueString()
	altNames := common.TsetToStrings(plan.SubjectAltNames)
	if tos {
		le = &iaas.ProxyLBACMESetting{
			Enabled:         true,
			CommonName:      commonName,
			SubjectAltNames: altNames,
		}
	}

	if plan.UpdateDelaySec.ValueInt32() > 0 {
		time.Sleep(time.Duration(plan.UpdateDelaySec.ValueInt32()) * time.Second)
	}

	elb, err := elbOp.UpdateSettings(ctx, elb.ID, &iaas.ProxyLBUpdateSettingsRequest{
		HealthCheck:          elb.HealthCheck,
		SorryServer:          elb.SorryServer,
		BindPorts:            elb.BindPorts,
		Servers:              elb.Servers,
		Rules:                elb.Rules,
		LetsEncrypt:          le,
		StickySession:        elb.StickySession,
		Timeout:              elb.Timeout,
		Gzip:                 elb.Gzip,
		BackendHttpKeepAlive: elb.BackendHttpKeepAlive,
		ProxyProtocol:        elb.ProxyProtocol,
		Syslog:               elb.Syslog,
		SettingsHash:         elb.SettingsHash,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("setting Enchanced LB[%s] ACME is failed: %s", elbID, err))
		return
	}
	if err := elbOp.RenewLetsEncryptCert(ctx, elb.ID); err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("renewing ACME Certificates at Enchanced LB[%s] is failed: %s", elb.ID, err))
		return
	}

	if err := waitForProxyLBCertAcquisitionFW(ctx, r.client, elb.ID.String(), plan.GetCertificatesTimeoutSec.ValueInt32()); err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("waiting for ACME certificate acquisition of Enchanced LB[%s] is failed: %s", elb.ID, err))
		return
	}

	if err := plan.updateState(ctx, r.client, elb); err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to update state for Enhanced LB[%s] ACME: %s", elb.ID, err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *enhancedLBACMEResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state enhancedLBACMEResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	elb := getELB(ctx, r.client, state.EnhancedLBID.ValueString(), &resp.State, &resp.Diagnostics)
	if elb == nil {
		return
	}

	if err := state.updateState(ctx, r.client, elb); err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to update state for Enhanced LB[%s] ACME: %s", elb.ID, err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *enhancedLBACMEResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is not supported
}

func (r *enhancedLBACMEResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state enhancedLBACMEResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	elbOp := iaas.NewProxyLBOp(r.client)
	elbID := state.EnhancedLBID.ValueString()

	common.SakuraMutexKV.Lock(elbID)
	defer common.SakuraMutexKV.Unlock(elbID)

	elb := getELB(ctx, r.client, elbID, &resp.State, &resp.Diagnostics)
	if elb == nil {
		return
	}

	_, err := elbOp.UpdateSettings(ctx, elb.ID, &iaas.ProxyLBUpdateSettingsRequest{
		HealthCheck:          elb.HealthCheck,
		SorryServer:          elb.SorryServer,
		BindPorts:            elb.BindPorts,
		Servers:              elb.Servers,
		Rules:                elb.Rules,
		LetsEncrypt:          &iaas.ProxyLBACMESetting{Enabled: false},
		StickySession:        elb.StickySession,
		Timeout:              elb.Timeout,
		Gzip:                 elb.Gzip,
		BackendHttpKeepAlive: elb.BackendHttpKeepAlive,
		ProxyProtocol:        elb.ProxyProtocol,
		Syslog:               elb.Syslog,
		SettingsHash:         elb.SettingsHash,
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("clearing ACME Setting of Enhanced LB[%s] is failed: %s", elb.ID, err))
		return
	}
}

func (model *enhancedLBACMEResourceModel) updateState(ctx context.Context, client *common.APIClient, data *iaas.ProxyLB) error {
	elbOp := iaas.NewProxyLBOp(client)
	certs, err := elbOp.GetCertificates(ctx, data.ID)
	if err != nil {
		// even if certificate is deleted, it will not result in an error
		return err
	}

	model.ID = types.StringValue(data.ID.String())
	model.Certificate = flattenEnhancedLBCerts(certs)

	return nil
}

func waitForProxyLBCertAcquisitionFW(ctx context.Context, client *common.APIClient, elbID string, timeoutSec int32) error {
	elbOp := iaas.NewProxyLBOp(client)

	waitCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("waiting for certificate acquisition failed: %s", waitCtx.Err())
		default:
			cert, err := elbOp.GetCertificates(ctx, common.SakuraCloudID(elbID))
			if err != nil {
				// even if certificate is deleted, it will not result in an error
				return err
			}
			if cert.PrimaryCert != nil && cert.PrimaryCert.ServerCertificate != "" {
				return nil
			}
			time.Sleep(5 * time.Second)
		}
	}
}
