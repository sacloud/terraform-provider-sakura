// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package object_storage

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	ver "github.com/sacloud/terraform-provider-sakura/version"
)

type objectStorageS3CompatModel struct {
	ID        types.String `tfsdk:"id"`
	Endpoint  types.String `tfsdk:"endpoint"`
	AccessKey types.String `tfsdk:"access_key"`
	SecretKey types.String `tfsdk:"secret_key"`
	Bucket    types.String `tfsdk:"bucket"`
}

type objectStorageObjectBaseModel struct {
	objectStorageS3CompatModel
	Key          types.String `tfsdk:"key"`
	ETag         types.String `tfsdk:"etag"`
	LastModified types.String `tfsdk:"last_modified"`
	VersionID    types.String `tfsdk:"version_id"`
	ContentType  types.String `tfsdk:"content_type"`
	StorageClass types.String `tfsdk:"storage_class"`
}

func (model *objectStorageS3CompatModel) updateS3State() {
	model.ID = types.StringValue(model.Bucket.ValueString())
	model.Endpoint = types.StringValue(getEndpoint(model.Endpoint.ValueString()))
}

func (model *objectStorageObjectBaseModel) updateState(objInfo *minio.ObjectInfo) {
	model.ID = types.StringValue(model.Bucket.ValueString() + "/" + model.Key.ValueString())
	model.Endpoint = types.StringValue(getEndpoint(model.Endpoint.ValueString()))
	model.ETag = types.StringValue(objInfo.ETag)
	model.LastModified = types.StringValue(objInfo.LastModified.String())
	model.VersionID = types.StringValue(objInfo.VersionID)
	model.ContentType = types.StringValue(objInfo.ContentType)
	model.StorageClass = types.StringValue(objInfo.StorageClass)
}

func getEndpoint(endpoint string) string {
	if endpoint == "" {
		return "s3.isk01.sakurastorage.jp"
	}
	return endpoint
}

func (model *objectStorageS3CompatModel) getMinIOClient() (*minio.Client, error) {
	endpoint := getEndpoint(model.Endpoint.ValueString())
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(model.AccessKey.ValueString(), model.SecretKey.ValueString(), ""),
		Region: "jp-north-1", Secure: true, BucketLookup: minio.BucketLookupPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}
	client.SetAppInfo("terraform-provider-sakura", ver.Version)
	return client, nil
}
