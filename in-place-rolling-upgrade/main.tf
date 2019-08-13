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

// main.tf - this file contains all data sources and resources.
// This will create a GKE cluster with the values supplied in the region,
// control_plane_version, and node_pool_version variables.

// Data Source to find avaiable gke versions in a specific zone
// https://www.terraform.io/docs/providers/google/d/google_container_engine_versions.html
data "google_container_engine_versions" "my_zone" {
  zone = var.zone
}

// This resource creates an HA Regional GKE cluster
// https://www.terraform.io/docs/providers/google/r/container_cluster.html
resource "google_container_cluster" "test" {
  name   = var.cluster_name
  region = var.region

  // in a regional cluster, this is the number of nodes per zone
  initial_node_count = var.num_nodes

  min_master_version = var.control_plane_version
  node_version       = var.node_pool_version

  // We specify the machine type for the node pool instances.
  node_config {
    machine_type = var.machine_type

    metadata = {
      disable-legacy-endpoints = "true"
    }
  }

  // Some operations have been observed to take longer thant the defaults
  timeouts {
    create = var.timeout_create
    update = var.timeout_update
    delete = var.timeout_delete
  }
}
