package e2e

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	helmclient "github.com/mittwald/go-helm-client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultKubeConfigPath = ".kube/config"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

// framework is container for components can be reused for each test
type framework struct {
	clientConfig *rest.Config
	clientSet    clientset.Interface
	namespace    string
	helmClient   helmclient.Client
	helmChart    *helmclient.ChartSpec
}

func abs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic("check path")
	}

	return absPath
}

func newFrameworkWithDefaultNamespace() (*framework, error) {
	var defaultNS = "default"

	homePath, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	kubeconfig := homePath + "/" + defaultKubeConfigPath
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	helmChartClient, err := helmclient.NewClientFromRestConf(&helmclient.RestConfClientOptions{
		Options: &helmclient.Options{
			Namespace: defaultNS,
		},
		RestConfig: config,
	})
	if err != nil {
		return nil, err
	}

	return &framework{
		clientConfig: config,
		clientSet:    clientSet,
		helmClient:   helmChartClient,
		namespace:    defaultNS,
		helmChart:    nil,
	}, nil

}

var frk *framework

// TODO(@bg): we may need to set up kubernetes cluster in e2e-test to run test for supported versions
var _ = BeforeSuite(func() {
	newFrk, err := newFrameworkWithDefaultNamespace()
	Expect(err).To(BeNil())
	frk = newFrk

	// list namespaces to ensure api-server accessibility
	list, err := frk.clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	Expect(err).To(BeNil())
	Expect(len(list.Items)).Should(BeNumerically(">=", 1))
})

// Note(@bg): assumption is that this test will be run on the two socket workstation with two NPUs per each socket.
// TODO: enhance test to parse resource name from node object
var _ = Describe("end-to-end test", func() {
	Context("test furiosa feature discovery", func() {
		It("check if Furiosa label already exists", checkNodeLabel("furiosa.ai/npu.count", false))

		It("delete furiosa-feature-discovery daemonset if exists", deleteDaemonSet("furiosa-feature-discovery", "default"))

		It("delete node-feature-discovery daemonset if exists", deleteDaemonSet("furiosa-feature-discovery-node-feature-discovery-worker", "default"))

		It("deploy feature discovery helm chart", deployHelmChart())

		It("wait for label update", func() {time.Sleep(120 * time.Second)})

		It("check if Furiosa label is updated", checkNodeLabel("furiosa.ai/npu.count", true))

		It("delete helm chart", deleteHelmChart())
	})
})

func composeValues() string {
	template := `namespace: default
daemonSet:
  priorityClassName: system-node-critical
  # Use OnDelete for change the plugin strategy
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  tolerations:
    - key: npu
      operator: Exists
  image:
    repository: registry.corp.furiosa.ai/furiosa/furiosa-feature-discovery
    tag: latest
    pullPolicy: Always
  resources:
    cpu: 100m
    memory: 64Mi
`
	return template
}

func strRand() string {
	return fmt.Sprintf("%d", rand.Int())
}

func deployHelmChart() func() {
	return func() {
		helmChartSpec := &helmclient.ChartSpec{
			ReleaseName:     "furiosa-feature-discovery" + strRand(),
			ChartName:       abs("../deployments/helm"), //path to helm chart
			Namespace:       frk.namespace,
			CreateNamespace: false,
			Wait:            false,
			Timeout:         5 * time.Minute,
			CleanupOnFail:   false,
			ValuesYaml:      composeValues(),
		}
		frk.helmChart = helmChartSpec

		_, err := frk.helmClient.InstallChart(context.TODO(), frk.helmChart, nil)
		Expect(err).To(BeNil())
	}
}

func checkNodeLabel(targetLabel string, expected bool) func() {
	return func() {
		nodeList, err := frk.clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		Expect(err).To(BeNil())
		Expect(len(nodeList.Items)).Should(BeNumerically(">=", 1))

		for _, node := range(nodeList.Items) {
			// fmt.Printf("%v\n", node.Labels)
			_, is_exist := node.Labels[targetLabel]
			Expect(is_exist).Should(BeEquivalentTo(expected))
		}
	}
}

func deleteDaemonSet(targetDaemonSet string, namespace string) func() {
	return func() {
		err := frk.clientSet.AppsV1().DaemonSets(namespace).Delete(context.TODO(), targetDaemonSet, metav1.DeleteOptions{})

		if err == nil {
			fmt.Printf("DaemonSet %s already exists. Delete it\n", targetDaemonSet)
		}
	}
}

func deleteHelmChart() func() {
	return func() {
		err := frk.helmClient.UninstallRelease(frk.helmChart)
		Expect(err).To(BeNil())
	}
}
