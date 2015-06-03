package config_test

import (
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/pivotalservices/goulash/config"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnvConfig", func() {
	Describe("SlackAuthToken", func() {
		AfterEach(func() {
			os.Unsetenv("GOULASH_TEST_CONFIG_SERVICE_NAME")
			os.Unsetenv("GOULASH_TEST_SLACK_AUTH_TOKEN")
		})

		It("returns a service-based audit log channel id", func() {
			err := os.Setenv("GOULASH_TEST_CONFIG_SERVICE_NAME", "config-service-name")
			Ω(err).ShouldNot(HaveOccurred())

			env := []string{
				`VCAP_APPLICATION={}`,
				`VCAP_SERVICES={
					"user-provided":[{
						"name":"config-service-name",
						"credentials":{"slack-auth-token":"slack-auth-token-value"}
					}]
				}`,
			}

			app, err := cfenv.New(cfenv.Env(env))
			Ω(err).ShouldNot(HaveOccurred())

			c := config.NewEnvConfig(
				app,
				"GOULASH_TEST_CONFIG_SERVICE_NAME",
				"",
				"slack-auth-token",
				"",
				"",
				"",
				"",
				"",
			)

			Ω(c.SlackAuthToken()).Should(Equal("slack-auth-token-value"))
		})

		It("returns an env-based audit log channel id", func() {
			app, err := cfenv.New(cfenv.Env([]string{`VCAP_APPLICATION={}`, `VCAP_SERVICES={}`}))
			Ω(err).ShouldNot(HaveOccurred())
			c := config.NewEnvConfig(app, "", "", "GOULASH_TEST_SLACK_AUTH_TOKEN", "", "", "", "", "")
			err = os.Setenv("GOULASH_TEST_SLACK_AUTH_TOKEN", "slack-auth-token-value")
			Ω(err).ShouldNot(HaveOccurred())

			Ω(c.SlackAuthToken()).Should(Equal("slack-auth-token-value"))
		})
	})
})
