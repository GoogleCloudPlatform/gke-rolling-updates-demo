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

// output.tf - by convention, this is where output values are defined
// outputs will be displayed at the end of a `terraform apply`
// These outputs will display all available GKE versions based on the zone
// variable defined by the user


// Display the available control plane versions for your zone
output "available_control_plane_versions" {
  value = data.google_container_engine_versions.my_zone.valid_master_versions
}

// Display the available node pool versions for your zone
output "available_node_versions" {
  value = data.google_container_engine_versions.my_zone.valid_node_versions
}
