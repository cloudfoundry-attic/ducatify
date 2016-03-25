package acceptance_test

import (
	"os/exec"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Manifest transformer", func() {
	var cmd *exec.Cmd
	BeforeEach(func() {
		cmd = exec.Command(binPath,
			"-diego", "fixtures/skeleton_vanilla.yml",
		)
	})

	It("prints YAML to stdout", func() {
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		var manifest interface{}
		err = yaml.Unmarshal(session.Out.Contents(), &manifest)
		Expect(err).NotTo(HaveOccurred())
	})

	It("outputs the same top-level keys as the input file", func() {
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		var manifest map[string]interface{}
		err = yaml.Unmarshal(session.Out.Contents(), &manifest)
		Expect(err).NotTo(HaveOccurred())

		Expect(manifest).To(HaveKey("jobs"))
		Expect(manifest).To(HaveKey("name"))
		Expect(manifest).To(HaveKey("networks"))
		Expect(manifest).To(HaveKey("properties"))
		Expect(manifest).To(HaveKey("releases"))
		Expect(manifest).To(HaveKey("resource_pools"))
		Expect(manifest).To(HaveKey("update"))
	})
})
