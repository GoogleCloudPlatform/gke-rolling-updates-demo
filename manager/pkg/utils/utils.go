package utils

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	container "cloud.google.com/go/container/apiv1"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

func WaitForOperation(ctx context.Context, client *container.ClusterManagerClient, project string, location string, operationId string) error {
	retryTime := 3
	for {
		req := &containerpb.GetOperationRequest{
			Name:        fmt.Sprintf("projects/%s/locations/%s", project, location),
			OperationId: operationId,
		}

		resp, err := client.GetOperation(ctx, req)
		if err != nil {
			return fmt.Errorf("unable to get operation: %s", err)
		}

		if resp.Status == containerpb.Operation_DONE {
			break
		}

		log.WithFields(log.Fields{
			"type":   resp.OperationType,
			"status": resp.Status,
		}).Info("waiting for operation")

		toSleep, _ := time.ParseDuration(fmt.Sprintf("%ds", retryTime))
		time.Sleep(toSleep)
	}

	return nil
}
