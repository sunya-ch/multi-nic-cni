# End-to-end Scaling Test

This test leverages [kwok](https://github.com/kubernetes-sigs/kwok) to simulate a big cluster with more than 100 nodes in minutes.

## Requirements
- make, yq, jq
- docker
- kubectl

## Cluster Preparation

### Kind cluster
> Need `kind` command installed.

*Steps:*

1. create kind cluster with 1000 max pods

    ```bash
    export CLUSTER_NAME=kind-1000
    make create-kind
    ```

2. build and load all requried images to kind cluster

    ```bash
    make build-load-images
    ```

3. deploy required controllers and CR (kwok, multi-nic-cni, net-attach-def CR)

    ```bash
    make prepare-controller
    ```

### Remote cluster

> Need kubernetes cluster deployed.

Confirm the current context:

```bash
kubectl config view --minify
```

*Steps:*

1. Deploy multi-nic-cni controller. Check [installation guide](https://foundation-model-stack.github.io/multi-nic-cni/user_guide/).

2. Configure to use fake daemon

    ```bash
    ./script.sh deploy_fake_ds
    ```

3. Deploy kwok controller 

    ```bash
    ./script.sh deploy_kwok
    ```

4. Set the following environments: OPERATOR_NAMESPACE, DAEMON_STUB_IMG, CNI_STUB_IMG

    ```bash
    export OPERATOR_NAMESPACE="openshift-operators"
    export DAEMON_STUB_IMG="ghcr.io/foundation-model-stack/multi-nic-daemon-stub:v1.0.3"
    export CNI_STUB_IMG="ghcr.io/foundation-model-stack/multi-nic-cni-stub:v1.0.3"
    ```

### Test cases

There are three test cases.
1. Scale cluster in steps from 10, 20, 50, 100 and to 200. Then, scale down with the same steps.
    ```bash
    ./script.sh test_step_scale
    ```
2. Allocate IPs for 5 pods to 10 nodes and then deallocate.
    ```bash
    ./script.sh test_allocate
    ```
3. Taint a node, allocate IPs for other available nodes, then untaint the node. Next, taint the node that already has pods deployed and then untaint it. 
    ```bash
    ./script.sh test_taint
    ```

For each state change, check corresponding MultiNicNetowrk, CIDR, and IPPool CRs.
