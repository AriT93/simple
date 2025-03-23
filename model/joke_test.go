package model_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/AriT93/ai-agent/model"
)

var _ = Describe("Joke Model", func() {
	Describe("JokeResponse", func() {
		It("should unmarshal a single joke correctly", func() {
			jsonData := []byte(`{
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
			}`)

			var joke model.JokeResponse
			err := json.Unmarshal(jsonData, &joke)

			Expect(err).NotTo(HaveOccurred())
			Expect(joke.Error).To(BeFalse())
			Expect(joke.Category).To(Equal("Programming"))
			Expect(joke.Type).To(Equal("single"))
			Expect(joke.Joke).To(Equal("Why do programmers prefer dark mode? Because light attracts bugs!"))
			Expect(joke.ID).To(Equal(1))
			Expect(joke.Safe).To(BeTrue())
			Expect(joke.Lang).To(Equal("en"))
			Expect(joke.Flags.Nsfw).To(BeFalse())
			Expect(joke.Flags.Religious).To(BeFalse())
			Expect(joke.Flags.Political).To(BeFalse())
			Expect(joke.Flags.Racist).To(BeFalse())
			Expect(joke.Flags.Sexist).To(BeFalse())
			Expect(joke.Flags.Explicit).To(BeFalse())
		})

		It("should unmarshal a twopart joke correctly", func() {
			jsonData := []byte(`{
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
			}`)

			var joke model.JokeResponse
			err := json.Unmarshal(jsonData, &joke)

			Expect(err).NotTo(HaveOccurred())
			Expect(joke.Error).To(BeFalse())
			Expect(joke.Category).To(Equal("Misc"))
			Expect(joke.Type).To(Equal("twopart"))
			Expect(joke.Setup).To(Equal("What's the best thing about Switzerland?"))
			Expect(joke.Delivery).To(Equal("I don't know, but the flag is a big plus!"))
			Expect(joke.ID).To(Equal(2))
			Expect(joke.Safe).To(BeTrue())
			Expect(joke.Lang).To(Equal("en"))
		})
	})
})
