package acceptance_test

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

func loadFixture(name string) map[string]interface{} {
	bytes, err := ioutil.ReadFile(filepath.Join("fixtures", name+".yml"))
	Expect(err).NotTo(HaveOccurred())

	var data map[string]interface{}
	err = yaml.Unmarshal(bytes, &data)
	Expect(err).NotTo(HaveOccurred())

	return data
}

var _ = Describe("Manifest transformer", func() {
	var (
		cmd                                   *exec.Cmd
		vanilla, expectedOutput, actualOutput map[string]interface{}
	)

	BeforeEach(func() {
		vanilla = loadFixture("skeleton_vanilla")
		expectedOutput = loadFixture("skeleton_transformed")

		cmd = exec.Command(binPath,
			"-diego", "fixtures/skeleton_vanilla.yml",
		)

		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		err = yaml.Unmarshal(session.Out.Contents(), &actualOutput)
		Expect(err).NotTo(HaveOccurred())
	})

	It("outputs the same top-level keys as the input file", func() {
		getKeys := func(m map[string]interface{}) []string {
			keys := []string{}
			for k, _ := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			return keys
		}

		Expect(getKeys(actualOutput)).To(Equal(getKeys(expectedOutput)))
	})

	It("leaves intact most top-level keys", func() {
		for _, key := range []string{"name", "networks", "update"} {
			Expect(actualOutput).To(HaveKey(key))
			Expect(expectedOutput[key]).To(Equal(vanilla[key]))
			Expect(actualOutput[key]).To(Equal(vanilla[key]))
		}
	})

	It("adds the ducati release", func() {
		Expect(actualOutput).To(HaveKey("releases"))
		Expect(actualOutput["releases"]).To(ConsistOf(expectedOutput["releases"]))
	})

	XIt("returns the expected transformed manifest", func() {
		for _, key := range []string{"jobs", "properties", "releases", "resource_pools"} {
			Expect(actualOutput).To(HaveKey(key))
			Expect(actualOutput[key]).To(Equal(expectedOutput[key]))
		}
	})
})
