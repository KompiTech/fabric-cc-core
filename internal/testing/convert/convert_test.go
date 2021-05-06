package convert_test

import (
	"testing"

	convert2 "github.com/KompiTech/fabric-cc-core/v2/internal/testing/convert"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestState(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "State suite")
}

var _ = Describe(`Convert`, func() {

	It(`Bool`, func() {
		bTrue, err := convert2.ToBytes(true)
		Expect(err).NotTo(HaveOccurred())
		Expect(bTrue).To(Equal([]byte(`true`)))

		bFalse, err := convert2.ToBytes(false)
		Expect(err).NotTo(HaveOccurred())
		Expect(bFalse).To(Equal([]byte(`false`)))

		eTrue, err := convert2.FromBytes(bTrue, convert2.TypeBool)
		Expect(err).NotTo(HaveOccurred())
		Expect(eTrue.(bool)).To(Equal(true))

		eFalse, err := convert2.FromBytes(bFalse, convert2.TypeBool)
		Expect(err).NotTo(HaveOccurred())
		Expect(eFalse.(bool)).To(Equal(false))
	})

	It(`String`, func() {
		const MyStr = `my-string`
		bStr, err := convert2.ToBytes(MyStr)
		Expect(err).NotTo(HaveOccurred())
		Expect(bStr).To(Equal([]byte(MyStr)))

		eStr, err := convert2.FromBytes(bStr, convert2.TypeString)
		Expect(err).NotTo(HaveOccurred())
		Expect(eStr.(string)).To(Equal(MyStr))
	})

	It(`Nil`, func() {
		bNil, err := convert2.ToBytes(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(bNil).To(Equal([]byte{}))
	})

})
