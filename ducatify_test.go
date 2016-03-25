package ducatify_test

import (
	"github.com/cloudfoundry-incubator/ducatify"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transform", func() {
	var (
		manifest    map[string]interface{}
		transformer *ducatify.Transformer
	)

	BeforeEach(func() {
		transformer = &ducatify.Transformer{}

		manifest = map[string]interface{}{
			"releases": []interface{}{
				map[string]interface{}{"name": "some-release", "version": "latest"},
				map[string]interface{}{"name": "another-release", "version": "whatever"},
			},
		}
	})
	It("adds the ducati release", func() {
		err := transformer.Transform(manifest)
		Expect(err).NotTo(HaveOccurred())

		Expect(manifest).To(HaveKey("releases"))
		releases := manifest["releases"]
		Expect(releases).To(ContainElement(
			map[string]interface{}{"name": "some-release", "version": "latest"},
		))
		Expect(releases).To(ContainElement(
			map[string]interface{}{"name": "another-release", "version": "whatever"},
		))
		Expect(releases).To(ContainElement(
			map[string]interface{}{"name": "ducati", "version": "latest"},
		))
	})

})
