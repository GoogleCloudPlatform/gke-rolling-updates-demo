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

package cluster

import (
	"context"
	"fmt"

	container "cloud.google.com/go/container/apiv1"
	"github.com/GoogleCloudPlatform/gke-rolling-updates-demo/manager/pkg/operation"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

type GKECluster struct {
	Client      *container.ClusterManagerClient
	Cluster     *containerpb.Cluster
	Project     string
	Location    string
	ClusterName string
	NodeCount   int32
}

func NewGKECluster(client *container.ClusterManagerClient, project string, location string, clusterName string, nodeCount int32) *GKECluster {
	return &GKECluster{
		Client:      client,
		Project:     project,
		Location:    location,
		ClusterName: clusterName,
		NodeCount:   nodeCount,
	}
}

func (c *GKECluster) Create(ctx context.Context) error {
	getReq := &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", c.Project, c.Location, c.ClusterName),
	}
	resp, err := c.Client.GetCluster(ctx, getReq)
	if resp != nil {
		c.Cluster = resp
		return nil
	}
	req := &containerpb.CreateClusterRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", c.Project, c.Location),
		Cluster: &containerpb.Cluster{
			Name:             c.ClusterName,
			InitialNodeCount: c.NodeCount,
		},
	}
	op, err := c.Client.CreateCluster(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to create cluster: %s", err)
	}
	err = operation.WaitForOperation(ctx, c.Client, c.Project, c.Location, op.Name)
	if err != nil {
		return fmt.Errorf("error waiting for operation: %s", err)
	}

	resp, err = c.Client.GetCluster(ctx, getReq)
	if err != nil {
		return fmt.Errorf("unable to fetch created cluster metadata: %s", err)
	}
	c.Cluster = resp
	return nil
}

func (c *GKECluster) UpgradeControlPlane(ctx context.Context, version string) error {
	req := &containerpb.UpdateClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", c.Project, c.Location, c.ClusterName),
		Update: &containerpb.ClusterUpdate{
			DesiredMasterVersion: version,
		},
	}
	resp, err := c.Client.UpdateCluster(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to upgrade master version: %s", err)
	}
	err = operation.WaitForOperation(ctx, c.Client, c.Project, c.Location, resp.Name)
	if err != nil {
		return fmt.Errorf("error waiting for operation: %s", err)
	}

	getReq := &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", c.Project, c.Location, c.ClusterName),
	}
	getResp, err := c.Client.GetCluster(ctx, getReq)
	if err != nil {
		return fmt.Errorf("unable to fetch created cluster metadata: %s", err)
	}
	c.Cluster = getResp

	return nil
}

func (c *GKECluster) UpgradeNodes(ctx context.Context, version string) error {
	req := &containerpb.UpdateClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", c.Project, c.Location, c.ClusterName),
		Update: &containerpb.ClusterUpdate{
			DesiredNodeVersion: c.Cluster.GetCurrentMasterVersion(),
		},
	}
	resp, err := c.Client.UpdateCluster(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to upgrade master version: %s", err)
	}
	err = operation.WaitForOperation(ctx, c.Client, c.Project, c.Location, resp.Name)
	if err != nil {
		return fmt.Errorf("error waiting for operation: %s", err)
	}

	getReq := &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", c.Project, c.Location, c.ClusterName),
	}
	getResp, err := c.Client.GetCluster(ctx, getReq)
	if err != nil {
		return fmt.Errorf("unable to fetch created cluster metadata: %s", err)
	}
	c.Cluster = getResp

	return nil
}

func (c *GKECluster) LatestMasterVersionForReleaseSeries(ctx context.Context, version string) (string, error) {
	req := &containerpb.GetServerConfigRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s", c.Project, c.Location),
	}
	resp, err := c.Client.GetServerConfig(ctx, req)
	if err != nil {
		return "", fmt.Errorf("unable to get container engine versions: %s", err)
	}

	if version == "latest" {
		return resp.ValidMasterVersions[0], nil
	}

	for _, v := range resp.ValidMasterVersions {
		if v[:len(version)] == version {
			return v, nil
		}
	}

	return "", fmt.Errorf("unable to find a version in that series")
}
