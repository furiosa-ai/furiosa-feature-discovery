.. _FeatureDiscovery:

####################################
Installing Furiosa Feature Discovery
####################################


Furiosa Feature Discovery and NFD
================================================================

The Furiosa Feature Discovery automatically labels Kubernetes nodes with information
about FuriosaAI NPU properties, such as the NPU family, count, and various versions.
Using these labels, you can schedule your Kubernetes workloads based on specific NPU requirements.

The Furiosa Feature Discovery leverage NFD(Node Feature Discovery) which is a tool that detects
hardware features and labels Kubernetes nodes. It is recommended to use NFD and
Furiosa Feature Discovery to ensure that the Cloud Native Toolkit is deployed only on nodes
equipped with FuriosaAI NPUs.

.. note::

  If multiple versions of device-level components (like firmware) exist on the same node, the related labels are removed to avoid conflicts.


Deploying Furiosa Feature Discovery with Helm
----------------------------------------------
With the helm chart you can easily install Furiosa feature discovery and NFD into your Kubernetes cluster.
Following command shows how to install them.
The Furiosa device plugin helm chart is available at https://github.com/furiosa-ai/helm-charts. To configure deployment as you need, you can modify ``charts/furiosa-feature-discovery/values.yaml``.

.. code-block:: sh

  helm repo add furiosa https://furiosa-ai.github.io/helm-charts
  helm repo update
  helm install furiosa-feature-discovery furiosa/furiosa-feature-discovery -n <namespace>


Labels
-----------------------------

The followings are the labels that the Furiosa Feature Discovery attaches and what they mean.

.. list-table::
   :align: center
   :header-rows: 1
   :widths: 130 160 260

   * - Label
     - Value
     - Description
   * - furiosa.ai/npu.count
     - n
     - # of NPU devices
   * - furiosa.ai/npu.family
     - rngd
     - Chip family
   * - furiosa.ai/npu.product
     - rngd, rngd-s, rngd-max
     - Chip product name
   * - furiosa.ai/npu.driver.version
     - x.y.z
     - NPU device driver version
   * - furiosa.ai/npu.driver.version.major
     - x
     - NPU device driver version major part
   * - furiosa.ai/npu.driver.version.minor
     - y
     - NPU device driver version minor part
   * - furiosa.ai/npu.driver.version.patch
     - z
     - NPU device driver version patch part
   * - furiosa.ai/npu.driver.version.metadata
     - abcxyz
     - NPU device driver version metadata
   * - furiosa.ai/npu.driver.version.prerelease
     - abcxyz
     - NPU device driver version pre-release part
   * - furiosa.ai/npu.firmware.version
     - x.y.z
     - NPU firmware version
   * - furiosa.ai/npu.firmware.version.major
     - x
     - NPU firmware version major part
   * - furiosa.ai/npu.firmware.version.minor
     - y
     - NPU firmware version minor part
   * - furiosa.ai/npu.firmware.version.patch
     - z
     - NPU firmware version patch part
   * - furiosa.ai/npu.firmware.version.metadata
     - abcxyz
     - NPU firmware version metadata
   * - furiosa.ai/npu.firmware.version.prerelease
     - abcxyz
     - NPU firmware version pre-release part


License
-------

.. code-block:: text

   Copyright 2023 FuriosaAI, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
