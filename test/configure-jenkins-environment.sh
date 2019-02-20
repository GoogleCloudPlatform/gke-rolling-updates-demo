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

function test_versions {
  test -n "${toVersion}" || { echo >&2 'no toVersion found.. Aborting'; exit 1; }
  test -n "${fromVersion}" || { echo >&2 'no fromVersion found.. Aborting'; exit 1; }
  test "${toVersionShortName}" != "${fromVersionShortName}" || { echo >&2 'toVersion and fromVersion selected are equal.. Aborting'; exit 1; }
  to="$(echo $toVersionShortName | tr -d '.')"
  from="$(echo $fromVersionShortName | tr -d '.')"
  test "${to}" -gt "${from}" || { echo >&2 'toVersion is older than fromVersion.. Aborting'; exit 1; }
}

# All of the scripts expect to find ".env" in the root folder
cp env .env
# .env is used as a configuration file for the rest of the project.
# Need to choose some values for the automated tests in Jenkins
GCLOUD_REGION=us-west2

# get list of availbale master versions on the specified cloud region
master_versions=$(gcloud container get-server-config --zone "${GCLOUD_REGION}" 2>/dev/null | awk '/validNodeVersions:/ {f=0;next}; f; /validMasterVersions/ {f=1}' | awk '{print $2}')
# find two gke versions with different k8s versions
to_from=$(echo $master_versions | awk '{to_long=$1; split($1,a,"-"); to_short=a[1]; for(i=2;i<= NF;i++) { split($i,b,"-"); if(b[1] != $to_short) {print $to_long","$i; break} }}')

toVersion=${to_from%,*}
fromVersion=${to_from#*,}
toVersionShortName=$(echo "$toVersion" | cut -f1 -d'-')
fromVersionShortName=$(echo "$fromVersion" | cut -f1 -d'-')

# make sure versions are relevant
test_versions

echo ""
echo "Selected GKE version to migrate from: ${fromVersion}"
echo "Selected GKE version to migrate to -: ${toVersion}"
echo ""

sed -i "s/export K8S_VER=/export K8S_VER=${fromVersionShortName}/g" .env
sed -i "s/export NEW_K8S_VER=/export NEW_K8S_VER=${toVersionShortName}/g" .env
sed -i "s/export GKE_VER=/export GKE_VER=${fromVersion}/g" .env
sed -i "s/export NEW_GKE_VER=/export NEW_GKE_VER=${toVersion}/g" .env

