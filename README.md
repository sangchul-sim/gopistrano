=======
gopistrano
==========

Automatic Deployment Tool for Beego + Tmux based on github.com/alanchavez88/gopistrano

## Requirements
* GoLang >= 1.3.3

## Installation
Run this if you want to install gopistrano binary

``` go
go get github.com/kyawmyintthein/gopistrano
```

That will compile and install gopistrano in your $GOPATH

To deploy a project, you need to create a Gopfile. A Gopfile is just a configuration file in plain-text that will contain the credentials to SSH into your server, and the path of the directory where you want to deploy your project to.

This is a sample Gopfile
```
username = yourusername
password = yourpassword
hostname = example.com
port = 22
repository = https://github.com/alanchavez88/theHarvester.git
keep_releases = 5
path = /home7/alanchav/gopistrano
use_sudo = false
ssh = sshkey_path
webserver_user = nobody
app_name = appname 

```
The file above will clone the git repository above into the path specified in the Gopfile. 

Currently gopistrano only supports git, other version controls will be added in the future. 

It also supports username and password authentication, PEM files and SSH Keys

It use https://tmux.github.io/ to run beego project in background. Tmux use current path to run the beego project in background.
But, need to do cross-compiles the beego project before deploy. Next version will support auto compile.


To deploy you have to run

``` sh
gopistrano deploy:setup

```

and then:

``` sh
gopistrano deploy
```

## Support

Need help with Gopistrano? email at kyawmyintthein2020@gmail.com.com
