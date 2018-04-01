# ergo

Ergo (έργο), greek name for work, is a list of utilities for the daily developer workflow

# Actions
Currently one action is available

* Compare multiple branches in means of commits ahead and commits behind a base branch

# Run
```
$ go build -o ergo main.go
$ ./ergo -directory path-to-repo --skipFetch -branches 'qabranch,productionbranch,mybranch,yourbranch'
```

# SSH access
For ssh access to repos make sure you have a running ssh-agent 
```
$ eval `ssh-agent`
Agent pid 4586
$ ssh-add 
```