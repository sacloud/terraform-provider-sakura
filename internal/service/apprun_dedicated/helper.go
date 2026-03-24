// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	lb "github.com/sacloud/apprun-dedicated-api-go/apis/loadbalancer"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	wn "github.com/sacloud/apprun-dedicated-api-go/apis/workernode"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

func listed[T, U any](yield func(*U) ([]T, *U, error)) (ret []T, err error) {
	var cursor *U

	for {
		var items []T

		items, cursor, err = yield(cursor)
		ret = append(ret, items...)

		if err != nil {
			return
		}

		if cursor == nil {
			break
		}
	}
	return
}

// This `~[16]byte` is ugly!
// But we have no other possible way to do this.
func intoUUID[T ~[16]byte](v types.String) (t T, err error) {
	u, err := uuid.Parse(v.ValueString())

	if err == nil {
		t = T(u)
	}

	return
}

// This `~[16]byte` is ugly!
// But we have no other possible way to do this.
func uuid2StringValue[T ~[16]byte](t T) types.String {
	return types.StringValue(uuid.UUID(t).String())
}

func intoRFC2822[T ~int | ~int64](t T) types.String {
	return types.StringValue(time.Unix(common.ToInt64(t), 0).Format(time.RFC822))
}

func intoInt32[T ~int | ~int32 | ~int16 | ~uint16 | ~int64](t *T) types.Int32 {
	if t == nil {
		return types.Int32Null()
	}

	return types.Int32Value(common.ToInt32(*t))
}

//////////////////////////////////////////////////////////////

func deleteLBs(ctx context.Context, api lb.LoadBalancerAPI) (err error) {
	err = provisionLBs(ctx, api)

	if err != nil {
		return
	}

	return waitLBs(ctx, api)
}

