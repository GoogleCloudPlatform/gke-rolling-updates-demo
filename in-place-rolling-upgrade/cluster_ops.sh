#!/bin/bash
# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# cluster_ops.sh - a runner script to create, upgrade, and delete gke clusters
# with the help of kubectl, terraform, and gcloud

# shellcheck source=.env

set -euo pipefail

SCRIPT_HOME="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_HOME="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

fail() {
  echo "ERROR: ${*}"
  exit 2
}

# Validate that this workstation has access to the required executables
command -v kubectl >/dev/null || fail "kubectl is not installed!"
command -v terraform >/dev/null || fail "terraform is not installed!"
command -v gcloud >/dev/null || fail "gcloud is not installed!"

# Print usage when needed
usage() {
  cat <<-EOM
USAGE: $(basename "$0") <action>
Where the <action> can be:
  auto
  create
  upgrade-control
  upgrade-nodes
  downgrade-nodes
  delete
EOM
  exit 1
}

# If no parameters are provided, print usage
if [[ $# -lt 1 ]]; then
  usage
fi

# Source the configuration file if it exists
if [ -f "${REPO_HOME}/.env" ] ; then
  source "${REPO_HOME}/.env"
fi

# Set GCLOUD_ZONE to default if it has not yet been set
if [ -z ${GCLOUD_ZONE:+exists} ]; then
  GCLOUD_ZONE=$(gcloud config get-value compute/zone)
  export GCLOUD_ZONE
  if [ "$GCLOUD_ZONE" == "(unset)" ]; then
    fail "GCLOUD_ZONE is not set"
  fi
fi

# Set GCLOUD_REGION to default if it has not yet been set
if [ -z ${GCLOUD_REGION:+exists} ]; then
  GCLOUD_REGION=$(gcloud config get-value compute/region)
  export GCLOUD_REGION
  if [ "${GCLOUD_REGION}" == "(unset)" ]; then
    fail "GCLOUD_REGION is not set"
  fi
fi

# Set GCLOUD_PROJECT to default if it has not yet been set
if [ -z ${GCLOUD_PROJECT:+exists} ]; then
  GCLOUD_PROJECT=$(gcloud config get-value core/project)
  export GCLOUD_PROJECT
  if [ "$GCLOUD_PROJECT" == "(unset)" ]; then
    fail "GCLOUD_PROJECT is not set"
  fi
fi

if [ -z ${CLUSTER_NAME:+exists} ]; then
  CLUSTER_NAME="in-place-upgrade"
  export CLUSTER_NAME
fi

# Check that the GKE_VER variable has been set
if [ -z ${GKE_VER:+exists} ]; then
  fail "Set the GKE_VER environment variable"
fi

# Check that the NEW_GKE_VER variable has been set
if [ -z ${NEW_GKE_VER:+exists} ]; then
  fail "Set the NEW_GKE_VER environment variable"
fi

terraform_apply() {
  CONTROL_PLANE_VERSION=$1
  NODE_POOL_VERSION=$2
  terraform plan \
    -var control_plane_version="${CONTROL_PLANE_VERSION}" \
    -var node_pool_version="${NODE_POOL_VERSION}" \
    -var machine_type="${MACHINE_TYPE}" \
    -var num_nodes="${NUM_NODES}" \
    -var region="${GCLOUD_REGION}" \
    -var zone="${GCLOUD_ZONE}"

  terraform apply \
    -var control_plane_version="${CONTROL_PLANE_VERSION}" \
    -var node_pool_version="${NODE_POOL_VERSION}" \
    -var machine_type="${MACHINE_TYPE}" \
    -var num_nodes="${NUM_NODES}" \
    -var region="${GCLOUD_REGION}" \
    -var zone="${GCLOUD_ZONE}"
}

create_cluster() {
  echo "Building a GKE cluster using the following values:"
  echo "GCLOUD_ZONE = ${GCLOUD_ZONE}"
  echo "GCLOUD_REGION = ${GCLOUD_REGION}"
  echo "GCLOUD_PROJECT = ${GCLOUD_PROJECT}"
  echo "GKE Version = ${GKE_VER}"

  # Initialize terraform by downloading the appropriate provider
  terraform init
  terraform_apply "${GKE_VER}" "${GKE_VER}"

  # Acquire the kubectl credentials
  gcloud container clusters get-credentials "rolling-upgrade-test" \
    --region "${GCLOUD_REGION}" \
    --project "${GCLOUD_PROJECT}"

  # Deploy the example application
  kubectl apply -f "${REPO_HOME}/manifests/hello-server.yaml"
  kubectl apply -f "${REPO_HOME}/manifests/hello-svc.yaml"
}

upgrade_control() {
  terraform_apply "${NEW_GKE_VER}" "${GKE_VER}"
}

upgrade_nodes() {
  terraform_apply "${NEW_GKE_VER}" "${NEW_GKE_VER}"
}

downgrade_nodes() {
  terraform_apply "${NEW_GKE_VER}" "${GKE_VER}"
}

tear_down() {
  terraform destroy \
    -var control_plane_version="${NEW_GKE_VER}" \
    -var node_pool_version="${GKE_VER}" \
    -var machine_type="${MACHINE_TYPE}" \
    -var num_nodes="${NUM_NODES}" \
    -var region="${GCLOUD_REGION}" \
    -var zone="${GCLOUD_ZONE}"

}

auto() {
  create_cluster
  upgrade_control
  upgrade_nodes
  "${SCRIPT_HOME}/validate.sh"
}

ACTION=$1
case "${ACTION}" in
  auto)
    auto
    ;;
  create)
    create_cluster
    ;;
  upgrade-control)
    upgrade_control
    ;;
  upgrade-nodes)
    upgrade_nodes
    ;;
  downgrade-nodes)
    downgrade_nodes
    ;;
  delete)
    tear_down
    ;;
  *)
    usage
    ;;
esac
