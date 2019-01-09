#!/usr/bin/env bash

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

set -euo pipefail

# "---------------------------------------------------------"
# "-                                                       -"
# "-         rolling updates expand contract               -"
# "-                                                       -"
# "-     this poc demonstrates the use of the expand       -"
# "-     and contract pattern for upgrading gke clusters,  -"
# "-     the pattern works by increasing the node pool     -"
# "-     size prior to the upgrade to provide additional   -"
# "-     headroom while upgrading, once the upgrade is     -"
# "-     complete the node pool is restored to its         -"
# "-     original size                                     -"
# "-                                                       -"
# "---------------------------------------------------------"



## source properties file
SCRIPT_HOME="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_HOME="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
# shellcheck source=.env
source "${REPO_HOME}/.env"

if [ -z ${CLUSTER_NAME:+exists} ]; then
  CLUSTER_NAME="expand-contract-upgrade"
  export CLUSTER_NAME
fi

################  functions  ####################


## validate use of this script
usage() {
  echo ""
  echo " Checking valid paramater passed to script ....."
  echo ""
  cat <<-EOM
USAGE: $(basename "$0") <action> [N]
Where the <action> can be:
  auto
  create
  upgrade-control
  upgrade-nodes
  resize <N>
  delete
N - The Number of nodes per zone to set the default node pool during resize
EOM
  exit 1
}

## check dependencies installed
check_dependencies() {
  echo ""
  echo "Checking dependencies are installed ....."
  echo ""
  command -v gcloud >/dev/null 2>&1 || { \
    echo >&2 "I require gcloud but it's not installed.  Aborting."; exit 1; }
  command -v kubectl >/dev/null 2>&1 || { \
    echo >&2 "I require kubectl but it's not installed.  Aborting."; exit 1; }
}


## check project exists
check_project() {
  echo ""
  echo "Checking the project specified for the demo exists ....."
  echo ""
  local EXISTS
  EXISTS=$(gcloud projects list | awk "/${GCLOUD_PROJECT} /" | awk '{print $1}')
  sleep 1
  if [[ "${EXISTS}" != "${GCLOUD_PROJECT}" ]] ; then
    echo ""
    echo "the ${GCLOUD_PROJECT} project does not exists"
    echo "please update properties file with "
    echo "a valid project"
    echo ""
    exit 1
  fi
}


## check api's enabled
check_apis() {
  echo ""
  echo "Checking the appropriate API's are enabled ....."
  echo ""
  COMPUTE_API=$(gcloud services list --project="${GCLOUD_PROJECT}" \
            --format='value(serviceConfig.name)' \
            --filter='serviceConfig.name:compute.googleapis.com' 2>&1)
  if [[ "${COMPUTE_API}" != "compute.googleapis.com" ]]; then
    echo "Enabling the Compute Engine API"
    gcloud services enable compute.googleapis.com --project="${GCLOUD_PROJECT}"
  fi
  CONTAINER_API=$(gcloud services list --project="${GCLOUD_PROJECT}" \
            --format='value(serviceConfig.name)' \
            --filter='serviceConfig.name:container.googleapis.com' 2>&1)
  if [[ "${CONTAINER_API}" != "container.googleapis.com" ]]; then
    echo "Enabling the Kubernetes Engine API"
    gcloud services enable container.googleapis.com --project="${GCLOUD_PROJECT}"
  fi
}

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

## create cluster
create_cluster() {
  # create cluster
  echo ""
  echo "Building a GKE cluster ....."
  echo ""
  gcloud container clusters create "${CLUSTER_NAME}" \
      --machine-type "${MACHINE_TYPE}" \
      --num-nodes "${NUM_NODES}" \
      --cluster-version "${K8S_VER}" \
      --project "${GCLOUD_PROJECT}" \
      --region "${GCLOUD_REGION}"
  # acquire the kubectl credentials
  gcloud container clusters get-credentials "${CLUSTER_NAME}" \
    --region "${GCLOUD_REGION}" \
    --project "${GCLOUD_PROJECT}"
}


## Creates the Shakespeare index and loads the data
load_data() {
  echo "Setting up port-forward to Elasticsearch client"
  kubectl -n default port-forward svc/elasticsearch 9200 1>&2>/dev/null &
  # Wait a couple seconds for connection to establish as that last command is
  # not blocking
  sleep 5

  echo "Creating the Shakespeare index"
  # The mapping file creates the index and sets the metadata needed by
  # Elasticsearch to parse the actual data
  curl -H "Content-Type: application/json" \
    -X PUT \
    -d @"${REPO_HOME}/data/mapping.json" \
    'http://localhost:9200/shakespeare'
  # The response does not make include a newline
  echo ""

  # Here we load the actual data.
  echo "Loading Shakespeare sample data into Elasticsearch"
  curl -H "Content-Type: application/x-ndjson" \
    -X POST \
    --data-binary @"${REPO_HOME}/data/shakespeare.json" \
    'http://localhost:9200/shakespeare/doc/_bulk?pretty' > /dev/null

  # If we've made it this far the data is loaded
  echo "Sample data successfully loaded!"

  pkill -P $$
}

