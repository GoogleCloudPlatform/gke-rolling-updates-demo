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
	"strings"

	container "cloud.google.com/go/container/apiv1"
	"github.com/GoogleCloudPlatform/gke-rolling-updates-demo/manager/pkg/operation"
	log "github.com/sirupsen/logrus"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

const (
	retryInterval = 3
)

type GKECluster struct {
	Client      *container.ClusterManagerClient
	Cluster     *containerpb.Cluster
	ClusterName string
	Project     string
	Location    string
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

// Create is used to instantiate the Cluster struct embedded into the GKECluster struct.
// First it checks if a cluster matching the signature of GKECluster already exists.
// If one does, GKECluster's Cluster value will be attached to that cluster; otherwise,
// a new GKE cluster is created.
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
	opStatus := make(chan operation.Status)
	go operation.Wait(ctx, opStatus, retryInterval, c.Client, c.Project, c.Location, op.Name)

	for status := range opStatus {
		if status.Error != nil {
			return status.Error
		}

		if status.Status == containerpb.Operation_DONE {
			log.WithFields(log.Fields{
				"operation_id":     op.Name,
				"operation_status": status.Status,
			}).Info("Operation completed")
			break
		}

		log.WithFields(log.Fields{
			"operation_id":     op.Name,
			"operation_status": status.Status,
		}).Info("Waiting for operation")
	}

	resp, err = c.Client.GetCluster(ctx, getReq)
	if err != nil {
		return fmt.Errorf("unable to fetch created cluster metadata: %s", err)
	}
	c.Cluster = resp
	return nil
}

// UpgradeControlPlane takes in a version matching the following structures and upgrades
// the cluster to that version:
//
// 1.x
// 1.x.x
// 1.x.x-gke.x
//
func (c *GKECluster) UpgradeControlPlane(ctx context.Context, version string) error {
	req := &containerpb.UpdateClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", c.Project, c.Location, c.ClusterName),
		Update: &containerpb.ClusterUpdate{
			DesiredMasterVersion: version,
		},
	}
	op, err := c.Client.UpdateCluster(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to upgrade master version: %s", err)
	}

	opStatus := make(chan operation.Status)
	go operation.Wait(ctx, opStatus, retryInterval, c.Client, c.Project, c.Location, op.Name)

	for status := range opStatus {
		if status.Error != nil {
			return status.Error
		}

		if status.Status == containerpb.Operation_DONE {
			log.WithFields(log.Fields{
				"operation_id":     op.Name,
				"operation_status": status.Status,
			}).Info("Operation completed")
			break
		}

		log.WithFields(log.Fields{
			"operation_id":     op.Name,
			"operation_status": status.Status,
		}).Info("Waiting for operation")
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

// UpgradeNodes takes in a specified version string, validates that the specified version
// is valid according to the API and master versions, and either executes the version change
// or bails out.
func (c *GKECluster) UpgradeNodes(ctx context.Context, version string) error {
	req := &containerpb.UpdateClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", c.Project, c.Location, c.ClusterName),
		Update: &containerpb.ClusterUpdate{
			DesiredNodeVersion: c.Cluster.GetCurrentMasterVersion(),
		},
	}
	op, err := c.Client.UpdateCluster(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to upgrade master version: %s", err)
	}

	opStatus := make(chan operation.Status)
	go operation.Wait(ctx, opStatus, retryInterval, c.Client, c.Project, c.Location, op.Name)

	for status := range opStatus {
		if status.Error != nil {
			return status.Error
		}

		if status.Status == containerpb.Operation_DONE {
			log.WithFields(log.Fields{
				"operation_id":     op.Name,
				"operation_status": status.Status,
			}).Info("Operation completed")
			break
		}

		log.WithFields(log.Fields{
			"operation_id":     op.Name,
			"operation_status": status.Status,
		}).Info("Waiting for operation")
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

// LatestMasterVersionForReleaseSeries takes in a version and returns the latest
// version in that series, i.e. if given version 1.9, it will return the latest
// version in the 1.9.x-gke.x series.
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

// LatestNodeVersionForReleaseSeries takes in a version and returns the latest
// version in that series that does not outpace the current master version (if
// any).
func (c *GKECluster) LatestNodeVersionForReleaseSeries(ctx context.Context, version string) (string, error) {
	req := &containerpb.GetServerConfigRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s", c.Project, c.Location),
	}
	resp, err := c.Client.GetServerConfig(ctx, req)
	if err != nil {
		return "", fmt.Errorf("unable to get container engine versions: %s", err)
	}

	if version == "latest" {
		return c.Cluster.GetCurrentMasterVersion(), nil
	}

	valid, err := getValidNodeVersion(resp.ValidNodeVersions, version, c.Cluster.GetCurrentMasterVersion())
	log.Infof("valid: %v, err %v", valid, err)
	if err != nil {
		return "", err
	}

	if valid != "" {
		return valid, nil
	}

	return "", fmt.Errorf("Unable to find a version in that series")
}

// getValidNodeVersion takes in a list of valid versions, the requested version, and the current master version
// and determines the latest valid node version given those constraints.
func getValidNodeVersion(validVersions []string, requestedVersion string, masterVersion string) (string, error) {
	log.WithFields(log.Fields{
		"master_version":    masterVersion,
		"requested_version": requestedVersion,
	}).Info("Determining if requested version is valid")

	parsedRequestedVersion := strings.Split(requestedVersion, ".")
	if len(parsedRequestedVersion) > 2 {
		parsedRequestedVersion[2] = strings.Split(parsedRequestedVersion[2], "-")[0]
	}

	if len(parsedRequestedVersion) > 4 {
		return "", fmt.Errorf("unexpected requested version: %s", requestedVersion)
	}

	parsedMasterVersion := strings.Split(masterVersion, ".")
	parsedMasterVersion[2] = strings.Split(parsedMasterVersion[2], "-")[0]

	if len(parsedMasterVersion) > 4 {
		return "", fmt.Errorf("unexpected master version: %s", masterVersion)
	}

	for i := range requestedVersion {
		if requestedVersion[i] > masterVersion[i] {
			log.WithFields(log.Fields{
				"requested_version": requestedVersion,
				"masterVersion":     masterVersion,
			}).Info("Requested version is greater than the current master version")
			return "", nil
		}
	}

	for _, v := range validVersions {
		isValid := true
		log.WithFields(log.Fields{
			"valid_node_version": v,
			"requested_version":  requestedVersion,
		}).Info("Comparing valid node version against requested version")
		parsedValidVersion := strings.Split(v, ".")
		parsedValidVersion[2] = strings.Split(parsedValidVersion[2], "-")[0]

		if len(parsedValidVersion) > 4 {
			return "", fmt.Errorf("unexpected valid version: %s", v)
		}

		for i := range parsedRequestedVersion {
			if parsedRequestedVersion[i] != parsedValidVersion[i] {
				log.WithFields(log.Fields{
					"field":             i,
					"requested_version": parsedRequestedVersion[i],
					"valid_version":     parsedValidVersion[i],
				}).Info("Requested version is not in the currently selected valid series...continuing")
				isValid = false
				break
			} else {
				log.WithFields(log.Fields{
					"field":             i,
					"requested_version": parsedRequestedVersion[i],
					"valid_version":     parsedValidVersion[i],
				}).Info("Requested version possibly in the currently selected valid series...continuing")
			}
		}

		if isValid == true {
			log.WithFields(log.Fields{
				"valid_version": v,
			}).Info("Found valid version")
			return v, nil
		}
	}

	return "", fmt.Errorf("unable to find a version in that series: requested version %s", requestedVersion)
}
