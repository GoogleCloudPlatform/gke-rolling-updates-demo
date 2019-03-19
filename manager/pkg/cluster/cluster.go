package cluster

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/gke-rolling-updates-demo/manager/pkg/utils"

	container "cloud.google.com/go/container/apiv1"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

type ManagedCluster struct {
	Client      *container.ClusterManagerClient
	Cluster     *containerpb.Cluster
	Project     string
	Location    string
	ClusterName string
	NodeCount   int32
}

func NewManagedCluster(client *container.ClusterManagerClient, project string, location string, clusterName string, nodeCount int32) *ManagedCluster {
	return &ManagedCluster{
		Client:      client,
		Project:     project,
		Location:    location,
		ClusterName: clusterName,
		NodeCount:   nodeCount,
	}
}

func (c *ManagedCluster) Create(ctx context.Context) error {
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
	err = utils.WaitForOperation(ctx, c.Client, c.Project, c.Location, op.Name)
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

func (c *ManagedCluster) UpgradeControlPlane(ctx context.Context, version string) error {
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
	err = utils.WaitForOperation(ctx, c.Client, c.Project, c.Location, resp.Name)
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

func (c *ManagedCluster) UpgradeNodes(ctx context.Context, version string) error {
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
	err = utils.WaitForOperation(ctx, c.Client, c.Project, c.Location, resp.Name)
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

func (c *ManagedCluster) LatestMasterVersionForReleaseSeries(ctx context.Context, version string) (string, error) {
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
