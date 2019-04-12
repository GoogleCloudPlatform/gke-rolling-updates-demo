#!/usr/bin/env bash
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

# validate.sh - a script to validate that an upgrade has completed successfully.
# It relies on the use of kubectl and jq.

# Stop immediately if something goes wrong
set -euo pipefail

# The absolute path to the root of the repository
REPO_HOME="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

fail() {
  echo "ERROR: ${*}"
  exit 2
}

# Validate that this workstation has access to the required executables
command -v kubectl >/dev/null || fail "kubectl is not installed!"
command -v jq >/dev/null || fail "jq is not installed!"

# Source the properties file
if [ -f "${REPO_HOME}/.env" ] ; then
  # shellcheck source=.env
  source "${REPO_HOME}/.env"
else
  echo "ERROR: Define a properties file '.env'"
  exit 1
fi

# Check that the NEW_GKE_VER variable has been set
if [ -z ${NEW_GKE_VER:+exists} ]; then
  echo "ERROR: Set the NEW_GKE_VER environment variable"
  exit 1
fi

# Validate the control plane version has been upgraded.  Returns:
# 0 - when the control plane version == NEW_GKE_VER
# 1 - when the control plane version != NEW_GKE_VER
validate_control() {
  CONTROL_VER=$(kubectl -n default version -o json | jq -r .serverVersion.gitVersion)
  # remove the preceeding v
  CONTROL_VER=${CONTROL_VER#v}
  if [[ "$CONTROL_VER" == "$NEW_GKE_VER" ]]; then
    return 0
  else
    echo "Control plane should be ${NEW_GKE_VER} but is ${CONTROL_VER}"
    return 1
  fi
}

# Validate that all nodes have been upgraded.
# Returns:
# 0 - when all node versions == NEW_GKE_VER
# 1 - when a node version != NEW_GKE_VER
validate_nodes() {
  NODES=$(kubectl -n default get nodes -o name)
  for NODE in ${NODES}; do
    # Find the kubelet version on each node.  This will match the gke version
    NODE_VER=$(kubectl -n default get "${NODE}" -o json | \
      jq -r '.status.nodeInfo.kubeletVersion')
    # remove the preceeding v, i.e. v1.10.4-gke.2 => 1.10.4-gke.2
    NODE_VER=${NODE_VER#v}
    if ! [[ "${NODE_VER}" == "${NEW_GKE_VER}" ]]; then
      echo -n "ERROR: ${NODE} has version ${NODE_VER}, "
      echo "but should have ${NEW_GKE_VER}"
      return 1
    fi
  done
  return 0
}

# Validate for correct number of hello-server pods
# Return:
# 0 - when all pods running
# 1 - when not all pods running
validate_pods() {
  # Find the current number of running pods
  PODS_AVAILABLE=$(kubectl -n default get deployment hello-server \
    -o jsonpath='{.status.availableReplicas}')
  PODS_REQUEST=$(kubectl -n default get deployments hello-server \
    -o jsonpath='{.status.replicas}')
  if ! [[ "${PODS_AVAILABLE}" == "${PODS_REQUEST}" ]]; then
    echo -n "ERROR: ${PODS_AVAILABLE} available pods, "
    echo "but should be ${PODS_REQUEST}"
    return 1
  fi
  return 0
}

# Validates that the upgrade was completed
validate() {
  echo "Validating the control plane version..."
  if validate_control ; then
    echo "Control plane is upgraded to ${NEW_GKE_VER}!"
  else
    exit 1
  fi
  echo "Validating the Nodes..."
  if validate_nodes ; then
    echo "All nodes upgraded to ${NEW_GKE_VER}!"
  else
    echo "ERROR: Not all nodes have been upgraded."
    exit 1
  fi
  echo "Validating the number of hello-server pods running..."
  if validate_pods ; then
    echo "All hello-server pods are running."
  else
    echo "ERROR: Not all pods available yet."
    exit 1
  fi
  return 0
}

# Time to validate
validate