## Installs the Elasticsearch cluster
setup_app() {
  echo "Installing Elasticsearch Cluster"
  kubectl -n default create -f "${REPO_HOME}/manifests/es-discovery-svc.yaml"
  kubectl -n default create -f "${REPO_HOME}/manifests/es-svc.yaml"
  kubectl -n default create -f "${REPO_HOME}/manifests/es-master-pdb.yaml"
  kubectl -n default create -f "${REPO_HOME}/manifests/es-master.yaml"
  kubectl -n default rollout status -f "${REPO_HOME}/manifests/es-master.yaml"

  kubectl -n default create -f "${REPO_HOME}/manifests/es-client-pdb.yaml"
  kubectl -n default create -f "${REPO_HOME}/manifests/es-client.yaml"
  kubectl -n default rollout status -f "${REPO_HOME}/manifests/es-client.yaml"

  kubectl -n default create -f "${REPO_HOME}/manifests/es-data-svc.yaml"
  kubectl -n default create -f "${REPO_HOME}/manifests/es-data-pdb.yaml"
  kubectl -n default create -f "${REPO_HOME}/manifests/es-data-stateful.yaml"
  kubectl -n default rollout status -f "${REPO_HOME}/manifests/es-data-stateful.yaml"

  load_data
}

# uninstall app
uninstall_app() {
 gcloud container clusters get-credentials "${CLUSTER_NAME}" \
    --region "${GCLOUD_REGION}" \
    --project "${GCLOUD_PROJECT}"
  echo "Uninstalling Elasticsearch Cluster"
  kubectl -n default delete -f "${REPO_HOME}"/manifests/ || true

  # You have to wait the default pod grace period before you can delete the pvcs
  GRACE=$(kubectl --namespace default get sts -l component=elasticsearch,role=data -o jsonpath='{..terminationGracePeriodSeconds}')
  PADDING=30
  echo "Sleeping $(( GRACE + PADDING )) seconds before deleting PVCs. The default pod grace period."
  sleep "$(( GRACE + PADDING ))"

  # Deleting and/or scaling a StatefulSet down will not delete the volumes associated with the StatefulSet.
  # This is done to ensure data safety, which is generally more valuable
  # than an automatic purge of all related StatefulSet resources.
  echo "Delete PVCs..."
  kubectl -n default delete pvc -l component=elasticsearch,role=data || true
  echo "Delete PVs..."
  kubectl delete pv $(kubectl get pv --all-namespaces | grep es-data | awk '{ print $1}')
  echo "kubectl get pvc --all-namespaces"
  kubectl -n default get pvc --all-namespaces
  echo "kubectl get pv --all-namespaces"
  kubectl -n default get pv --all-namespaces
  echo "Debug message"
}


## increase size of the node pool
resize_node_pool() {
  local SIZE=$1
  echo ""
  echo "Resizing the node pool to $SIZE nodes ....."
  echo ""
  gcloud container clusters resize "${CLUSTER_NAME}" \
    --size "${SIZE}" \
    --region "${GCLOUD_REGION}" \
    --project "${GCLOUD_PROJECT}" \
    --quiet
}


## upgrade the control plane
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


## tear down the demo
tear_down() {
  echo ""
  echo "Tearing down the infrastructure ....."
  echo ""
  uninstall_app
  delete_cluster
}

# delete cluster
delete_cluster() {
echo "Delete Cluster"
  if gcloud container clusters describe "${CLUSTER_NAME}" \
    --project "${GCLOUD_PROJECT}" \
    --region "${GCLOUD_REGION}"; then
echo "here 1"
gcloud container clusters list --filter="STATUS:RUNNING AND NAME:$CLUSTER_NAME"
  # Cluster might be still upgrading. Wait up to 5 mins and then delete it
echo "here 2"
  COUNTER=0
  until [ $(gcloud container clusters list --filter="STATUS:RUNNING AND NAME:$CLUSTER_NAME" | wc -l) -ne 0 -o $COUNTER -ge 5 ]; do
    echo Waiting for cluster upgrade to finish...
    sleep 60
    COUNTER=$[$COUNTER+1]
  done
echo "here 3"
  gcloud container clusters delete $"${CLUSTER_NAME}" \
    --project "${GCLOUD_PROJECT}" \
    --region "${GCLOUD_REGION}" \
    --quiet
  fi
}

# After the node pool is expanded, the control plane instances will likely be
# vertically scaled automatically by Kubernetes Engine to handle the increased
# load of more instances.  When the control plane is upgrading, no other cluster
# modifications can occur.
wait_for_upgrade() {
  echo "Checking for master upgrade"
  OP_ID=$(gcloud container operations list \
    --project "${GCLOUD_PROJECT}" \
    --region "${GCLOUD_REGION}" \
    --filter 'TYPE=UPGRADE_MASTER' \
    --filter 'STATUS=RUNNING' \
    --format 'value(name)' \
    | head -n1 )
  if [[ "${OP_ID}" =~ ^operation-.* ]]; then
    echo "Master upgrade in process.  Waiting until complete..."
    gcloud container operations wait "${OP_ID}" \
      --region "${GCLOUD_REGION}"
  fi
}

auto() {
  create_cluster
  setup_app
  resize_node_pool 2
  # Unfortunate race condition here, a little sleep should be enough
  sleep 10
  wait_for_upgrade
  upgrade_control
  wait_for_upgrade
  upgrade_nodes
  resize_node_pool 1
  "${SCRIPT_HOME}/validate.sh"
}

################  execution  ####################

# validate script called correctly
if [[ $# -lt 1 ]]; then
  usage
fi

# check dependencies installed
check_dependencies

# check project exist
check_project

# check apis enabled
check_apis

ACTION=$1
case "${ACTION}" in
  auto)
    auto
    ;;
  create)
    create_cluster
    setup_app
    ;;
  resize)
    resize_node_pool "$2"
    ;;
  upgrade-control)
    upgrade_control
    ;;
  upgrade-nodes)
    upgrade_nodes
    ;;
  delete)
    tear_down
    ;;
  *)
    usage
    ;;
esac
