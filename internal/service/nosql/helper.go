// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package nosql

import (
	"context"
	"fmt"
	"time"

	"github.com/sacloud/nosql-api-go"
	v1 "github.com/sacloud/nosql-api-go/apis/v1"
)

func waitNosqlReady(ctx context.Context, client *v1.Client, id string) error {
	dbOp := nosql.NewDatabaseOp(client)
	errCount := 0

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("failed to wait for NoSQL[%s] ready: %s", id, waitCtx.Err())
		default:
			res, err := dbOp.Read(ctx, id)
			if err != nil {
				errCount += 1
				if errCount > 5 {
					return fmt.Errorf("exceeds 5 retry limit during NoSQL[%s] ready check: %w", id, err)
				}
				time.Sleep(10 * time.Second)
				continue
			}
			health, err := nosql.NewInstanceOp(client, res.ID.Value, res.Remark.Value.Nosql.Value.Zone.Value).GetNodeHealth(ctx)
			if err != nil {
				errCount += 1
				if errCount > 5 {
					return fmt.Errorf("exceeds 5 retry limit during NoSQL[%s] ready check: %w", id, err)
				}
				time.Sleep(10 * time.Second)
			} else {
				if res.Availability.Value == "available" && res.Instance.Value.Status.Value == "up" &&
					string(health) == "healthy" {
					return nil
				}
				time.Sleep(30 * time.Second)
			}
		}
	}
}

func waitNosqlDown(ctx context.Context, client *v1.Client, id string) error {
	dbOp := nosql.NewDatabaseOp(client)
	errCount := 0

	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("failed to wait for NoSQL down: %s", waitCtx.Err())
		default:
			res, err := dbOp.Read(ctx, id)
			if err != nil {
				errCount += 1
				if errCount > 5 {
					return fmt.Errorf("exceeds 5 retry limit in down check: %w", err)
				}
				time.Sleep(10 * time.Second)
			} else {
				if res.Instance.Value.Status.Value == "down" {
					return nil
				}
				time.Sleep(20 * time.Second)
			}
		}
	}
}

func waitNosqlProcessingDone(ctx context.Context, client *v1.Client, id string, jobType string) error {
	dbOp := nosql.NewDatabaseOp(client)
	errCount := 0

	waitCtx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("failed to wait for NoSQL %s processing: %s", jobType, waitCtx.Err())
		default:
			res, err := dbOp.GetStatus(ctx, id)
			if err != nil {
				errCount += 1
				if errCount > 5 {
					return fmt.Errorf("exceeds 5 retry limit in %s processing: %w", jobType, err)
				}
				time.Sleep(10 * time.Second)
			} else {
				for _, job := range res.Jobs {
					if job.JobType.Value == jobType {
						if job.JobStatus.Value == "Done" {
							return nil
						}
					}
				}
				time.Sleep(20 * time.Second)
			}
		}
	}
}
