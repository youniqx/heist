package e2e_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Random API", func() {
	When("A random string is required", func() {
		It("Can generate a random string with a given length", func() {
			result, err := vaultAPI.GenerateRandomString(32)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(32))
		})

		It("Errors if the given length is 0", func() {
			result, err := vaultAPI.GenerateRandomString(0)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})

	When("A random byte array is required", func() {
		It("Can generate a byte array with a given length", func() {
			result, err := vaultAPI.GenerateRandomBytes(32)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(32))
		})

		It("Errors if the given length is 0", func() {
			result, err := vaultAPI.GenerateRandomBytes(0)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})
})
