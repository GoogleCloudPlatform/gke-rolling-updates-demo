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

// variables.tf - this is where all variables are defined.  The user must
// provide these for any invocation of `terraform plan`, `apply`, or `destroy`.

variable "region" {
  description = "GCP region where Kubernetes Engine cluster would be created"
}

variable "zone" {
  description = "GCP zone where Kubernetes Engine cluster would be created"
}

variable "control_plane_version" {
  description = "GCP Kubernetes Engine cluster control plane version"
}

variable "node_pool_version" {
  description = "GCP Kubernetes Engine cluster node pool version"
}

variable "num_nodes" {
  description = "GCP Kubernetes Engine cluster number of worker nodes per zone"
}

variable "machine_type" {
  description = "GCP Kubernetes Engine cluster node machine type"
}

variable "timeout_create" {
  description = "Timeout to create a test cluster and sampple app"
  default     = "30m"
}

variable "timeout_update" {
  description = "Timeout to update a test cluster"
  default     = "25m"
}

variable "timeout_delete" {
  description = "Timeout to delete a test cluster and sampple app"
  default     = "25m"
}
