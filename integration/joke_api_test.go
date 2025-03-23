package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/AriT93/ai-agent/jokeclient"
)

var _ = Describe("Joke API Integration", func() {
	var client *jokeclient.Client

	BeforeEach(func() {
		client = jokeclient.NewClient()
	})

	// These tests hit the real API
	// Use --focus="Live API Tests" to run them
	Describe("Live API Tests", func() {
		It("should fetch a programming joke", func() {
			joke, err := client.FetchJoke("Tell me a programming joke")
			
			Expect(err).NotTo(HaveOccurred())
			Expect(joke).NotTo(BeEmpty())
		})

		It("should fetch a twopart joke", func() {
			joke, err := client.FetchJoke("Tell me a twopart joke")
			
			Expect(err).NotTo(HaveOccurred())
			Expect(joke).NotTo(BeEmpty())
			// A twopart joke should contain a newline
			Expect(joke).To(ContainSubstring("\n\n"))
		})

		It("should respect category requests", func() {
			joke, err := client.FetchJoke("Tell me a Christmas joke")
			
			Expect(err).NotTo(HaveOccurred())
			Expect(joke).NotTo(BeEmpty())
		})

		It("should handle blacklist flags", func() {
			joke, err := client.FetchJoke("Tell me a joke but nothing nsfw or political")
			
			Expect(err).NotTo(HaveOccurred())
			Expect(joke).NotTo(BeEmpty())
		})
	})
})
