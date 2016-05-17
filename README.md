Goulash
======
**Goulash** is a way to drive [Slack's](https://slack.com) API through [Slash Commands](https://api.slack.com/slash-commands).

Slack currently only has browser-based interface for inviting a single-channel guest or restricted account to a channel. [@levels.io](https://twitter.com/levelsio) did a [write-up](http://levels.io/slack-typeform-auto-invite-sign-ups/) on how he solved the problem of programmatically inviting people to a channel. This project takes that idea and implements a similar solution using [Slash Commands](https://api.slack.com/slash-commands) to perform specific levels of invites.

With this handler you can, through a Slash Command:

* Invite a single-channel guest to a channel or private group
* Invite a restricted account to a channel or private group
* Find out if an email address is associated with a single-channel guest or restricted account
* Configure an audit log for all invitations, successful and otherwise
* List the operations **Goulash** currently supports

## Requirements

**Goulash** is written in [Go](https://www.golang.org). See Go's [Getting Started documentation](https://www.golang.org/doc/install) if you don't already have it installed. **Goulash** has been tested in go1.4.2 darwin/amd64 and go1.4.2 linux/amd64.

Additionally, **Goulash** ships with [Godeps](https://github.com/tools/godep). If you plan to deploy it to [Cloud Foundry](http://pivotal.io/platform-as-a-service/pivotal-cloud-foundry), you'll need `godep`:

```
$ go get github.com/tools/godep
```

## Usage

### Get the source:

```
$ go get -t github.com/pivotalservices/goulash/...

// If you plan on deploying to Cloud Foundry
$ cd $GOPATH/src/github.com/pivotalservices/goulash
$ godep restore 
```

### Configure a new Slash Command on Slack:

Create a [new Slash Command](https://my.slack.com/services/new/slash-commands/).

This Slash Command should point at the endpoint **Goulash** is reachable at. The command you use will be what you set as the `SLACK_SLASH_COMMAND` environment variable. It will be displayed at times in response to a user, as in when directing them to request information from **Goulash** on what commands it supports.

### Set up your environment:

See descriptions below.

```
$ export SLACK_AUTH_TOKEN=xoxp-0000000000-0000000000-0000000000-000000
$ export SLACK_USER_ID=slackinviter
$ export SLACK_TEAM_NAME=slackteamname
$ export SLACK_SLASH_COMMAND="/slack-slash-command"
 
// Optional
$ export SLACK_AUDIT_LOG_CHANNEL_ID=G00000000
$ export UNINVITABLE_DOMAIN=example.com
$ export UNINVITABLE_DOMAIN_MESSAGE="Invites for this domain are forbidden."
$ export CONFIG_SERVICE_NAME=config-service
```

|Name|Required|Description|
|---|---|---|---|
|SLACK_AUTH_TOKEN|yes|The token of the user to use for inviting. This user must be an admin.
|SLACK_USER_ID|yes|The name of the user that will be doing the inviting. This should be the name associated with the token.
|SLACK_TEAM_NAME|yes|The name of the slack team that the invitations will be done for.
|SLACK_SLASH_COMMAND|yes|The name of the Slash Command associated with the **Goulash** endpoint.
|VCAP_APP_PORT|no|The port to listen on. Defaults to 8080.
|SLACK_AUDIT_LOG_CHANNEL_ID|no|ID of channel to use as audit log. See note below.
|UNINVITABLE_DOMAIN|no|Email addresses with this domain will be prohibited from being invited.
|UNINVITABLE_DOMAIN_MESSAGE|no|The message to show a user when they try to invite someone from an uninvitable domain.
|CONFIG_SERVICE_NAME|no|The name of a Cloud Foundry User-Provided Service that will provide the Slack auth token.

*You can get the ID of a channel by clicking its name from within Slack, and then choosing "Add a service integration". The ID is at the end of the URL.*

### Build and run Goulash:

```
$ cd $GOPATH/src/github.com/pivotalservices/goulash
$ go build -o goulash cmd/goulash/main.go
$ ./goulash
```

#### Running on Cloud Foundry
**Goulash** can be run on [Cloud Foundry](http://pivotal.io/platform-as-a-service/pivotal-cloud-foundry) without making any changes as it is already set up to listen on `VCAP_APP_PORT`. Simply set your environment via `cf set-env` with all of the required environment variables above (except `VCAP_APP_PORT`, of course), and `cf push` the app: 

```
$ cf push your-app-name -b https://github.com/cloudfoundry/go-buildpack.git
```

Don't have your own Cloud Foundry? Take a look at [Pivotal Web Services](http://run.pivotal.io).

#### Using a Cloud Foundry User-Provided Service for Slack Auth Token Storage

If you're using Cloud Foundry, you may choose to use a [User-Provided Service](http://docs.cloudfoundry.org/devguide/services/user-provided.html) to store your Slack auth token.

First, create the user-provided service with 'slack-auth-token' as the credential name and bind it to your app:

```
$ cf create-user-provided-service your-service-name -p "slack-auth-token"
$ cf bind-service your-app-name your-service-name
```

Create an environment variable, `CONFIG_SERVICE_NAME`, that contains the name of the service, and remove one that contained the auth token:

```
$ cf set-env your-app-name CONFIG_SERVICE_NAME=your-service-name
$ cf unset-env your-app-name SLACK_AUTH_TOKEN
```

Restage the app to apply the changes:

```
$ cf restage your-app-name
```

## Contributing
Pull requests are welcomed. Any PR must include test coverage.

```
$ cd $GOPATH/src/github.com/pivotalservices/goulash
$ ginkgo action confi handler slackapi
```

Before submitting a PR it is recommended to use [Concourse](http://concourse.ci) and its [`fly` tool](http://concourse.ci/fly-cli.html) to run `gometalinter` and `ginkgo` in an isolated environment: 

```
$ vagrant init concourse/lite
$ vagrant up
$ cd $GOPATH/src/github.com/pivotalservices/goulash
$ fly --target "http://192.168.100.4:8080" execute --config ci/unit.yml
```

## Maintainers
* [Kris Hicks](mailto:krishicks@gmail.com)
