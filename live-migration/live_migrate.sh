#!/bin/bash

if [ -z ${OPERATOR_NAMESPACE} ]; then
    OPERATOR_NAMESPACE=openshift-operators
fi

if [ -z ${CLUSTER_NAME} ]; then
    CLUSTER_NAME="default"
fi

#############################################
# utility functions

get_netname() {
    kubectl get multinicnetwork -ojson|jq .items| jq '.[].metadata.name'| tr -d '"'
}

get_controller() {
    kubectl get po -n ${OPERATOR_NAMESPACE}|grep multi-nic-cni-operator-controller-manager|awk '{print $1}'
}

get_controller_log() {
    controller=$(get_controller)
    kubectl logs $controller -n ${OPERATOR_NAMESPACE} -c manager
}

get_status() {
    kubectl get multinicnetwork -o custom-columns=NAME:.metadata.name,ConfigStatus:.status.configStatus,RouteStatus:.status.routeStatus,TotalHost:.status.discovery.existDaemon,HostWithSecondaryNIC:.status.discovery.infoAvailable,ProcessedHost:.status.discovery.cidrProcessed,Time:.status.lastSyncTime
}

get_secondary_ip() {
   PODNAME=$1
   kubectl get po $PODNAME -ojson|jq .metadata.annotations|jq '.["k8s.v1.cni.cncf.io/network-status"]'| jq -r |jq .[1].ips[0]
}

apply() {
    export REPLACEMENT=$1
    export YAMLFILE=$2
    yq -e ${REPLACEMENT} ${YAMLFILE}.yaml|kubectl apply -f -
}

catfile() {
    export REPLACEMENT=$1
    export YAMLFILE=$2
    yq -e ${REPLACEMENT} ${YAMLFILE}.yaml|cat
}

create_replacement() {
    export LOCATION=$1
    export REPLACE_VALUE=$2
    echo "(${LOCATION}=${REPLACE_VALUE})"
}

#############################################

#############################################
# cr handling

status_cr="cidrs.multinic ippools.multinic hostinterfaces.multinic"
activate_cr="multinicnetworks.multinic"
config_cr="configs.multinic deviceclasses.multinic"

_snapshot_resource() {
    dir=$1
    mkdir -p $dir
    kind=$2
    item=$3
    kubectl get $kind $item -ojson | jq 'del(.metadata.resourceVersion,.metadata.uid,.metadata.selfLink,.metadata.creationTimestamp,.metadata.generation,.metadata.ownerReferences)' | yq eval - -P > $dir/$kind-$item.yaml
    echo "snapshot $dir/$kind-$item.yaml"
}

_snapshot() {
    dir=$1
    cr=$2
    itemlist=$(kubectl get $cr -ojson |jq '.items'| jq 'del(.[].status,.[].metadata.finalizers,.[].metadata.resourceVersion,.[].metadata.uid,.[].metadata.selfLink,.[].metadata.creationTimestamp,.[].metadata.generation,.[].metadata.ownerReferences)') 
    echo {"apiVersion": "v1", "items": $itemlist, "kind": "List"}| yq eval - -P > $dir/$cr.yaml
}

snapshot() {
    mkdir -p snapshot
    # update l2 file used to stop controller to modify the route
    cp multinicnetwork_l2.yaml snapshot/multinicnetwork_l2.yaml
    netname=$(get_netname)
    yq -e -i .metadata.name=\"$netname\" snapshot/multinicnetwork_l2.yaml
    echo "rename multinicnetwork_l2.yaml with $netname"
    # snapshot state
    snapshot_dir="snapshot/${CLUSTER_NAME}"
    mkdir -p $snapshot_dir
    for cr in $status_cr $activate_cr
    do
        _snapshot $snapshot_dir $cr
    done
    ls $snapshot_dir
    echo "saved in $snapshot_dir"
}

deploy_status_cr() {
    snapshot_dir="snapshot/${CLUSTER_NAME}"
    for cr in $status_cr
    do
        kubectl apply -f $snapshot_dir/$cr.yaml
    done
}

#############################################

#############################################
# route handling

deactivate_route_config() {
    kubectl apply -f snapshot/multinicnetwork_l2.yaml
    sleep 5
    configSTR=$(kubectl get multinicnetwork $(get_netname) -ojson|jq '.spec.multiNICIPAM')
    if [[ "$configSTR" == "false" ]]; then
        echo "Deactivate route configuration."
    fi
}

