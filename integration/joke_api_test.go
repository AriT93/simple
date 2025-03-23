package integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/AriT93/ai-agent/jokeclient"
)

var _ = Describe("Joke API Integration", func() {
	var client *jokeclient.Client

	BeforeEach(func() {
		client = jokeclient.NewClient()
		// Increase timeout for integration tests
		client.Timeout = 10 * time.Second
	})

	// These tests hit the real API and may be flaky
	// Use --skip="Live API Tests" to skip them if needed
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
			// Use a simpler request with fewer blacklist flags to reduce chance of timeout
			joke, err := client.FetchJoke("Tell me a joke but nothing nsfw")
			
			Expect(err).NotTo(HaveOccurred())
			Expect(joke).NotTo(BeEmpty())
		})
	})
})
