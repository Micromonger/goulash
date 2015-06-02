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
$ go get github.com/pivotalservices/goulash.git

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
$ cf push my_app -b https://github.com/cloudfoundry/go-buildpack.git
```

Don't have your own Cloud Foundry? Take a look at [Pivotal Web Services](http://run.pivotal.io).

## Contributing
Pull requests are welcomed. Any PR must include test coverage and pass [gometalinter](https://github.com/alecthomas/gometalinter).

```
$ cd $GOPATH/src/github.com/pivotalservices/goulash
$ gometalinter --exclude=(Godeps|fakes) ./...
$ ginkgo *
```

Before submitting a PR it is recommended to use [Concourse](http://concourse.ci) and its [`fly` tool](http://concourse.ci/fly-cli.html) to run `gometalinter` and `ginkgo` in an isolated environment: 

```
$ vagrant init concourse/lite
$ vagrant up
$ cd $GOPATH/src/github.com/pivotalservices/goulash
$ fly -c ci/unit.yml
```

## Maintainers
* [Kris Hicks](mailto:krishicks@gmail.com)