func provisionLBs(ctx context.Context, api lb.LoadBalancerAPI) error {
	t := time.NewTicker(7 * time.Second)
	defer t.Stop()

	for {
		ok, err := provisionLBsInternal(ctx, api)

		if saclient.IsNotFoundError(err) {
			return nil // no lb no problem
		}

		if err != nil {
			return err
		}

		if ok {
			return nil
		}

		select {
		case <-t.C:
			continue

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func provisionLBsInternal(ctx context.Context, api lb.LoadBalancerAPI) (bool, error) {
	list, err := listed(func(i *v1.LoadBalancerID) ([]v1.ReadLoadBalancerSummary, *v1.LoadBalancerID, error) {
		return api.List(ctx, 10, i)
	})

	if saclient.IsNotFoundError(err) {
		return true, nil // no lb no problem
	}

	if err != nil {
		return false, err
	}

	ok := true
	err = errors.Join(common.MapTo(list, func(i v1.ReadLoadBalancerSummary) error {
		o, e := provisionLBsInternalNodes(ctx, api, i.LoadBalancerID)
		ok = ok && o
		return e
	})...)

	return ok, err
}

func provisionLBsInternalNodes(ctx context.Context, api lb.LoadBalancerAPI, id v1.LoadBalancerID) (bool, error) {
	lb, err := api.Read(ctx, id)

	if saclient.IsNotFoundError(err) {
		return true, nil // no lb no problem
	}

	if err != nil {
		return false, err
	}

	if lb.Deleting {
		return true, nil // provisioned, ok to proceed
	}

	list, err := api.ListNodes(ctx, id, 10, nil)

	if saclient.IsNotFoundError(err) {
		return true, nil // no lb no problem
	}

	if err != nil {
		return false, err
	}

	ok := true
	err = errors.Join(common.MapTo(list, func(i v1.ReadLoadBalancerNodeSummary) error {
		o, e := provisionLBsInternalNode(ctx, api, id, i.LoadBalancerNodeID)
		ok = ok && o
		return e
	})...)

	return ok, err
}

func provisionLBsInternalNode(ctx context.Context, api lb.LoadBalancerAPI, lid v1.LoadBalancerID, nid v1.LoadBalancerNodeID) (bool, error) {
	node, err := api.ReadNode(ctx, lid, nid)

	if saclient.IsNotFoundError(err) {
		return true, nil // no lb no problem
	}

	if err != nil {
		return false, err
	}

	switch node.Status {
	case v1.LoadBalancerNodeStatusHealthy:
		tflog.Debug(ctx, "ASG LB Node healthy", map[string]any{"id": uuid.UUID(nid).String()})
		return true, nil

	default:
		tflog.Debug(ctx, "ASG LB Node unhealthy", map[string]any{
			"id":     uuid.UUID(nid).String(),
			"status": node.Status,
		})
		return false, nil
	}
}

func waitLBs(ctx context.Context, api lb.LoadBalancerAPI) error {
	t := time.NewTicker(7 * time.Second)
	defer t.Stop()

	for {
		ok, err := waitLBsInternal(ctx, api)

		if saclient.IsNotFoundError(err) {
			return nil // no lb no problem
		}

		if err != nil {
			return err
		}

		if ok {
			return nil
		}

		select {
		case <-t.C:
			continue

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func waitLBsInternal(ctx context.Context, api lb.LoadBalancerAPI) (bool, error) {
	list, err := listed(func(i *v1.LoadBalancerID) ([]v1.ReadLoadBalancerSummary, *v1.LoadBalancerID, error) {
		return api.List(ctx, 10, i)
	})

	if saclient.IsNotFoundError(err) {
		return true, nil // no lb no problem
	}

	if err != nil {
		return false, err
	}

	ok := true
	err = errors.Join(common.MapTo(list, func(i v1.ReadLoadBalancerSummary) error {
		o, e := waitLBsInternalLB(ctx, api, i.LoadBalancerID)
		ok = ok && o
		return e
	})...)

	return ok, err
}

func waitLBsInternalLB(ctx context.Context, api lb.LoadBalancerAPI, id v1.LoadBalancerID) (bool, error) {
	lb, err := api.Read(ctx, id)

	if saclient.IsNotFoundError(err) {
		return true, nil // no lb no problem
	}

	if err != nil {
		return false, err
	}

	if lb.Deleting {
		// OK this is likely
		return false, nil
	}

	err = api.Delete(ctx, lb.LoadBalancerID)

	if saclient.IsNotFoundError(err) {
		return true, nil // no lb no problem
	}

	if err != nil {
		return false, err
	}

	// reaching here means "do it again" situation
	tflog.Debug(ctx, "ASG LB deleting", map[string]any{"id": uuid.UUID(id).String()})
	return false, nil
}

//////////////////////////////////////////////////////////////

func deleteWNs(ctx context.Context, api wn.WorkerNodeAPI) (err error) {
	t := time.NewTicker(7 * time.Second)
	defer t.Stop()

	for {
		ok, err := drainWNs(ctx, api)

		if saclient.IsNotFoundError(err) {
			return nil // no lb no problem
		}

		if err != nil {
			return err
		}

		if ok {
			return nil
		}

		select {
		case <-t.C:
			continue

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func drainWNs(ctx context.Context, api wn.WorkerNodeAPI) (bool, error) {
	list, err := listed(func(i *v1.WorkerNodeID) ([]wn.WorkerNodeDetail, *v1.WorkerNodeID, error) {
		return api.List(ctx, 10, i)
	})

	if saclient.IsNotFoundError(err) {
		return true, nil // no lb no problem
	}

	if err != nil {
		return false, err
	}

	ok := true
	err = errors.Join(common.MapTo(list, func(i wn.WorkerNodeDetail) error {
		o, e := drainWNsNode(ctx, api, i.WorkerNodeID)
		ok = ok && o
		return e
	})...)

	return ok, err
}

func drainWNsNode(ctx context.Context, api wn.WorkerNodeAPI, id v1.WorkerNodeID) (bool, error) {
	wn, err := api.Read(ctx, id)

	if saclient.IsNotFoundError(err) {
		return true, nil // no lb no problem
	}

	if err != nil {
		return false, err
	}

	if wn.Creating {
		tflog.Debug(ctx, "ASG Worker Node under provision", map[string]any{"id": uuid.UUID(id).String()})
		return false, nil
	}

	err = api.Update(ctx, id, true)

	if saclient.IsNotFoundError(err) {
		return true, nil // no lb no problem
	}

	if err != nil {
		return false, err
	}

	tflog.Debug(ctx, "ASG Worker Node draining", map[string]any{"id": uuid.UUID(id).String()})
	return true, nil
}
