# Kubernetes Engine Expand And Contract Update

## Table of Contents

<!--ts-->

* [Introduction](#introduction)
* [Architecture](#architecture)
* [Deployment](#deployment)
   * [Manual Deployment](#manual-deployment)
   * [Automated Deployment](#automated-deployment)
* [Validation](#validation)
* [Tear Down](#tear-down)
* [Troubleshooting](#troubleshooting)
* [Relevant Material](#relevant-material)

<!--te-->

## Introduction

This demo illustrates how to use the 'expand and contract' pattern to upgrade
a Kubernetes Engine cluster. The pattern is designed to avoid issues with
resource availability in the course of a Kubernetes Engine upgrade.

The amount of available resources is often called Headroom.  Evaluating a
cluster's headroom requires looking at two related aspects:
1.  **Cluster Headroom:** The sum of all unused resources across all nodes.
1.  **Node Headroom:** The amount of unused resources on a specific node.
It is possible for a cluster to have sufficient Cluster Headroom to handle
losing a single node while not having enough Node Headroom to reschedule every
pod from the lost node.  This is especially true with StatefulSet pods that
and only attach Google Cloud Disks in a specific Availability Zone.

The Expand and Contract Upgrade pattern increases both Node Headroom and Cluster
Headroom by adding 1 or more new nodes to the node pool prior to starting the
upgrade.  Once the upgrade has completed, the extra nodes are removed.

## Architecture

In this demo, we walk through the steps to correctly perform an expand and
contract upgrade.

1.  We will create a regional Kubernetes Engine cluster and deploy an
    application to it
1.  We will increase the node pool size of the cluster
1.  We will upgrade the Kubernetes Engine control plane and the Kubernetes
    Engine node pool separately
1.  We will monitor pod activity during the upgrade process
1.  We will resize the cluster back down to its original size upon successful
    completion of the Kubernetes Engine upgrade.

To complete this example, you will run `cluster_ops.sh` contained in
this repository. It uses `gcloud` and `kubectl` commands to interact with
the Google Cloud Platform and the Kubernetes Engine cluster.



## Deployment

### Manual Deployment

Run each command below from the root of this repository. The validation
section describes commands to monitor the status of the cluster and application
during the upgrade procedure.

Note, every time the `cluster_ops.sh` script is run, it will always check the
the following items: dependencies are installed, the project specified in the
properties file exists, and that the appropriate api's have been enabled

```console

Checking dependencies are installed .....

Checking the project specified for the demo exists .....

Checking the appropriate api's are enabled .....
```

1.  **Create the Kubernetes Engine cluster:**
    The `create` action will create a regional Kubernetes Engine Cluster and
    deploy the example application.

    ```console
    ./cluster_ops.sh create
    ```

    After a few minutes the Kubernetes Engine cluster will be created, the
    Elasticearch cluster will be installed, and an index containing the works
    of Shakespeare will loaded. The last several lines of output will look like
    this:

    ```console
    Creating the Shakespeare index
    {"acknowledged":true,"shards_acknowledged":true,"index":"shakespeare"}
    Loading Shakespeare sample data into Elasticsearch
      % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                     Dload  Upload   Total   Spent    Left  Speed
    100 62.6M  100 38.5M  100 24.1M  1642k  1029k  0:00:24  0:00:24 --:--:-- 3719k
    Sample data successfully loaded!
    ```

1.  **Increase size of the Kubernetes Engine node pool:**
    In this step we will use the `resize` command to increase the size of the
    node pool, we will also pass the command a numeric argument to indicate
    the new size of the node pool

    ```console
    ./cluster_ops.sh resize 2
    ```

    After several minutes the resize process will complete and the node
    pool size will now reflect the new size. The last several lines of output
    will look something like this:

    ```console
    Resizing the node pool to 2 nodes .....

    Resizing expand-contract-cluster...
    .....................done.
    Updated [https://container.googleapis.com/v1/projects/rolling-updates-poc-expand/zones/us-central1/clusters/expand-contract-cluster].
    ```

1.  **Upgrade the control plane:**

    ```console
    ./cluster_ops.sh upgrade-control
    ```

    After several minutes, the upgrade will be complete and the output should
    look similar to this:

    ```console
    Upgrading the K8s control plane .....

    Upgrading expand-contract-cluster...
    .........................................................................
    ........done.
    Updated [https://container.googleapis.com/v1/projects/rolling-updates-poc-expand/zones/us-central1/clusters/expand-contract-cluster].
    ```

1.  **Upgrade the cluster nodes:**

    ```console
    ./cluster_ops.sh upgrade-nodes
    ```
    After several minutes, the upgrade will be complete and the output should
    look similar to this:

    ```console
    Upgrading the K8s nodes .....

    Upgrading expand-contract-cluster...
    ...........................................................................
    .........................................................................
    ........done.
    Updated [https://container.googleapis.com/v1/projects/rolling-updates-poc-expand/zones/us-central1/clusters/expand-contract-cluster].
    ```

1.  **Decrease size of the Kubernetes Engine node pool:**
    In this step we will use the `resize` command to decrease the size of the
    node pool back to its original size, we will again use a numeric argument
    to indicate the new size of the node pool

    ```console
    ./cluster_ops.sh resize 1
    ```

    After several minutes the resize process will complete and the node
    pool size will now reflect the new size. The last several lines of output
    will look something like this:

    ```console
    Resizing the node pool to 1 nodes .....

    Resizing expand-contract-cluster...
    ................................................................................................................................................................................done.
    Updated [https://container.googleapis.com/v1/projects/rolling-updates-poc-expand/zones/us-central1/clusters/expand-contract-cluster].
    ```

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
    while true; do kubectl get pods; sleep 5; done
    ```

    To exit the loop, use `ctrl + c`.

    As each control plane node is replaced, other running control plane nodes
    will serve requests to the `kubectl` commands providing a zero-downtime
    upgrade. When the control plane upgrade is complete, you can see the new
    `Server Version` with:

    ```console
    kubectl version
    ```

*   **`gcloud` monitoring** You can also monitor the progress of cluster
    upgrades with the the `glcoud` command. Both completed and in-progress
    upgrades will be listed. Find the appropriate operation ID to get details
    of an upgrade.

        ```console
    gcloud container operations list
    gcloud container operations describe <OPERATION_ID> \
      --region <cluster-region>
    ```

*   **Cloud console monitoring** You can also monitor the progress of cluster
    upgrades under GCP Kubernetes Engine, select your cluster and see the progress
    showing in %.

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

    In one terminal, configure a port-forward from the elasticsearch service to
    your workstation's localhost:

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
To delete the Kubernetes Engine cluster and all other resources generated during
this example run the following command:

```console
./cluster_ops.sh delete
```

## Troubleshooting

* `E0717 09:45:59.417020    1245 portforward.go:178] lost connection to pod`

  The port-forward command will occasionally fail, especially as the cluster is
  being manipulated.  Execute the following command to reconnect:
  ```console
  kubectl port-forward svc/elasticsearch 9200
  ```

* `Currently upgrading cluster` Error:

  ```console
  ERROR: (gcloud.container.node-pools.delete) ResponseError: code=400, message=Operation operation-1529415957904-496c7278 is currently upgrading cluster blue-green-test. Please wait and try again once it is done.
  ```

* `IN_USE_ADDRESSES` Quota Error:

  ```console
  ERROR: (gcloud.container.clusters.create) ResponseError: code=403, message=Insufficient regional quota to satisfy request for resource: "IN_USE_ADDRESSES". The request requires '9.0' and is short '1.0'. The regional quota is '8.0' with '8.0' available.
  ```

  1.  Open the GCP Console and navigate to `IAM & admin` -> `Quotas`.
  1.  Filter the quotas by selecting your region under `Location`.
  1.  Check the box next to `Compute Engine API In-use IP addresses global`,
      then click `EDIT QUOTAS`.
  1.  Follow the steps to increase the quota. Quotas are not immediately
      increased.

* `CPUS` Quota Error:

  ```console
  ERROR: (gcloud.container.node-pools.create) ResponseError: code=403, message=Insufficient regional quota to satisfy request for resource: "CPUS". The request requires '12.0' and is short '3.0'. The regional quota is '24.0' with '9.0' available.
  ```
  1.  Open the GCP Console and navigate to `IAM & admin` -> `Quotas`.
  1.  Filter the quotas by selecting your region under `Location`.
  1.  Check the box next to `Compute Engine API CPUs`, then click `EDIT QUOTAS`.
  1.  Follow the steps to increase the quota. Quotas are not immediately
      increased.

* `Upgrade` Error after resize:

  ```console
  ERROR: (gcloud.container.clusters.upgrade) ResponseError: code=400,
  message=Operation operation-1528990089723-411a9049 is currently upgrading
  cluster expand-contract-cluster. Please wait and try again once it is done.
  ```

  1.  This is expected behavior if a resize of a node pool goes beyond 5 nodes,
      Kubernetes Engine will automatically scale up the control plane to manage
      additional resources, not to be confused with the Kubernetes Engine
      version upgrade we are doing as part of this demo.
  1.  Wait for the cluster to to be in a green state and continue with next
      step in the demo.

## Relevant Material

* `PodDisruptionBudgets` - [Kubernetes Disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/)
* `readinessProbe` - [Pod lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)
* [Kubernetes Engine Release Notes](https://cloud.google.com/kubernetes-engine/release-notes)
