/*
Copyright 2019 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package operation

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	container "cloud.google.com/go/container/apiv1"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

type Status struct {
	Status containerpb.Operation_Status
	Error  error
}

// Wait waits for a GKE operation (not a project operation) to complete and
// outputs current status to the provided channel
func Wait(ctx context.Context, opChan chan Status, retryInterval int, client *container.ClusterManagerClient, project string, location string, operationID string) {
	for {
		req := &containerpb.GetOperationRequest{
			Name:        fmt.Sprintf("projects/%s/locations/%s", project, location),
			OperationId: operationID,
		}

		resp, err := client.GetOperation(ctx, req)
		opChan <- Status{Status: resp.Status, Error: err}

		if resp.Status == containerpb.Operation_DONE {
			close(opChan)
			break
		}

		log.WithFields(log.Fields{
			"type":   resp.OperationType,
			"status": resp.Status,
		}).Info("Waiting for operation")

		toSleep, _ := time.ParseDuration(fmt.Sprintf("%ds", retryInterval))
		time.Sleep(toSleep)
	}
}
