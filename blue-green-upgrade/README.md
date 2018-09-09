# Kubernetes Engine blue/green Rolling Update

## Table of Contents

<!--ts-->

* [Introduction](#introduction)
* [Architecture](#architecture)
* [Prerequisites](#prerequisites)
   * [Run Demo in a Google Cloud Shell](#run-demo-in-a-google-cloud-shell)
   * [Supported Operating Systems](#supported-operating-systems)
   * [Tools](#tools)
   * [Configuration](#configuration)
       * [.env Properties](#env-properties)
       * [Selecting your versions](#selecting-your-versions)
* [Deployment](#deployment)
   * [Manual Deployment](#manual-deployment)
   * [Automated Deployment](#automated-deployment)
* [Validation](#validation)
* [Tear Down](#tear-down)
* [Troubleshooting](#troubleshooting)
* [Relevant Material](#relevant-material)

<!--te-->

## Introduction

This code repository demonstrates a Kubernetes Engine cluster upgrade using the
blue/green, or 'lift and shift', upgrade strategy.  This upgrade strategy is a
great choice for clusters containing mission-critical stateful apps that require
extra care and attention during upgrades and migrations.

Some workloads may have specific concerns that can not be accounted for with
`readinessProbes` and `PodDisruptionBudgets` alone.  In these cases, the
blue/green approach will give you the necessary control to upgrade the cluster
with minimal disruption to the applications running in the cluster.

## Architecture

In a blue/green upgrade, a duplicate node pool of equal size is created with the
new Kubernetes Engine version.  The node pools with old and new Kubernetes
Engine versions are run simultaneously.  This allows individual pods or entire
nodes to be migrated to the new Kubernetes Engine version one at a time as the
operator sees fit.

This example will walk through creating a Kubernetes Engine cluster, deploying
an Elasticsearch cluster, loading an index containing the works of Shakespeare,
upgrading the Kubernetes Engine Control Plane, creating the new node pool,
migrating the application to the new node pool, and terminating the old node
pool.

To complete this example, you will run `cluster_ops.sh` contained in
this directory. It uses the `gcloud` and `kubectl` commands to interact with
the Google Cloud Platform and the Kubernetes Engine cluster.

It has been noted by many in the Kubernetes community that running stateful
applications on Kubernetes is not for beginners.  A familiarity with both the
application and Kubernetes are a must to do so successfully.

There are two possibilities when you run a stateful datastore on Kubernetes:

1. You are a very experienced K8s user and know exactly what is around the corner.
2. You have no idea what is around the corner, and you are going to learn very fast.

## Prerequisites

A Google Cloud account and project is required for this.  Access to an existing Google Cloud
project with the Kubernetes Engine service enabled If you do not have a Google Cloud account
please signup for a free trial [here](https://cloud.google.com).

### Run Demo in a Google Cloud Shell

Click the button below to run the demo in a [Google Cloud Shell](https://cloud.google.com/shell/docs/).

[![Open in Cloud Shell](http://gstatic.com/cloudssh/images/open-btn.svg)](https://console.cloud.google.com/cloudshell/open?git_repo=https%3A%2F%2Fgithub.com%2FGoogleCloudPlatform%2Fgke-rolling-updates-demo&page=editor&tutorial=blue-green-upgrade%2FREADME.md)

All the tools for the demo are installed. When using Cloud Shell execute the following
command in order to setup gcloud cli.

```console
gcloud init
```

### Supported Operating Systems

This project will run on macOS, Linux, or in a [Google Cloud Shell](https://cloud.google.com/shell/docs/).

### Tools

When not using Cloud Shell, the following tools are required.

1.  [gcloud](https://cloud.google.com/sdk/downloads) (Google Cloud SDK version >= 200.0.0)
1.  [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) >= 1.10.4
1.  bash or bash compatible shell
1.  [watch](https://en.wikipedia.org/wiki/Watch_(Unix)
1.  [jq](https://stedolan.github.io/jq/)
1.  A Google Cloud Platform project with the Kubernetes Engine API enabled.
    ```
    gcloud services enable container.googleapis.com
    ```

### Configuration

##### `.env` Properties

A number of environment variables must be set to run the `cluster_ops.sh`
script.  The required variables are provided and explained in the `env` file
found in the root of this repository. Make a copy in the root of this repository:
```
cp env .env
```
Update the `.env` file with appropriate values for your use.  It will be
sourced by the `cluster_ops.sh` script each time it is run.

##### Selecting your versions

In the `.env` file, you must select two Kubernetes versions, `K8S_VER` and
`NEW_K8S_VER`, supply only the open source Kubernetes semver version number and
Kubernetes Engine will select the appropriate Kubernetes Engine patch version
when creating and upgrading the cluster.  This example was tested using the
following versions:
```
K8S_VER=1.9.7
NEW_K8S_VER=1.10.4
```

## Deployment

### Manual Deployment

You can run `cluster_ops.sh` from anywhere in your file system but if you copy
paste these commands exactly, you should first cd into the directory containing
the script. The validation section describes commands to monitor the status of
the cluster and application during the upgrade procedure.

1.  **Create the Kubernetes Engine cluster:**
    The `create` action will create a regional Kubernetes Engine Cluster and
    deploy the example application.

    ```console
    ./cluster_ops.sh create
    ```

    You will be prompted to continue, input `Y`.  After a few minutes the
    Kubernetes Engine cluster will be created, the Elasticearch cluster will be
    installed, and an index containing the works of Shakespeare will loaded. The
    last several lines of output will look like this:

    ```console
    Creating the Shakespeare index
    {"acknowledged":true,"shards_acknowledged":true,"index":"shakespeare"}
    Loading Shakespeare sample data into Elasticsearch
      % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                     Dload  Upload   Total   Spent    Left  Speed
    100 62.6M  100 38.5M  100 24.1M  1642k  1029k  0:00:24  0:00:24 --:--:-- 3719k
    Sample data successfully loaded!
    ```

1.  **Upgrade the control plane:**
    ```console
    ./cluster_ops.sh upgrade-control
    ```
    You will be prompted to proceed, enter `Y`.  After several minutes, the
    upgrade will be complete and the output should look similar to this:
    ```console
    Upgrading control plane to version 1.10.2
    Master of cluster [blue-green-test] will be upgraded from version
    [1.9.7-gke.3] to version [1.10.2]. This operation is long-running and
    will block other operations on the cluster (including delete) until it
     has run to completion.

    Do you want to continue (Y/n)?  Y

    Upgrading blue-green-test...done.
    Updated [https://container.googleapis.com/v1/projects/my-test-project/zones/us-east1/clusters/blue-green-test].
    ```

1.  **Create the new node pool:**
    Now that the control plane is upgraded to the new version, we can create a
    node pool running the new Kubernetes version.
    ```console
    ./cluster_ops.sh new-node-pool
    ```
    This command will also cordon the nodes in the default node pool.  Once a
    node is cordoned, the Kubernetes scheduler will no longer schedule new pods
    on that node. Existing pods on a cordoned node are not automatically moved.

    ```console
    Creating node pool new-pool...done.
    Created [https://container.googleapis.com/v1/projects/my-test-project/zones/us-east1/clusters/blue-green-test/nodePools/new-pool].
    NAME      MACHINE_TYPE   DISK_SIZE_GB  NODE_VERSION
    new-pool  n1-standard-4  100           1.10.2-gke.3
    Cordoning nodes in old node pool
    node "gke-blue-green-test-default-pool-1265945e-6bl1" cordoned
    node "gke-blue-green-test-default-pool-509edc38-vll6" cordoned
    node "gke-blue-green-test-default-pool-bbe63a14-wq08" cordoned
    ```

1.  **Migrate the workloads:** You can now migrate your applications as slow or
    fast as you would like.  For stateful applications that have consensus
    requirements, sharded data, or replication concerns, you may want to migrate
    a single pod at a time and monitor the application's health before
    introducing more disruptions.  Once all stateful applications have been
    migrated, you can migrate the remaining workloads one node at a time with
    the `drain` command.

    * Migrate a single pod
      ```console
      kubectl delete pod <pod-name>
      ```
    * Migrate an entire node:
      ```console
      kubectl drain <node-name> --delete-local-data --ignore-daemonsets [--force]
      ```

    ###### Migrating Elasticsearch Master Nodes

    The Elasticsearch cluster has one Master node and two "Master Eligible"
    nodes.  When the Master is deleted, the remaining nodes will re-elect a new
    master.  The new master will then update the cluster state and publish the
    new state to all members of the cluster.  During this time period (40-60s)
    all cluster level API calls and many index metadata API calls like the ones
    below will fail with a timeout:

    ```console
    /_cluster/health
    /_cat/master
    /_cat/indices
    /_cat/shards
    ```

    Search API queries should continue without interruption:

    ```console
    /shakespeare/_search?q=happy%20dagger
    ```

    To minimize the number of Master re-elections, determine the current master
    and migrate the 2 Master Eligible nodes first:

    First, set up a port-forward between the Elasticsearch client service and
    your workstation's localhost:

    ```console
    kubectl port-forward svc/elasticsearch 9200
    ```

    This API call will display the current master:

    ```console
    curl localhost:9200/_cat/master
    ```

    The current Master pod name is the 4th column.

    ```console
    gxcdTgBpRoejJGZSZYI7kA 10.12.2.3 10.12.2.3 es-master-5bf75c4d7b-rd67l
    ```

    Find the other two Master Eligible nodes:

    ```console
    kubectl get pods -l component=elasticsearch,role=master
    ```

    The Master and Master Eligible nodes are displayed:

    ```console
    NAME                         READY     STATUS    RESTARTS   AGE
    es-master-5bf75c4d7b-2pnlh   1/1       Running   0          10m
    es-master-5bf75c4d7b-7mkcf   1/1       Running   0          10m
    es-master-5bf75c4d7b-rd67l   1/1       Running   0          10m
    ```

    Delete one of the Master Eligible nodes

    ```console
    kubectl delete pod es-master-5bf75c4d7b-2pnlh
    ```
    Watch the cluster health in a loop (see the **Application Health** heading in
    the [Troubleshooting](#troubleshooting) section below) and wait for the new Master Eligible
    node to join the cluster.  Once the cluster state has returned to normal,
    delete the other Master Eligible Node pod.  Again, wait for the new Master
    Eligible Node to join the cluster.  Finally, delete the Master.  Confirm
    that search queries continue to work while the Master re-election occurs.

    ###### Migrating Elasticsearch Data nodes

    The data nodes can be migrated in any order but care must be taken to ensure
    that the data is available throughout the migration process.  The
    Shakespeare index is split into 5 `primary shards`, and 1 `replica shard` per
    `primary shard`. This gives a total of `5 x 2 = 10` shards that are spread
    out among the data nodes.  Elasticsearch ensures that a `primary shard's`
    corresponding `replica shard` will not be located on the same data node
    whenever possible.

    Delete the data nodes one at a time while watching the cluster health in a
    loop.  After deleting a data node, the cluster `status` will change to
    `yellow`.  This means that all `primary shards` are active but not all
    `replica shards` are available.  After the new data node is created, it will
    take some time further for the cluster to ensure all shards are properly
    allocated across the data nodes and return to the `green` `status`.

    ###### Migrating the rest of the pods

    Now that the Elasticsearch Master and Data nodes have been migrated to the
    new node pool, the rest of the workloads are stateless and can be quickly
    migrated one node-pool node at a time.
    Migrate the rest of the nodes in the default pool:
    ```console
    ./cluster_ops.sh drain-default-pool
    ```


1.  **Delete the old node pool:** Now that all workloads have been migrated to
    the new node pool, it is time to delete the old node pool.  Perform one
    final check to ensure all the necessary pods have been migrated:
    ```console
    kubectl get pods --all-namespaces -o wide
    ```
    With the `-o wide` flag you can see on which node each pod is scheduled.
    The only pods left on the default node pool should be `kube-proxy`,
    `fluentd`, and other daemonset pods.  Once you have confirmed that all
    nodes are drained, proceed with deleting the default node pool:
    ```console
    ./cluster_ops.sh delete-default-pool
    ```
    You will be prompted to enter `Y` to proceed.  The output will be similar
    to this:
    ```console
    Deleting the default node pool
    The following node pool will be deleted.
    [default-pool] in cluster [blue-green-test] in [us-east1]

    Do you want to continue (Y/n)?  Y

    Deleting node pool default-pool...done.
    Deleted [https://container.googleapis.com/v1/projects/my-test-project/zones/us-east1/clusters/blue-green-test/nodePools/default-pool].
    ```
    If you receive an error because the cluster `is currently upgrading`, check
    the [Troubleshooting](#troubleshooting) section below.

### Automated Deployment

The cluster creation, upgrade, and validation can be run with one command:
```console
./cluster_ops.sh auto
```


## Validation

*   **Control Plane Upgrade:** While the control plane is upgrading, you can
    verify that Regional Kubernetes Engine clusters have an HA control plane by
    querying the API server in a loop:
    ```console
    watch kubectl get pods --all-namespaces
    ```
    As each control plane node is replaced, other running control plane nodes
    will serve requests to the `kubectl` commands providing a zero-downtime
    upgrade.
    When the control plane upgrade is complete, you can see the new `Server
    Version` with
    ```
    kubectl version --short
    ```
*   **`gcloud` monitoring:** You can monitor the progress of cluster
    upgrades with the `glcoud` command.  Version upgrades, node pool additions
    and removals are referred to as "Cluster Operations".  Both completed and
    in-progress operations are logged by Kubernetes Engine and can be inspected.
    Find the appropriate `OPERATION_ID` in the `NAME` column by listing all
    cluster operations in the region where the cluster was created.
    ```console
    gcloud container operations list --region <GLCOUD_REGION>
    ```
    Copy the appropriate `OPERATION_ID` and use it to query Kubernetes Engine
    for details about the current cluster operation.
    ```console
    gcloud container operations describe <OPERATION_ID> \
      --region <GCLOUD_REGION>
    ```
    ** Cloud console monitoring:** You can monitor the progress of cluster upgrades
    using GCP console under Kubernetes Engine, select your cluster and monitor
    the process/progress in %.

*   **Default node pool cordon:** After the control plane is upgraded, you can
    verify that the default node pool has been cordoned:
    ```console
    kubectl get nodes
    ```
    For each node in the default pool, the node status has changed from `Ready`
    to `Ready,SchedulingDisabled`:
    ```console
    NAME                                             STATUS                     ROLES     AGE       VERSION
    gke-blue-green-test-default-pool-6fab6061-6zk5   Ready,SchedulingDisabled   <none>    40m       v1.9.6-gke.1
    gke-blue-green-test-default-pool-cab59d39-0c3k   Ready,SchedulingDisabled   <none>    40m       v1.9.6-gke.1
    gke-blue-green-test-default-pool-ded2c6b1-1qwr   Ready,SchedulingDisabled   <none>    40m       v1.9.6-gke.1
    gke-blue-green-test-new-pool-3d0a2cb6-629s       Ready                      <none>    3m        v1.10.2-gke.3
    gke-blue-green-test-new-pool-4ea2cc03-tvcx       Ready                      <none>    3m        v1.10.2-gke.3
    gke-blue-green-test-new-pool-cf15ea8e-8457       Ready                      <none>    3m        v1.10.2-gke.3
    ```
*   **Rescheduling:** As pods are deleted and nodes are drained, you can view
    the progress of rescheduling:
    ```console
    kubectl get pods --all-namespaces
    ```

*   **Application Health:** Throughout all upgrade steps, an HA application
    with appropriate number of pods should continue running uninterrupted.  The
    Elasticsearch cluster in this example will continue serving search queries
    as long as the cluster health is `green` or `yellow`.  It has 3 Data Nodes,
    3 Client Nodes, and 3 Master Eligible Nodes with one elected Master.

    If you have not already, set up a port-forward to the Elasticsearch client
    service to your workstation's localhost:

    ```console
    kubectl port-forward svc/elasticsearch 9200
    ```

    Then in another terminal check the cluster health in a loop:

    ```console
    while true; do \
        date "+%H:%M:%S,%3N" \
        curl --max-time 1 'http://localhost:9200/_cluster/health' | jq .
        echo "" \
        sleep 1 \
    done
    ```

    A healthy cluster with all nodes available will look like this:

    ```console
    {
      "cluster_name": "myesdb",
      "status": "green",
      "timed_out": false,
      "number_of_nodes": 9,
      "number_of_data_nodes": 3,
      "active_primary_shards": 5,
      "active_shards": 10,
      "relocating_shards": 0,
      "initializing_shards": 0,
      "unassigned_shards": 0,
      "delayed_unassigned_shards": 0,
      "number_of_pending_tasks": 0,
      "number_of_in_flight_fetch": 0,
      "task_max_waiting_in_queue_millis": 0,
      "active_shards_percent_as_number": 100
    }
    ```

    In yet another terminal window, you can run a loop to test the availability
    of the search API which should continue working during a Master re-election:
    ```console
    while true; do \
        date "+%H:%M:%S,%3N" \
        curl --max-time 1 'http://localhost:9200/shakespeare/_search?q=happy%20dagger'
        echo "" \
        sleep 1 \
    done
    ```

*   **Completed Upgrade:** After the upgrade steps have been completed, the
    `validation.sh` script will check the control plane version and each
    node's version.  Execute it from within this directory:
    ```console
    ./validation.sh
    ```
    Successful output will look like this:
    ```console
    Validating the control plane version...
    Control plane is upgraded to 1.10.4-gke.2!
    Validating the Nodes...
    All nodes upgraded to 1.10.4-gke.2!
    Validating the number of hello-server pods running...
    All hello-server pods have been running.
    ```


## Tear Down

To delete the Kubernetes Engine cluster and all other resources generated
during this example run the following command:
```console
./cluster_ops.sh delete
```

## Troubleshooting

* `E0717 09:45:59.417020    1245 portforward.go:178] lost connection to pod`
  The port-forward command will occasionally fail, especially as the cluster is
  being manipulated.  Execute the command to reconnect.

* `Currently upgrading cluster` Error:
  ```
  ERROR: (gcloud.container.node-pools.delete) ResponseError: code=400, message=Operation operation-1529415957904-496c7278 is currently upgrading cluster blue-green-test. Please wait and try again once it is done.
  ```
  Because the Kubernetes Engine control plane is a managed service, there are
  times when it will be upgraded for you.  During these times, many cluster
  operations like upgrading versions, adding or removing node pools, are
  temporarily blocked.  These automatic upgrades do not change the version of
  the control plane.  They can be triggered by:
  * An increase in the number of nodes - the control plane will be vertically
    scaled to handle the increased API server load.
  * When a node pool is added or removed - the control plane will be upgraded
    to account for the new configuration.

  You can monitor the progress of any cluster operation:
  ```console
  gcloud container operations list [--region <GCLOUD_REGION>]
  gcloud container operations describe <operation-id> [--region <GCLOUD_REGION>]
  ```

* `IN_USE_ADDRESSES` Quota Error:
  ```
  ERROR: (gcloud.container.clusters.create) ResponseError: code=403, message=Insufficient regional quota to satisfy request for resource: "IN_USE_ADDRESSES". The request requires '9.0' and is short '1.0'. The regional quota is '8.0' with '8.0' available.
  ```
  1.  Open the GCP Console and navigate to `IAM & admin` -> `Quotas`.
  1.  Filter the quotas by selecting your region under `Location`.
  1.  Check the box next to `Compute Engine API In-use IP addresses global`,
      then click `EDIT QUOTAS`.
  1.  Follow the steps to increase the quota.  Quotas are not immediately
      increased.
* `CPUS` Quota Error:
  ```
  ERROR: (gcloud.container.node-pools.create) ResponseError: code=403, message=Insufficient regional quota to satisfy request for resource: "CPUS". The request requires '12.0' and is short '3.0'. The regional quota is '24.0' with '9.0' available.
  ```
  1.  Open the GCP Console and navigate to `IAM & admin` -> `Quotas`.
  1.  Filter the quotas by selecting your region under `Location`.
  1.  Check the box next to `Compute Engine API CPUs`, then click `EDIT QUOTAS`.
  1.  Follow the steps to increase the quota.  Quotas are not immediately
      increased.

## Relevant Material

* `PodDisruptionBudgets` - [Kubernetes Disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/)
* `readinessProbe` - [Pod lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)
* [Kubernetes Engine Release Notes](https://cloud.google.com/kubernetes-engine/release-notes)
* [Migrating a Node Pool](https://cloud.google.com/kubernetes-engine/docs/tutorials/migrating-node-pool)
