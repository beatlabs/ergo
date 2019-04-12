# ergo [![Build Status](https://travis-ci.com/thebeatapp/ergo.svg?branch=master)](https://travis-ci.com/thebeatapp/ergo)

Ergo (έργο), greek name for work, is a list of utilities for the daily release workflow.

## Installation
```
$ go get github.com/thebeatapp/ergo
```

## Usage

```
Usage:
  ergo [flags]
  ergo [command]

Available Commands:
  deploy      Deploy base branch to target branches
  draft       Create a draft release [github]
  help        Help about any command
  status      the status of branches compared to base branch
  tag         Create a tag on branch
  version     the version of ergo

Flags:
      --base string       Base branch for the comparison.
      --branches string   Comma separated list of branches
  -h, --help              help for ergo
      --owner string
      --path string       Location to store or retrieve from the repo (default ".")
      --repo string
```

## CLI commands

#### Status

Getting the status of remote repo branches compared to a base branch.

```bash
ergo status \
--owner dbaltas \
--repo ergo \
--base master \
--branches stable,testsuite/baseNew,testsuite/base,testsuite/featureA,testsuite/featureB,testsuite/featureC
```

![ergo sample output](static/ergo-status.png)

#### Draft

Create a draft release having description of the commit diff. It will try to increment the last found tag version.

```bash
ergo draft \
--owner dbaltas \
--repo ergo \
--base master \
--branches release-gr,release-it
```

#### Deploy

Push the release tag into the release branches (and update the release body accordingly). You need to have published the draft release first.

```bash
ergo deploy \
--owner dbaltas \
--repo ergo \
--releaseInterval 15m \
--branches release-pe,release-mx,release-co,release-cl,release-gr
```

## Github Access
To communicate with github you will need a [personal access token](https://github.com/settings/tokens) added on the configuration file as `access-token` on github

## Configuration
Configuration is read from $HOME/.ergo.yaml

You have to use this in order to:
- Add your github access token
- Provide ergo with defaults. In the CLI commands you may skip some of the parameters in case that there are defaults values set.
- Information about the draft release body and what will change at the time of the release.

[Sample config file](.ergo.yml.dist)
