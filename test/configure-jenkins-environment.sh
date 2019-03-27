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

set -x

function test_versions() {
  test "${toVersionShortName}" != "${fromVersionShortName}" || {
    echo >&2 'toVersion and fromVersion selected are equal.. Aborting'
    exit 1
  }
  to="$(echo "$toVersionShortName" | sed s/-gke//g | tr -d '.')"
  from="$(echo "$fromVersionShortName" | sed s/-gke//g | tr -d '.')"
  test "${to}" -gt "${from}" || {
    echo >&2 'toVersion is older than fromVersion.. Aborting'
    exit 1
  }
}

# All of the scripts expect to find ".env" in the root folder
cp env .env

# get list of availbale master versions on the specified cloud region
master_versions="$(gcloud container get-server-config --region "${REGION}" --project "${PROJECT_ID}" --flatten validMasterVersions --format 'value(validMasterVersions)' 2>/dev/null)"

toVersionShortName="$(echo "${master_versions}" | head -n 1)"
fromVersionShortName="$(echo "${master_versions}" | head -n 2 | tail -n 1)"

# make sure versions are relevant
test_versions

echo ""
echo "Selected GKE version to migrate from: ${fromVersionShortName}"
echo "Selected GKE version to migrate to -: ${toVersionShortName}"
echo ""

sed -i "s/^export K8S_VER=/export K8S_VER=${fromVersionShortName}/g" .env
sed -i "s/^export NEW_K8S_VER=/export NEW_K8S_VER=${toVersionShortName}/g" .env
sed -i "s/^export GKE_VER=/export GKE_VER=${fromVersionShortName}/g" .env
sed -i "s/^export NEW_GKE_VER=/export NEW_GKE_VER=${toVersionShortName}/g" .env
