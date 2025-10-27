// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/encrypt"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type objectStorageObjectResource struct{}

var (
	_ resource.Resource                = &objectStorageObjectResource{}
	_ resource.ResourceWithImportState = &objectStorageObjectResource{}
)

func NewObjectStorageObjectResource() resource.Resource {
	return &objectStorageObjectResource{}
}

func (r *objectStorageObjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_object"
}

type objectStorageObjectResourceModel struct {
	objectStorageObjectBaseModel
	ACL                  types.String   `tfsdk:"acl"`
	Source               types.String   `tfsdk:"source"`
	Content              types.String   `tfsdk:"content"`
	ContentBase64        types.String   `tfsdk:"content_base64"`
	ContentLanguage      types.String   `tfsdk:"content_language"`
	ContentEncoding      types.String   `tfsdk:"content_encoding"`
	CacheControl         types.String   `tfsdk:"cache_control"`
	ServerSideEncryption types.String   `tfsdk:"server_side_encryption"`
	UserMetadata         types.Map      `tfsdk:"user_metadata"`
	UserTags             types.Map      `tfsdk:"user_tags"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
	//Expires         types.String   `tfsdk:"expires"`
}

func (r *objectStorageObjectResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaResourceId("Object Storage Object"),
			"endpoint":   SchemaResourceEndpoint("Object Storage Object"),
			"access_key": SchemaResourceAccessKey("Object Storage Object"),
			"secret_key": SchemaResourceSecretKey("Object Storage Object"),
			"bucket":     SchemaResourceBucket("Object Storage Object"),
			"key": schema.StringAttribute{
				Required:    true,
				Description: "The key of the Object Storage Object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"acl": schema.StringAttribute{
				Optional:    true,
				Description: "The ACL of the Object Storage Object.",
				Validators: []validator.String{
					stringvalidator.OneOf("private", "public-read"),
				},
			},
			"source": schema.StringAttribute{
				Optional:    true,
				Description: "The path to a file that will be uploaded as the Object Storage Object. Conflicts with `content` and `content_base64`.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("content"), path.MatchRelative().AtParent().AtName("content_base64")),
				},
			},
			"content": schema.StringAttribute{
				Optional:    true,
				Description: "The content of the Object Storage Object. Conflicts with `source` and `content_base64`.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("source"), path.MatchRelative().AtParent().AtName("content_base64")),
				},
			},
			"content_base64": schema.StringAttribute{
				Optional:    true,
				Description: "The base64-encoded content of the Object Storage Object. Conflicts with `source` and `content`.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("source"), path.MatchRelative().AtParent().AtName("content")),
				},
			},
			"last_modified": schema.StringAttribute{
				Computed:    true,
				Description: "The last modified time of the Object Storage Object.",
			},
			"etag": schema.StringAttribute{
				Computed:    true,
				Description: "The ETag of the Object Storage Object.",
			},
			"cache_control": schema.StringAttribute{
				Optional:    true,
				Description: "The cache control setting for the Object Storage object",
			},
			"content_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The content type of the Object Storage Object.",
			},
			"content_language": schema.StringAttribute{
				Optional:    true,
				Description: "The content language of the Object Storage Object.",
			},
			"content_encoding": schema.StringAttribute{
				Optional:    true,
				Description: "The content encoding of the Object Storage Object.",
			},
			"storage_class": schema.StringAttribute{
				Computed:    true,
				Description: "The storage class of the Object Storage Object.",
				/*
					Default:     stringdefault.StaticString("STANDARD"),
					Validators: []validator.String{
						stringvalidator.OneOf("STANDARD"),
					},
				*/
			},
			"server_side_encryption": schema.StringAttribute{
				// TODO: Support KMS and SSE-C in minio
				Optional:    true,
				Description: "The server-side encryption algorithm to use for the Object Storage Object. Supported value is now `AES256(S3)`.",
				Validators: []validator.String{
					stringvalidator.OneOf("S3", "AES256"),
				},
			},
			"user_metadata": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A map of user-defined metadata to store with the Object Storage Object.",
			},
			"user_tags": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A map of user-defined tags to store with the Object Storage Object.",
			},
			"version_id": schema.StringAttribute{
				Computed:    true,
				Description: "The version ID of the Object Storage Object.",
			},
			//"expires": schema.StringAttribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *objectStorageObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *objectStorageObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan objectStorageObjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	if err := uploadObject(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state objectStorageObjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := state.getMinIOClient()
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}

	objInfo, err := client.StatObject(ctx, state.Bucket.ValueString(), state.Key.ValueString(), minio.GetObjectOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to get object: %s", err.Error()))
		return
	}

	state.updateState(&objInfo)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *objectStorageObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan objectStorageObjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	if err := uploadObject(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state objectStorageObjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	client, err := state.getMinIOClient()
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}

	if state.VersionID.ValueString() == "" {
		err = client.RemoveObject(ctx, state.Bucket.ValueString(), state.Key.ValueString(), minio.RemoveObjectOptions{ForceDelete: true})
		if err != nil {
			resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to delete object: %s", err.Error()))
			return
		}
	} else {
		objCh := make(chan minio.ObjectInfo)
		go func() {
			defer close(objCh)
			opts := minio.ListObjectsOptions{Prefix: state.Key.ValueString(), WithVersions: true, Recursive: true}
			for object := range client.ListObjects(ctx, state.Bucket.ValueString(), opts) {
				if object.Err != nil {
					resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to list object with key(%s): %s", state.Key.ValueString(), object.Err.Error()))
				}
				objCh <- object
			}
		}()

		errCh := client.RemoveObjects(ctx, state.Bucket.ValueString(), objCh, minio.RemoveObjectsOptions{})
		for e := range errCh {
			resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to delete %s/%s, error: %s", e.ObjectName, e.VersionID, e.Err.Error()))
		}
	}
}

func uploadObject(ctx context.Context, model *objectStorageObjectResourceModel) error {
	var body io.ReadSeeker
	if model.Source.ValueString() != "" {
		path, err := common.ExpandHomeDir(model.Source.ValueString())
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file(%s): %w", path, err)
		}
		body = file
		defer func() {
			err := file.Close()
			if err != nil {
				tflog.Warn(ctx, fmt.Sprintf("failed to close source file(%s): %s", path, err))
			}
		}()
	} else if !model.Content.IsNull() && !model.Content.IsUnknown() {
		body = bytes.NewReader([]byte(model.Content.ValueString()))
	} else if model.ContentBase64.ValueString() != "" {
		contentRaw, err := base64.StdEncoding.DecodeString(model.ContentBase64.ValueString())
		if err != nil {
			return fmt.Errorf("failed to decode content_base64: %w", err)
		}
		body = bytes.NewReader(contentRaw)
	} else {
		return fmt.Errorf("one of source, content or content_base64 must be specified")
	}

	client, err := model.getMinIOClient()
	if err != nil {
		return err
	}

	opts := minio.PutObjectOptions{StorageClass: model.StorageClass.ValueString()}
	contentType := model.ContentType.ValueString()
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	opts.ContentType = contentType
	if !model.ContentType.IsNull() && !model.ContentType.IsUnknown() {
		opts.ContentType = model.ContentType.ValueString()
	}
	if !model.ContentLanguage.IsNull() && !model.ContentLanguage.IsUnknown() {
		opts.ContentLanguage = model.ContentLanguage.ValueString()
	}
	if !model.ContentEncoding.IsNull() && !model.ContentEncoding.IsUnknown() {
		opts.ContentEncoding = model.ContentEncoding.ValueString()
	}
	if !model.CacheControl.IsNull() && !model.CacheControl.IsUnknown() {
		opts.CacheControl = model.CacheControl.ValueString()
	}
	if !model.UserMetadata.IsNull() && !model.UserMetadata.IsUnknown() {
		opts.UserMetadata = common.TmapToStrMap(model.UserMetadata)
	}
	if !model.UserTags.IsNull() && !model.UserTags.IsUnknown() {
		opts.UserTags = common.TmapToStrMap(model.UserTags)
	}
	if !model.ServerSideEncryption.IsNull() && !model.ServerSideEncryption.IsUnknown() {
		sse := getSSE(model.ServerSideEncryption.ValueString())
		if sse == nil {
			return fmt.Errorf("unsupported server_side_encryption: %s", model.ServerSideEncryption.ValueString())
		}
		opts.ServerSideEncryption = sse
	}
	if !model.ACL.IsNull() && !model.ACL.IsUnknown() {
		acl := strings.ToLower(model.ACL.ValueString())
		if opts.UserMetadata == nil {
			opts.UserMetadata = map[string]string{
				"x-amz-acl": acl,
			}
		} else {
			opts.UserMetadata["x-amz-acl"] = acl
		}
	}

	_, err = client.PutObject(ctx, model.Bucket.ValueString(), model.Key.ValueString(), body, -1, opts)
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	objInfo, err := client.StatObject(ctx, model.Bucket.ValueString(), model.Key.ValueString(), minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object information: %s", err.Error())
	}

	model.updateState(&objInfo)

	return nil
}

func getSSE(algorithm string) (sse encrypt.ServerSide) {
	switch strings.ToUpper(algorithm) {
	case "S3", "AES256":
		return encrypt.NewSSE()
	}
	return nil
}
