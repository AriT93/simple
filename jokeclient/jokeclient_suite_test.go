package jokeclient_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJokeClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "JokeClient Suite")
}
