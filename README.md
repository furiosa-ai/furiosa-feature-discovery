# Furiosa Feature Discovery

## Overview
The Furiosa Feature Discovery is a tool that automatically labels Kubernetes nodes with NPU properties if a node has one or more NPU device.

## Labels
The followings are the labels that the Furiosa Feature Discovery attaches and what they mean. 

|      Label                            |     Value                      | Description                          |
|---------------------------------------|--------------------------------|--------------------------------------|
|furiosa.ai/npu.count                   | n                              | # of NPU devices                     |
|furiosa.ai/npu.family                  | warboy, rngd                   | Chip family                          |
|furiosa.ai/npu.product                 | warboy, rngd, rngd-s, rngd-max | Chip product name                    |
|furiosa.ai/npu.driver.version          | x.y.z                          | NPU device driver version            |
|furiosa.ai/npu.driver.version.major    | x                              | NPU device driver version major part |
|furiosa.ai/npu.driver.version.minor    | y                              | NPU device driver version minor part |
|furiosa.ai/npu.driver.version.patch    | z                              | NPU device driver version patch part |
|furiosa.ai/npu.driver.version.metadata | abcxyz                         | NPU device driver version metadata   |

## Deployment

### Kubernetes
The helm chart is available at [deployment/helm](deployments/helm) directory. To configure deployment as you need, you can modify [deployments/helm/values.yaml](deployments/helm/values.yaml).

## Usage Examples
Simply, if you want to constrain your Pod to nodes with a specific NPU hardware, your Pod manifest can be as following:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: shell-test
  labels:
    env: test
spec:
  containers:
  - name: ubuntu
    image: ubuntu:focal
    imagePullPolicy: IfNotPresent
    command: ["sleep"]
    args: ["120"]
    resources:
      limits:
        furiosa.ai/warboy: 1 # requesting 1 warboy NPU
  nodeSelector:
    furiosa.ai/npu.product: warboy
```

There are more options to deploy your Kubernetes application according to Node labels.
Please refer to [Assigning Pods to Nodes](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) for more details.