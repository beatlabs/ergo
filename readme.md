[![Build Status](https://travis-ci.org/dbaltas/ergo.svg?branch=master)](https://travis-ci.org/dbaltas/ergo)
# ergo

Ergo (έργο), greek name for work, is a list of utilities for the daily developer workflow

![ergo sample output](ergo-sample-output.png)

# Actions
Currently one action is available

* Compare multiple branches in means of commits ahead and commits behind a base branch

# Run
```
$ go build -o ergo main.go
$ ./ergo -directory path-to-repo -branches 'qabranch,productionbranch,mybranch,yourbranch'
```

# Usage
```
Ergo aids to compare multiple branches.
On cases where deployment is done by pushing on a git branch:
* it can draft a github release,
* deploy on multiple branches and update the release notes with the time of release

Usage:
  ergo [flags]
  ergo [command]

Available Commands:
  deploy      Deploy base branch to target branches
  draft       Create a draft release on github comparing one target branch with the base branch
  help        Help about any command
  status      Print the status of branches compared to baseBranch
  version     Print the version of ergo

Flags:
      --baseBranch string   Base branch for the comparison.
      --branches string     Comma separated list of branches
      --detail              Print commits in detail
      --directory string    Location to store or retrieve from the repo (default ".")
  -h, --help                help for ergo
      --repoUrl string      git repo Url. ssh and https supported
      --skipFetch           Skip fetch. When set you may not be up to date with remote

Use "ergo [command] --help" for more information about a command.
```


# SSH access
For ssh access to repos make sure you have a running ssh-agent 
```
$ eval `ssh-agent`
Agent pid 4586
$ ssh-add 
```

# Config
Configuration is read from $HOME/.ergo.yaml

Sample config file
```yaml
generic:
  remote: origin
  release-repos: "ergo,periscope,ergo-functional-test-repo"
  base-branch: "master"
  status-branches: "develop,staging,master,release-es,release-gr"
  release-branches: "release-es,release-gr"
github:
  access-token: "access-token-goes-here"
  release-body-prefix: "### Added"
release:
  branch-map:
    release-gr: ":greece:"
    release-es: ":es:"
    ft-release-gr: ":greece:"
    ft-release-es: ":es:"
    ft-release-it: ":it:"
  on-deploy:
    body-branch-suffix-find: "-No-red.svg"
    body-branch-suffix-replace: "-green.svg"
repos:
  ergo-functional-test-repo:
    status-branches: "master,ft-release-gr,ft-release-es,ft-release-it"
    release-branches: "ft-release-es,ft-release-gr,ft-release-it"
```