activate_route_config() {
    snapshot_dir="snapshot/${CLUSTER_NAME}"
    kubectl apply -f $snapshot_dir/$activate_cr.yaml
    sleep 5
    configSTR=$(kubectl get multinicnetwork $(get_netname) -ojson|jq '.spec.multiNICIPAM')
    if [[ "$configSTR" == "true" ]]; then
        echo "Activate route configuration."
    fi
}

#############################################

#############################################
# operator resource handling: controller, daemon, crd

_clean_resource() {
    for cr in $status_cr $activate_cr $config_cr
    do
        kubectl delete $cr --all
    done
    wait_daemon_terminated
}

clean_resource() {
    deactivate_route_config
    _clean_resource
}

wait_daemon_terminated() {
    kubectl wait --for=delete daemonset/multi-nicd -n ${OPERATOR_NAMESPACE} --timeout=300s
    # wait for all terminated
    daemonTerminated=$(kubectl get po -n ${OPERATOR_NAMESPACE}|grep multi-nicd|wc -l|tr -d ' ')
    while [ "$daemonTerminated" != 0 ] ; 
    do
        echo "Wait for daemonset to be fully terminated...($daemonTerminated left)"
        sleep 10
        daemonTerminated=$(kubectl get po -n ${OPERATOR_NAMESPACE}|grep multi-nicd|wc -l|tr -d ' ')
    done
    echo "Done"
}

uninstall_operator() {
    version=$1
    # uninstall operator
    kubectl delete subscriptions.operators.coreos.com multi-nic-cni-operator -n $OPERATOR_NAMESPACE
    kubectl delete clusterserviceversion multi-nic-cni-operator.v${version} -n $OPERATOR_NAMESPACE
    kubectl delete ds multi-nicd -n $OPERATOR_NAMESPACE
}

clean_crd() {
    for cr in $status_cr $activate_cr $config_cr
    do
        kubectl delete crd $cr.fms.io
    done
}

# after reinstall operator
# deactivate_route_config

patch_daemon() {
    kubectl patch config.multinic multi-nicd --type merge --patch '{"spec": {"daemon": {"imagePullPolicy": "Always"}}}'
    kubectl delete po -l app=multi-nicd -n ${OPERATOR_NAMESPACE}
}

wait_daemon() {
    # wait for daemon creation
    sleep 5
    daemonCreate=$(kubectl get ds multi-nicd -n ${OPERATOR_NAMESPACE}|wc -l|tr -d ' ')
    while [ "$daemonCreate" == 0 ] ; 
    do
        echo "Wait for daemonset to be created by controller..."
        sleep 2
        daemonCreate=$(kubectl get ds multi-nicd -n ${OPERATOR_NAMESPACE}|wc -l|tr -d ' ')
    done
    echo "Wait for daemonset to be ready"
    kubectl rollout status daemonset multi-nicd -n ${OPERATOR_NAMESPACE} --timeout 300s
}

restart_controller() {
    controller=$(get_controller)
    kubectl delete po $controller -n ${OPERATOR_NAMESPACE}
    echo "Wait for deployment to be available"
    kubectl wait deployment -n ${OPERATOR_NAMESPACE} multi-nic-cni-operator-controller-manager --for condition=Available=True --timeout=90s
    ready=$(echo $(get_controller_log)|grep ConfigReady)
    while [ -z "$ready" ];
    do
        sleep 5
        echo "Wait for config to be ready..."
        ready=$(echo $(get_controller_log)|grep ConfigReady)
    done
    echo "Config Ready"
}

check_cidr() {
    nhost=$1
    nvlan=$2
    cidr=$(kubectl get cidr $(get_netname) -ojson|jq .spec.cidr)
    vlanlen=$(echo $cidr| jq '. | length')
    if [ "$nhost" != 0 ] && [ "$vlanlen" != $nvlan ] ; then
        echo >&2 "Fatal error: interface length $vlanlen != $nvlan"
        exit 2
    else
        i=0
        while [ "$i" -lt $nvlan ]; do
            hosts=$(echo $cidr| jq .[${i}].hosts)
            hostlen=$(echo $hosts| jq '.|length')
            if [ "$hostlen" != $nhost ] ; then
                echo >&2 "Fatal error: host length $hostlen != $nhost"
                exit 2
            fi
            i=$(( i + 1 ))
        done 
    fi
}

wait_node() {
    for nodename in $(kubectl get nodes |awk '(NR>1){print $1}'); do
        kubectl wait node ${nodename} --for condition=Ready --timeout=1000s
    done
}

wait_node_readiness() {
    wait_node
    nhost=$(kubectl get nodes -l node-role.kubernetes.io/worker -ojson|jq '.items|length')
    until $(check_cidr $nhost $1); 
    do
        echo "wait for CIDR ready, sleep 10s"
        sleep 10
    done
}

