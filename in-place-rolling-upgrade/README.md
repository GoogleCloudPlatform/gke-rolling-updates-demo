# In-place rolling upgrade and downgrade

<!--ts-->

* [Introduction](#introduction)
* [Architecture](#architecture)
* [Prerequisites](#prerequisites)
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

This code repository demonstrates the in-place rolling upgrades and downgrades
of Kubernetes Engine clusters.

## Architecture

The in-place rolling upgrade is the simplest upgrade procedure and is ideal for
clusters containing stateless workloads where little attention must be paid to
restarting and rescheduling of application instances (pods).

In-place means that each node will be drained, terminated, and replaced with
a new node running the new Kubernetes Engine version.  Each node is replaced
one at a time.  For the control plane, only upgrades are supported.  For the
node pools, upgrades and downgrades are both supported.  Downgrades to the node
pool are only permitted one minor version at a time.

We will be using a Kubernetes Engine [Regional Cluster](https://cloud.google.com/kubernetes-engine/docs/concepts/multi-zone-and-regional-clusters) as the
control plane has a multi-node HA architecture and can be upgraded without any
downtime for the API server.

Kubernetes Engine clusters are upgraded in two distinct phases.  First, the
control plane is updated and second the node pool[s] are upgraded. This order
must be followed because node versions can never advance beyond the version of
the control plane.

Kubernetes Engine performs automated control plane upgrades.  The timing of
these upgrades are announced on the [Kubernetes Engine Release Notes](https://cloud.google.com/kubernetes-engine/release-notes) Page.
You may also upgrade your control plane manually as soon as a new version is
available. For nodes, you can opt-in to automated upgrades.  When node pools
are configured for automated upgrades, Kubernetes Engine will make sure the
control plane is upgraded first.

Supported versions of Kubernetes Engine are region and zone dependent and can
change frequently.  Check for the available versions in a region/zone of
interest with the following command:
```console
gcloud container get-server-config [--zone <zone-name>] [--region <region-name>]
```

## Prerequisites

### Tools

1. [gcloud](https://cloud.google.com/sdk/downloads) (Google Cloud SDK version >= 200.0.0)
1. [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) >= 1.10.0
1. [terraform](https://www.terraform.io/downloads.html) >= 10.8
1.  [jq](https://stedolan.github.io/jq/)
1. bash or bash compatible shell
1. A Google Cloud Platform project with the Kubernetes Engine API enabled.
    ```
    gcloud services list
    gcloud services enable container.googleapis.com
    ```

### Configuration

##### `.env` Properties

A number of environment variables must be set to run the `cluster_ops.sh`
script.  The required variables are provided and explained in the `env` file
found in the root of this repository. Make a copy in the root of this
repository:
```
cp env .env
```
Update the `.env` file with appropriate values for your use.  It will be
sourced by the `cluster_ops.sh` script each time it is run.

##### Selecting your versions

In the `.env` file, you must select two Kubernetes versions, `GKE_VER` and
`NEW_GKE_VER`.  Due to the way the current Google Cloud Terraform provider
handles the version state, you must use the full Kubernetes Engine patch
version.  This example was tested using the following versions:
```
GKE_VER=1.9.7-gke.3
NEW_GKE_VER=1.10.4-gke.2
```
With this command, you can find the currently available Kubernetes Engine
versions:
```
gcloud container get-server-config --region <your-region>
```

**Note:** The example application manifests are using the apps/v1 API for the
`hello-server` deployment so GKE_VER must be greater than `1.9.0-gke.0`.

**Information** about upcoming automated upgrades and Kuberenetes Engine
version deprecations can be found in the [Kubernetes Engine Release Notes](https://cloud.google.com/kubernetes-engine/release-notes).

## Deployment

The following steps walk through creating the cluster, upgrading the control
plane, upgrading the node pool, and downgrading the node pool.

### Manual Deployment

Run each command from the root of this repository.

1.  Create the Kubernetes Engine cluster and deploy the example application:
    ```console
    ./cluster_ops.sh create
    ```
    After the terraform init, plan and apply, the application will be created.
    The last output should look like this:
    ```console
    Fetching cluster endpoint and auth data.
    kubeconfig entry generated for upgrade-test.
    deployment.apps "hello-server" created
    service "hello-server" created
    ```

1.  Upgrade the control plane
    ```console
    ./cluster_ops.sh upgrade-control
    ```
    When the upgrade is complete, terraform will output this success message
    followed by the control plane and node version outputs defined in
    `outputs.tf`:
    ```console
    Apply complete! Resources: 0 added, 1 changed, 0 destroyed.
    ```

1.  Upgrade the default node pool
    ```console
    ./cluster_ops.sh upgrade-nodes
    ```
    The output will be very similar to the previous step.

1.  Downgrade the default node pool.
    ```console
    ./cluster_ops.sh downgrade-nodes
    ```
    The output will be very similar to the previous step.

### Automated Deployment

The cluster creation, upgrade, and validation can be run with one command:
```console
./cluster_ops.sh auto
```

## Validation

During upgrades and downgrades, each node will take several minutes to replace.
You can monitor the progress in another terminal or cloud console as the
cluster changes proceed.

*   While the control plane is upgrading, you can verify that Regional
    Kubernetes Engine clusters have an HA control plane by querying the API
    server in a loop:
    ```console
    watch kubectl get pods
    ```
    As each control plane node is replaced, other running control plane nodes
    will serve requests to the `kubectl` commands providing a zero-downtime
    upgrade.
    When the control plane upgrade is complete, you can see the new `Server
    Version` with
    ```
    kubectl version
    ```
*   You can also monitor the progress of cluster upgrades with the the `glcoud`
    command.  Both completed and in-progress upgrades will be listed.  Find the
    appropriate operation ID to get details of an upgrade.
    ```console
    gcloud container operations list
    gcloud container operations describe <OPERATION_ID> \
      --region <cluster-region>
    ```
*   While the node pool is upgrading and downgrading you can watch the nodes
    get cordoned, removed, and created with the new kubernetes version with
    the following command:
    ```console
    watch kubectl get nodes
    ```
    You can also watch the pods get rescheduled as each node is drained:
    ```console
    watch kubectl get pods
    ```
*   Provided your applications have an HA architecture with enough replicas,
    throughout all upgrade and downgrade steps the applications and services
    within the cluster should continue running uninterrupted.  In this example
    we have installed a simple stateless webapp with 3 pods and a load balanced
    service to monitor during the upgrade/downgrade procedures.

    Find the IP address of the `hello-server` Service Load Balancer and test
    the ip in your browser. (`<External-IP:8080`)

    ```console
    kubectl get svc
    NAME           TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)          AGE
    hello-server   LoadBalancer   10.39.246.40   35.237.184.173   8080:31766/TCP   2m
    kubernetes     ClusterIP      10.39.240.1    <none>           443/TCP          50m
    ```

    ![hello-app-from-browser](../images/hello-app-browser.png)

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

All the resources created in this example are stored in the terraform state
file and can be deleted with terraform:
```console
./cluster_ops.sh delete
```

## Troubleshooting

* `Currently upgrading cluster` Error:
  ```
  ERROR: (gcloud.container.node-pools.delete) ResponseError: code=400, message=Operation operation-1529415957904-496c7278 is currently upgrading cluster blue-green-test. Please wait and try again once it is done.
  ```

* `default credentials` Error, or `Permission Denied` when running Terraform:
    ```console
    * provider.google: google: could not find default credentials. See https://developers.google.com/accounts/docs/application-default-credentials for more information.
    ```
    Set your [credentials](https://www.terraform.io/docs/providers/google/index.html#configuration-reference) through any of the available methods.
    The quickest being:
    ```console
    gcloud auth application-default login
    ```
*  Kubernetes Engine update failures:
    An audit log of all cluster operations is kept by the system. All completed
    updates as well as in-progress updates can be inspected:
    ```console
    gcloud container operations list
    gcloud container operations describe [OPERATION_ID]
    ```
*  Some operations have been observed to take longer thant the defaults
    the next timeouts have been used in Terraform scripts
    create = "30m" // default 30m
    update = "15m" // default 10m
    delete = "15m" // default 10m
    They can be set or updated in case any issues during running the scripts.

## Relevant Material

* `hello-app` from the [Kubernetes Engine Quickstart](https://cloud.google.com/kubernetes-engine/docs/quickstart)
* [Kubernetes Engine Release Notes](https://cloud.google.com/kubernetes-engine/release-notes)
