#!/bin/bash
# Copyright 2019 Google LLC
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

set -euo pipefail

SCRIPT_HOME="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_HOME="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

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
if [ -f "${REPO_HOME}/.env" ]; then
  # shellcheck source=.env
  source "${REPO_HOME}/.env"
fi

# Set GCLOUD_ZONE to default if it has not yet been set
GCLOUD_ZONE_DEFAULT=$(gcloud config get-value compute/zone)
if [ "${GCLOUD_ZONE_DEFAULT}" == "(unset)" ]; then
  # check if defined in env file
  if [ -z ${GCLOUD_ZONE:+exists} ]; then
    fail "GCLOUD_ZONE is not set"
  fi
else
  GCLOUD_ZONE="$GCLOUD_ZONE_DEFAULT"
  export GCLOUD_ZONE
fi

# Set GCLOUD_REGION to default if it has not yet been set
GCLOUD_REGION_DEFAULT=$(gcloud config get-value compute/region)
if [ "${GCLOUD_REGION_DEFAULT}" == "(unset)" ]; then
  # check if defined in env file
  if [ -z ${GCLOUD_REGION:+exists} ]; then
    fail "GCLOUD_REGION is not set"
  fi
else
  GCLOUD_REGION="$GCLOUD_REGION_DEFAULT"
  export GCLOUD_REGION
fi

# Set GCLOUD_PROJECT to default if it has not yet been set
GCLOUD_PROJECT_DEFAULT=$(gcloud config get-value project)
if [ "${GCLOUD_PROJECT_DEFAULT}" == "(unset)" ]; then
  # check if defined in env file
  if [ -z ${GCLOUD_PROJECT:+exists} ]; then
    fail "GCLOUD_PROJECT is not set"
  fi
else
  GCLOUD_PROJECT="$GCLOUD_PROJECT_DEFAULT"
  export GCLOUD_PROJECT
fi

if [ -z ${CLUSTER_NAME:+exists} ]; then
  CLUSTER_NAME="rolling-upgrade-test"
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

  terraform plan -input=false \
    -var control_plane_version="${CONTROL_PLANE_VERSION}" \
    -var node_pool_version="${NODE_POOL_VERSION}" \
    -var machine_type="${MACHINE_TYPE}" \
    -var num_nodes="${NUM_NODES}" \
    -var region="${GCLOUD_REGION}" \
    -var zone="${GCLOUD_ZONE}"

  terraform apply -input=false -auto-approve \
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
  cd "$SCRIPT_HOME"
  terraform init
  terraform_apply "${GKE_VER}" "${GKE_VER}"

  # Acquire the kubectl credentials
  gcloud container clusters get-credentials "rolling-upgrade-test" \
    --region "${GCLOUD_REGION}" \
    --project "${GCLOUD_PROJECT}"

  # Deploy the example application
  kubectl -n default apply -f "${REPO_HOME}/manifests/hello-server.yaml"
  kubectl -n default apply -f "${REPO_HOME}/manifests/hello-svc.yaml"
}

upgrade_control() {
  echo ""
  echo "Upgrading the K8s control plane ....."
  echo ""
  gcloud container clusters upgrade "${CLUSTER_NAME}" \
    --cluster-version="${NEW_K8S_VER}" \
    --region "${GCLOUD_REGION}" \
    --project "${GCLOUD_PROJECT}" \
    --master \
    --quiet
}

## updgrade the node clusters
upgrade_nodes() {
  echo ""
  echo "Upgrading the K8s nodes ....."
  echo ""
  gcloud container clusters upgrade "${CLUSTER_NAME}" \
    --cluster-version="${NEW_K8S_VER}" \
    --region "${GCLOUD_REGION}" \
    --project "${GCLOUD_PROJECT}" \
    --quiet
}

downgrade_nodes() {
  cd "$SCRIPT_HOME"
  terraform_apply "${NEW_GKE_VER}" "${GKE_VER}"
}

tear_down() {
  cd "$SCRIPT_HOME"
  terraform init
  terraform destroy \
    -auto-approve \
    -var control_plane_version="${NEW_GKE_VER}" \
    -var node_pool_version="${GKE_VER}" \
    -var machine_type="${MACHINE_TYPE}" \
    -var num_nodes="${NUM_NODES}" \
    -var region="${GCLOUD_REGION}" \
    -var zone="${GCLOUD_ZONE}" \
    -var timeout_create="${TIMEOUT_CREATE}" \
    -var timeout_update="${TIMEOUT_UPDATE}" \
    -var timeout_delete="${TIMEOUT_DELETE}"
}

# After the master is upgraded, the control plane instances get upgraded
# and all other cluster operations will fail until the upgrade has completed.
wait_for_upgrade() {
  echo "Checking for master upgrade"
  OP_ID=$(gcloud container operations list \
    --project "${GCLOUD_PROJECT}" \
    --region "${GCLOUD_REGION}" \
    --filter 'TYPE=UPGRADE_MASTER' \
    --filter 'STATUS=RUNNING' \
    --format 'value(name)' |
    head -n1)
  if [[ "${OP_ID}" =~ ^operation-.* ]]; then
    echo "Master upgrade in process.  Waiting until complete..."
    gcloud container operations wait "${OP_ID}" \
      --region "${GCLOUD_REGION}" \
      --project "${GCLOUD_PROJECT}"
  fi
}

auto() {
  create_cluster
  upgrade_control
  wait_for_upgrade
  upgrade_nodes
  wait_for_upgrade
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
