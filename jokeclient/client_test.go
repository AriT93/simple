package jokeclient_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/AriT93/ai-agent/jokeclient"
)

var _ = Describe("Joke Client", func() {
	var (
		client *jokeclient.Client
		server *httptest.Server
	)

	BeforeEach(func() {
		client = jokeclient.NewClient()
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("NewClient", func() {
		It("should create a new client with default values", func() {
			Expect(client.BaseURL).To(Equal("https://v2.jokeapi.dev/joke"))
			Expect(client.Timeout).To(Equal(5 * time.Second))
		})
	})

	Describe("FetchJoke", func() {
		Context("with a mock server", func() {
			BeforeEach(func() {
				// Create a test server that returns a predefined joke
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					
					// Debug the request path and query
					fmt.Printf("Test server received request: %s with query: %v\n", r.URL.Path, r.URL.Query())
					
					// Check for programming joke request
					if strings.Contains(r.URL.Path, "Programming") {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{
							"error": false,
							"category": "Programming",
							"type": "single",
							"joke": "Why do programmers prefer dark mode? Because light attracts bugs!",
							"flags": {
								"nsfw": false,
								"religious": false,
								"political": false,
								"racist": false,
								"sexist": false,
								"explicit": false
							},
							"id": 1,
							"safe": true,
							"lang": "en"
						}`))
						return
					} 
					
					// Check for twopart joke request
					if r.URL.Query().Get("type") == "twopart" {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{
							"error": false,
							"category": "Misc",
							"type": "twopart",
							"setup": "What's the best thing about Switzerland?",
							"delivery": "I don't know, but the flag is a big plus!",
							"flags": {
								"nsfw": false,
								"religious": false,
								"political": false,
								"racist": false,
								"sexist": false,
								"explicit": false
							},
							"id": 2,
							"safe": true,
							"lang": "en"
						}`))
						return
					}
					
					// Default response
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"error": false,
						"category": "Misc",
						"type": "single",
						"joke": "Default joke response",
						"flags": {
							"nsfw": false,
							"religious": false,
							"political": false,
							"racist": false,
							"sexist": false,
							"explicit": false
						},
						"id": 3,
						"safe": true,
						"lang": "en"
					}`))
				}))
				
				// Override the client's BaseURL to use our test server
				client.BaseURL = server.URL
			})

			It("should fetch a programming joke", func() {
				joke, err := client.FetchJoke("Tell me a programming joke")
				
				Expect(err).NotTo(HaveOccurred())
				Expect(joke).To(Equal("Why do programmers prefer dark mode? Because light attracts bugs!"))
			})

			It("should fetch a twopart joke", func() {
				joke, err := client.FetchJoke("Tell me a twopart joke")
				
				Expect(err).NotTo(HaveOccurred())
				Expect(joke).To(ContainSubstring("What's the best thing about Switzerland?"))
				Expect(joke).To(ContainSubstring("I don't know, but the flag is a big plus!"))
			})

			It("should handle default joke requests", func() {
				joke, err := client.FetchJoke("Tell me any joke")
				
				Expect(err).NotTo(HaveOccurred())
				Expect(joke).To(Equal("Default joke response"))
			})
		})

		Context("with error conditions", func() {
			BeforeEach(func() {
				// Create a test server that returns errors
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Printf("Error test server received request: %s\n", r.URL.Path)
					
					if strings.Contains(r.URL.Path, "Error") {
						w.WriteHeader(http.StatusInternalServerError)
						return
					} 
					
					if strings.Contains(r.URL.Path, "BadJSON") {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{invalid json`))
						return
					} 
					
					if strings.Contains(r.URL.Path, "UnknownType") {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{
							"error": false,
							"category": "Misc",
							"type": "unknown",
							"flags": {},
							"id": 4,
							"safe": true,
							"lang": "en"
						}`))
						return
					}
					
					// Default response for any other path
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{}`))
				}))
				
				// Override the client's BaseURL to use our test server
				client.BaseURL = server.URL
			})

			It("should handle server errors", func() {
				_, err := client.FetchJoke("Error")
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API request failed with status code: 500"))
			})

			It("should handle invalid JSON", func() {
				_, err := client.FetchJoke("BadJSON")
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to unmarshal JSON"))
			})

			It("should handle unknown joke types", func() {
				_, err := client.FetchJoke("UnknownType")
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unknown joke type"))
			})
		})
	})
})
