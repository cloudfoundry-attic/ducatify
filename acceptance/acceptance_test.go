package acceptance_test

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

func loadFixture(name string) ([]byte, map[string]interface{}) {
	bytes, err := ioutil.ReadFile(filepath.Join("fixtures", name+".yml"))
	Expect(err).NotTo(HaveOccurred())

	var data map[string]interface{}
	err = yaml.Unmarshal(bytes, &data)
	Expect(err).NotTo(HaveOccurred())

	return bytes, data
}

func findElementWithName(slice interface{}, name string) interface{} {
	for _, el := range slice.([]interface{}) {
		elAsMap := el.(map[interface{}]interface{})
		if elAsMap["name"] == name {
			return el
		}
	}
	Fail("missing expected element " + name)
	return nil
}

var _ = Describe("Manifest transformer", func() {
	var (
		cmd *exec.Cmd

		vanilla, expectedOutput, actualOutput map[string]interface{}

		actualBytes, expectedBytes []byte
	)

	BeforeEach(func() {
		_, vanilla = loadFixture("skeleton_vanilla")
		expectedBytes, expectedOutput = loadFixture("skeleton_transformed")

		cmd = exec.Command(binPath,
			"-diego", "fixtures/skeleton_vanilla.yml",
			"-cfCreds", "fixtures/cf_creds.yml",
		)

		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		actualBytes = session.Out.Contents()

		err = yaml.Unmarshal(actualBytes, &actualOutput)
		Expect(err).NotTo(HaveOccurred())
	})

	It("generates the expected output", func() {
		Expect(actualOutput).To(Equal(expectedOutput))
	})

	It("matches the yaml of the expected output", func() {
		Expect(actualBytes).To(MatchYAML(expectedBytes))
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

	It("leaves most of the deployment configuration intact", func() {
		for _, key := range []string{"name", "networks", "update", "resource_pools"} {
			Expect(actualOutput).To(HaveKey(key))
			Expect(expectedOutput[key]).To(Equal(vanilla[key]))
			Expect(actualOutput[key]).To(Equal(vanilla[key]))
		}
	})

	It("updates the releases", func() {
		Expect(actualOutput).To(HaveKey("releases"))
		Expect(actualOutput["releases"]).To(ConsistOf(expectedOutput["releases"]))
	})

	It("adds the ducati_db job", func() {
		Expect(actualOutput).To(HaveKey("jobs"))
		actualDBJob := findElementWithName(actualOutput["jobs"], "ducati_db")
		expectedDBJob := findElementWithName(expectedOutput["jobs"], "ducati_db")
		Expect(actualDBJob).To(Equal(expectedDBJob))
	})

	It("adds the ducati job to every cell", func() {
		Expect(actualOutput).To(HaveKey("jobs"))
		for i := 1; i <= 2; i++ {
			jobName := fmt.Sprintf("cell_z%d", i)
			actualJob := findElementWithName(actualOutput["jobs"], jobName)
			expectedJob := findElementWithName(expectedOutput["jobs"], jobName)
			Expect(actualJob).To(Equal(expectedJob))
		}
	})

	It("adds the ducati job to the colocated VMs", func() {
		Expect(actualOutput).To(HaveKey("jobs"))
		jobName := "colocated_z3"
		actualJob := findElementWithName(actualOutput["jobs"], jobName)
		expectedJob := findElementWithName(expectedOutput["jobs"], jobName)
		Expect(actualJob).To(Equal(expectedJob))
	})

	It("does not modify arbitrary jobs", func() {
		Expect(actualOutput).To(HaveKey("jobs"))
		actualJob := findElementWithName(actualOutput["jobs"], "brain_z2")
		expectedJob := findElementWithName(expectedOutput["jobs"], "brain_z2")
		Expect(actualJob).To(Equal(expectedJob))
	})

	It("adds new garden properties", func() {
		actualGardenProps := actualOutput["properties"].(map[interface{}]interface{})["garden"]
		expectedGardenProps := expectedOutput["properties"].(map[interface{}]interface{})["garden"]
		Expect(actualGardenProps).To(Equal(expectedGardenProps))
	})

	It("adds network_id property to nsync", func() {
		actualNsyncProps := actualOutput["properties"].(map[interface{}]interface{})["diego"].(map[interface{}]interface{})["nsync"]
		expectedNsyncProps := expectedOutput["properties"].(map[interface{}]interface{})["diego"].(map[interface{}]interface{})["nsync"]
		Expect(actualNsyncProps).To(Equal(expectedNsyncProps))
	})

	It("adds ducati properties", func() {
		actualDucatiProps := actualOutput["properties"].(map[interface{}]interface{})["ducati"]
		expectedDucatiProps := expectedOutput["properties"].(map[interface{}]interface{})["ducati"]
		Expect(actualDucatiProps).To(Equal(expectedDucatiProps))
	})
})
