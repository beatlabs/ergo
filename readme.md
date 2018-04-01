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

# Configuration
```
$ ergo -h
Usage of ergo:
  -baseBranch string
    	Base branch for the comparison. (default "master")
  -branches string
    	Comma separated list of branches
  -directory string
    	Location to store or retrieve from the repo
  -repoUrl string
    	git repo Url. ssh and https supported
  -skipFetch
    	Skip fetch. When set you may not be up to date with remote
```


# SSH access
For ssh access to repos make sure you have a running ssh-agent 
```
$ eval `ssh-agent`
Agent pid 4586
$ ssh-add 
```