# Kubeflow Job Test

## MLBench
source: https://mlbench.github.io

All reduce test.

Run without CPE *- only test on first secondary NIC with gloo*

```bash
./live_migrate.sh mlbench
```

Run with [CPE](https://github.com/IBM/cpe-operator) *- vary over choices of interfaces and backend (gloo, nccl)*

```bash
./live_migrate.sh mlbench_with_cpe
```

**expected output:**
```bash
# ./live_migrate.sh mlbench_with_cpe
benchmarkoperator.cpe.cogadvisor.io/pytorch-job-operator unchanged
configmap/multi-nic-mlbench-cfm configured
configmap/multi-nic-pytorch-test-cfm unchanged
benchmark.cpe.cogadvisor.io/mlbench created
Wait for job to be completed, sleep 1m
2/6 completed, sleep 10s
2/6 completed, sleep 10s
2/6 completed, sleep 10s
3/6 completed, sleep 10s
3/6 completed, sleep 10s
4/6 completed, sleep 10s
4/6 completed, sleep 10s
4/6 completed, sleep 10s
5/6 completed, sleep 10s
5/6 completed, sleep 10s
- build: init
  configurations: {}
  performanceKey: Latency(us)
  performanceValue: "5979.000000"
  scenarioID: backend=nccl;nccl;8;8;8;8;8;8;iface=eth0;eth0;eth0;eth0
- build: init
  configurations: {}
  performanceKey: Latency(us)
  performanceValue: "15238.000000"
  scenarioID: backend=gloo;gloo;0;0;0;0;0;0;iface=eth0;eth0;eth0;eth0
- build: init
  configurations: {}
  performanceKey: Latency(us)
  performanceValue: "5095.000000"
  scenarioID: backend=nccl;nccl;8;8;8;8;8;8;iface=eth0,net1-0;eth0,net1-0;eth0,net1-0;eth0,net1-0
- build: init
  configurations: {}
  performanceKey: Latency(us)
  performanceValue: "13093.000000"
  scenarioID: backend=gloo;gloo;0;0;0;0;0;0;iface=eth0,net1-0;eth0,net1-0;eth0,net1-0;eth0,net1-0
- build: init
  configurations: {}
  performanceKey: Latency(us)
  performanceValue: "4685.000000"
  scenarioID: backend=nccl;nccl;8;8;8;8;8;8;iface=net1-0;net1-0;net1-0;net1-0
- build: init
  configurations: {}
  performanceKey: Latency(us)
  performanceValue: "10106.000000"
  scenarioID: backend=gloo;gloo;0;0;0;0;0;0;iface=net1-0;net1-0;net1-0;net1-0
```

#### Modification
To make any changes to job spec such as nodeSelector, gpu request/limit, edit [pytorch-job.yaml](./mlbench/pytorch-job.yaml).