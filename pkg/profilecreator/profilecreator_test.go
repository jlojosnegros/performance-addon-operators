package profilecreator

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift-kni/performance-addon-operators/pkg/controller/performanceprofile/components"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PerformanceProfileCreator: MCP and Node Matching", func() {
	var node *v1.Node
	var labelSelector *metav1.LabelSelector
	BeforeEach(func() {
		node = newTestNode("test-node")
		labelSelector = metav1.AddLabelToSelector(&metav1.LabelSelector{}, "node-role.kubernetes.io/worker-cnf", "")
	})

	Context("Identifying Nodes targetted by MCP", func() {
		It("should find a match", func() {
			nodeLabel := map[string]string{
				"node-role.kubernetes.io/worker-cnf": "",
			}
			node.Labels = nodeLabel
			nodes := newTestNodeList(node)
			matchedNodes, err := GetMatchedNodes(nodes, labelSelector)
			Expect(err).ToNot(HaveOccurred())
			Expect(matchedNodes).ToNot(BeNil())
			Expect(len(matchedNodes)).ToNot(Equal(0))
			Expect(matchedNodes[0].GetName()).To(Equal("test-node"))
		})
		It("should not find a match", func() {
			nodeLabel := map[string]string{
				"node-role.kubernetes.io/foo": "",
			}
			node.Labels = nodeLabel
			nodes := newTestNodeList(node)
			matchedNodes, err := GetMatchedNodes(nodes, labelSelector)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(matchedNodes)).To(Equal(0))
		})
	})
})

var _ = Describe("PerformanceProfileCreator: Getting MCP from Must Gather", func() {
	var mustGatherDirPath, mcpName, mcpNodeSelectorKey, mustGatherDirAbsolutePath string
	var err error
	Context("Identifying Nodes targetted by MCP", func() {
		It("gets the MCP successfully", func() {
			mcpName = "worker-cnf"
			mcpNodeSelectorKey = "node-role.kubernetes.io/worker-cnf"
			mustGatherDirPath = "../../testdata/must-gather/must-gather.local.directory"
			mustGatherDirAbsolutePath, err = filepath.Abs(mustGatherDirPath)
			Expect(err).ToNot(HaveOccurred())
			mcp, err := GetMCP(mustGatherDirAbsolutePath, mcpName)
			k, _ := components.GetFirstKeyAndValue(mcp.Spec.NodeSelector.MatchLabels)
			Expect(err).ToNot(HaveOccurred())
			Expect(k).To(Equal(mcpNodeSelectorKey))
		})
		It("fails to get MCP as an MCP with that name doesn't exist", func() {
			mcpName = "foo"
			mustGatherDirPath = "../../testdata/must-gather/must-gather.local.directory"
			mustGatherDirAbsolutePath, err = filepath.Abs(mustGatherDirPath)
			mcp, err := GetMCP(mustGatherDirAbsolutePath, mcpName)
			Expect(mcp).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
		It("fails to get MCP due to misconfigured must-gather path", func() {
			mcpName = "worker-cnf"
			mustGatherDirPath = "../../testdata/must-gather/foo-path"
			mustGatherDirAbsolutePath, err = filepath.Abs(mustGatherDirPath)
			Expect(err).ToNot(HaveOccurred())
			_, err := GetMCP(mustGatherDirAbsolutePath, mcpName)
			Expect(err).To(HaveOccurred())
		})

	})
})

var _ = Describe("PerformanceProfileCreator: Getting Nodes from Must Gather", func() {
	var mustGatherDirPath, mustGatherDirAbsolutePath string
	var err error

	Context("Identifying Nodes in the cluster", func() {
		It("gets the Nodes successfully", func() {
			mustGatherDirPath = "../../testdata/must-gather/must-gather.local.directory"
			mustGatherDirAbsolutePath, err = filepath.Abs(mustGatherDirPath)
			Expect(err).ToNot(HaveOccurred())
			nodes, err := GetNodeList(mustGatherDirAbsolutePath)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(nodes)).To(Equal(5))
		})
		It("fails to get Nodes due to misconfigured must-gather path", func() {
			mustGatherDirPath = "../../testdata/must-gather/foo-path"
			mustGatherDirAbsolutePath, err = filepath.Abs(mustGatherDirPath)
			_, err := GetNodeList(mustGatherDirAbsolutePath)
			Expect(err).To(HaveOccurred())
		})

	})
})

var _ = Describe("PerformanceProfileCreator: Consuming GHW Snapshot from Must Gather", func() {
	var mustGatherDirPath, mustGatherDirAbsolutePath string
	var node *v1.Node
	var err error

	Context("Identifying Nodes Info of the nodes cluster", func() {
		It("gets the Nodes Info successfully", func() {
			node = newTestNode("cnfd1-worker-0.fci1.kni.lab.eng.bos.redhat.com")
			mustGatherDirPath = "../../testdata/must-gather/must-gather.local.directory"
			mustGatherDirAbsolutePath, err = filepath.Abs(mustGatherDirPath)
			Expect(err).ToNot(HaveOccurred())
			handle, err := NewGHWHandler(mustGatherDirAbsolutePath, node)
			Expect(err).ToNot(HaveOccurred())
			cpuInfo, err := handle.CPU()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(cpuInfo.Processors)).To(Equal(2))
			Expect(int(cpuInfo.TotalCores)).To(Equal(40))
			Expect(int(cpuInfo.TotalThreads)).To(Equal(80))
			topologyInfo, err := handle.Topology()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(topologyInfo.Nodes)).To(Equal(2))
		})
		It("fails to get Nodes Info due to misconfigured must-gather path", func() {
			mustGatherDirPath = "../../testdata/must-gather/foo-path"
			mustGatherDirAbsolutePath, err = filepath.Abs(mustGatherDirPath)
			_, err := NewGHWHandler(mustGatherDirAbsolutePath, node)
			Expect(err).To(HaveOccurred())
		})
		It("fails to get Nodes Info for a node that does not exist", func() {
			node = newTestNode("foo")
			mustGatherDirPath = "../../testdata/must-gather/must-gather.local.directory"
			mustGatherDirAbsolutePath, err = filepath.Abs(mustGatherDirPath)
			_, err := NewGHWHandler(mustGatherDirAbsolutePath, node)
			Expect(err).To(HaveOccurred())
		})

	})
})