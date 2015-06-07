package config_test

import (
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnvConfig", func() {
	Describe("SlackAuthToken", func() {
		var logger lager.Logger

		BeforeEach(func() {
			logger = lager.NewLogger("testlogger")
		})

		AfterEach(func() {
			os.Unsetenv("GOULASH_TEST_CONFIG_SERVICE_NAME")
			os.Unsetenv("GOULASH_TEST_SLACK_AUTH_TOKEN")
		})

		It("returns a service-based audit log channel id", func() {
			err := os.Setenv("GOULASH_TEST_CONFIG_SERVICE_NAME", "config-service-name")
			Ω(err).ToNot(HaveOccurred())

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
			Ω(err).ToNot(HaveOccurred())

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
				logger,
			)

			Ω(c.SlackAuthToken()).To(Equal("slack-auth-token-value"))
		})

		It("returns an env-based audit log channel id", func() {
			app, err := cfenv.New(cfenv.Env([]string{`VCAP_APPLICATION={}`, `VCAP_SERVICES={}`}))
			Ω(err).ToNot(HaveOccurred())
			c := config.NewEnvConfig(app, "", "", "GOULASH_TEST_SLACK_AUTH_TOKEN", "", "", "", "", "", logger)
			err = os.Setenv("GOULASH_TEST_SLACK_AUTH_TOKEN", "slack-auth-token-value")
			Ω(err).ToNot(HaveOccurred())

			Ω(c.SlackAuthToken()).To(Equal("slack-auth-token-value"))
		})
	})
})
