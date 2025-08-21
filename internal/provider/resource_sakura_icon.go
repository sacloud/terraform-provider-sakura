package sakura

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mitchellh/go-homedir"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
)

type iconResource struct {
	client *APIClient
}

var (
	_ resource.Resource                = &iconResource{}
	_ resource.ResourceWithConfigure   = &iconResource{}
	_ resource.ResourceWithImportState = &iconResource{}
)

func NewIconResource() resource.Resource {
	return &iconResource{}
}

func (r *iconResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_icon"
}

func (r *iconResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

// TODO: model.goに切り出してdata sourceと共通化する
type iconResourceModel struct {
	ID            types.String   `tfsdk:"id"`
	Name          types.String   `tfsdk:"name"`
	Source        types.String   `tfsdk:"source"`
	Base64Content types.String   `tfsdk:"base64content"`
	Tags          types.Set      `tfsdk:"tags"`
	URL           types.String   `tfsdk:"url"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

func (r *iconResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   schemaResourceId("Icon"),
			"name": schemaResourceName("Icon"),
			"tags": schemaResourceTags("Icon"),
			"source": schema.StringAttribute{
				Optional:    true,
				Description: "The file path to upload as the Icon.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("base64content")),
				},
			},
			"base64content": schema.StringAttribute{
				Optional:    true,
				Description: "The base64 encoded content to upload as the Icon.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("source")),
				},
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL for getting the icon's raw data.",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *iconResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *iconResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan iconResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutCreate(ctx, plan.Timeouts, timeout5min)
	defer cancel()

	iconOp := iaas.NewIconOp(r.client)
	createReq, err := expandIconCreateRequest(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Icon Create Request Error", err.Error())
		return
	}
	icon, err := iconOp.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Icon Create API Error", err.Error())
		return
	}
	plan.updateState(ctx, icon)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *iconResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state iconResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	icon := getIcon(ctx, r.client, sakuraCloudID(state.ID.ValueString()), &resp.State, &resp.Diagnostics)
	if icon == nil {
		return
	}
	state.updateState(ctx, icon)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *iconResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan iconResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutUpdate(ctx, plan.Timeouts, timeout5min)
	defer cancel()

	iconOp := iaas.NewIconOp(r.client)
	_, err := iconOp.Read(ctx, sakuraCloudID(plan.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Icon Read API Error", err.Error())
		return
	}

	_, err = iconOp.Update(ctx, sakuraCloudID(plan.ID.ValueString()), expandIconUpdateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Icon Update API Error", err.Error())
		return
	}

	gotIcon := getIcon(ctx, r.client, sakuraCloudID(plan.ID.ValueString()), &resp.State, &resp.Diagnostics)
	if gotIcon == nil {
		return
	}
	plan.updateState(ctx, gotIcon)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *iconResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state iconResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutDelete(ctx, state.Timeouts, timeout5min)
	defer cancel()

	iconOp := iaas.NewIconOp(r.client)
	icon := getIcon(ctx, r.client, sakuraCloudID(state.ID.ValueString()), &resp.State, &resp.Diagnostics)
	if icon == nil {
		return
	}
	if err := iconOp.Delete(ctx, icon.ID); err != nil {
		resp.Diagnostics.AddError("Icon Delete API Error", err.Error())
		return
	}
}

func (d *iconResourceModel) updateState(ctx context.Context, icon *iaas.Icon) {
	d.ID = types.StringValue(icon.ID.String())
	d.Name = types.StringValue(icon.Name)
	d.URL = types.StringValue(icon.URL)
	d.Tags = stringsToTset(icon.Tags)
}

func getIcon(ctx context.Context, client *APIClient, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.Icon {
	iconOp := iaas.NewIconOp(client)
	icon, err := iconOp.Read(ctx, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Icon Read API Error", err.Error())
		return nil
	}
	return icon
}

func expandIconBody(d *iconResourceModel) (string, error) {
	var body string
	if !d.Source.IsNull() {
		source := d.Source.ValueString()
		path, err := homedir.Expand(source)
		if err != nil {
			return "", fmt.Errorf("expanding homedir in source (%s) is failed: %s", source, err)
		}
		file, err := os.Open(filepath.Clean(path))
		if err != nil {
			return "", fmt.Errorf("opening SakuraCloud Icon source(%s) is failed: %s", source, err)
		}
		data, err := io.ReadAll(file)
		if err != nil {
			return "", fmt.Errorf("reading SakuraCloud Icon source file is failed: %s", err)
		}
		body = base64.StdEncoding.EncodeToString(data)
	} else if !d.Base64Content.IsNull() {
		body = d.Base64Content.ValueString()
	} else {
		return "", fmt.Errorf(`"source" or "base64content" field is required`)
	}
	return body, nil
}

func expandIconCreateRequest(d *iconResourceModel) (*iaas.IconCreateRequest, error) {
	body, err := expandIconBody(d)
	if err != nil {
		return nil, fmt.Errorf("creating SakuraCloud Icon is failed: %s", err)
	}
	return &iaas.IconCreateRequest{
		Name:  d.Name.ValueString(),
		Tags:  tsetToStrings(d.Tags),
		Image: body,
	}, nil
}

func expandIconUpdateRequest(d *iconResourceModel) *iaas.IconUpdateRequest {
	return &iaas.IconUpdateRequest{
		Name: d.Name.ValueString(),
		Tags: tsetToStrings(d.Tags),
	}
}
