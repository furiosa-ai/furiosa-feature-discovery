package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

func abs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic("check path")
	}

	return absPath
}

var _ = Describe("end-to-end test", func() {
	Context("test furiosa feature discovery", Ordered, func() {
		It("check if Furiosa label already exists", checkNodeLabel("furiosa.ai/npu.count", false))

		It("delete furiosa-feature-discovery daemonset if exists", deleteDaemonSet("furiosa-feature-discovery", "kube-system"))

		It("delete node-feature-discovery daemonset if exists", deleteDaemonSet("furiosa-feature-discovery-node-feature-discovery-worker", "kube-system"))

		It("deploy feature discovery helm chart", e2e.DeployHelmChart("furiosa-feature-discovery", abs("../deployments/helm"), composeValues()))

		It("wait for label update", func() { time.Sleep(120 * time.Second) })

		It("check if Furiosa label is updated", checkNodeLabel("furiosa.ai/npu.count", true))

		It("delete helm chart", e2e.DeleteHelmChart())
	})
})

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func composeValues() string {
	imageRegistry := getEnv("E2E_TEST_IMAGE_REGISTRY", "registry.corp.furiosa.ai/furiosa")
	imageName := getEnv("E2E_TEST_IMAGE_NAME", "furiosa-feature-discovery")
	imageTag := getEnv("E2E_TEST_IMAGE_TAG", "latest")

	template := fmt.Sprintf(`namespace: kube-system
daemonSet:
  priorityClassName: system-node-critical
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  image:
    repository: %s/%s
    tag: %s
    pullPolicy: Always
  resources:
    cpu: 100m
    memory: 64Mi
`, imageRegistry, imageName, imageTag)
	return template
}

func checkNodeLabel(targetLabel string, expected bool) func() {
	return func() {
		nodeList, err := e2e.BackgroundContext().ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		Expect(err).To(BeNil())
		Expect(len(nodeList.Items)).Should(BeNumerically(">=", 1))

		for _, node := range nodeList.Items {
			_, isExist := node.Labels[targetLabel]
			Expect(isExist).Should(BeEquivalentTo(expected))
		}
	}
}

func deleteDaemonSet(targetDaemonSet string, namespace string) func() {
	return func() {
		err := e2e.BackgroundContext().ClientSet.AppsV1().DaemonSets(namespace).Delete(context.TODO(), targetDaemonSet, metav1.DeleteOptions{})

		if err == nil {
			fmt.Printf("DaemonSet %s already exists. Delete it\n", targetDaemonSet)
		}
	}
}
