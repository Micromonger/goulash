package slackapi_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSlackapi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Slackapi Suite")
}