#############################################

#############################################
# iperf live

# ./live_migrate.sh live_iperf3 <SERVER_HOST_NAME> <CLIENT_HOST_NAME> <LIVE_TIME>
live_iperf3() {
   SERVER_HOST_NAME=$1
   CLIENT_HOST_NAME=$2
   LIVE_TIME=$3
   NETWORK_NAME=$(get_netname)
   NETWORK_REPLACEMENT=$(create_replacement .metadata.annotations.\"k8s.v1.cni.cncf.io/networks\" \"${NETWORK_NAME}\")
   SERVER_HOSTNAME_REPLACEMENT=$(create_replacement .spec.nodeName \"${SERVER_HOST_NAME}\")
   CLIENT_HOSTNAME_REPLACEMENT=$(create_replacement .spec.nodeName \"${CLIENT_HOST_NAME}\")

   SERVER_NAME="multi-nic-iperf3-server"
   CLIENT_NAME="multi-nic-iperf3-client"

   SERVER_NAME_REPLACEMENT=$(create_replacement .metadata.name \"${SERVER_NAME}\")
   # deploy server pod
   apply ${SERVER_NAME_REPLACEMENT},${NETWORK_REPLACEMENT},${SERVER_HOSTNAME_REPLACEMENT} ./test/iperf3/server

   # wait until server available
   kubectl wait pod ${SERVER_NAME} --for condition=ready --timeout=90s

   SECONDARY_IP=$(get_secondary_ip ${SERVER_NAME}| tr -d '"')
   CLIENT_NAME_REPLACEMENT=$(create_replacement .metadata.name \"${CLIENT_NAME}\")
   # deploy client pod
   apply ${CLIENT_NAME_REPLACEMENT},${NETWORK_REPLACEMENT},${CLIENT_HOSTNAME_REPLACEMENT} ./test/iperf3/client

   if [[ "${SECONDARY_IP}" == "null" ]]; then
        echo >&2 "cannot get secondary IP of server ${SERVER_NAME}"
        exit 2
   fi
   # wait until client available
   kubectl wait pod ${CLIENT_NAME} --for condition=ready --timeout=90s
   # run live client
   kubectl exec -it ${CLIENT_NAME} -- iperf3 -c ${SECONDARY_IP} -t ${LIVE_TIME} -p 30000

   # clean up
   kubectl delete pod ${CLIENT_NAME} ${SERVER_NAME}
}

# kubeflow test
mlbench_base() {
    # require training operator
    if [[ ! $(kubectl get crd pytorchjobs.kubeflow.org) ]]; then
        echo >&2 "pytorchjobs.kubeflow.org not available, please install training-operator"
        exit 2
    fi
    NETWORK_NAME=$(get_netname)
    MASTER_NETWORK_REPLACEMENT=$(create_replacement .spec.pytorchReplicaSpecs.Master.template.metadata.annotations.\"k8s.v1.cni.cncf.io/networks\" \"${NETWORK_NAME}\")
    WORKER_NETWORK_REPLACEMENT=$(create_replacement .spec.pytorchReplicaSpecs.Worker.template.metadata.annotations.\"k8s.v1.cni.cncf.io/networks\" \"${NETWORK_NAME}\")
    catfile ${MASTER_NETWORK_REPLACEMENT},${WORKER_NETWORK_REPLACEMENT} ./test/kubeflow/mlbench/pytorch-job
}

mlbench_with_cpe() {
    # deploy operator
    kubectl apply -f ./test/kubeflow/mlbench/cpe_benchmark_operator.yaml
    # deploy configmap
    kubectl apply -f ./test/kubeflow/mlbench/pytorch-cfm.yaml

    spec=$(mlbench_base|yq .spec)  yq -e '.spec.benchmarkSpec = strenv(spec)' ./test/kubeflow/mlbench/cpe_benchmark.yaml|kubectl apply -f -
    echo "Wait for job to be completed, sleep 1m"
    sleep 60
    jobCompleted=$(kubectl get benchmark mlbench -ojson|jq -r .status.jobCompleted)
    while [ "$jobCompleted" != "6/6" ] ; 
    do  
        echo "$jobCompleted completed, sleep 10s"
        sleep 10
        jobCompleted=$(kubectl get benchmark mlbench -ojson|jq -r .status.jobCompleted)
    done
    kubectl get benchmark mlbench -oyaml|yq .status.bestResults
}

mlbench() {
    # deploy configmap
    kubectl apply -f ./test/kubeflow/mlbench/pytorch-cfm.yaml

    mlbench_base|kubectl apply -f -
}
#############################################


"$@"


