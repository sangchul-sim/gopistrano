package main

/*
var deploymentScript string = `#!/bin/bash
# comment line below if you want quiet output
#set -x

DEPLOYMENT_PATH=$1
REPOSITORY=$2
KEEP_RELEASES=$3
APPNAME=$4
# variable init
CUR_TIMESTAMP="$(date +'%Y%m%d%H%M%S')"

# update code base with remote_cache strategy
if [ -d "$DEPLOYMENT_PATH/shared/cached-copy" ]
then
	cd "$DEPLOYMENT_PATH/shared/cached-copy"
	git checkout master
	git fetch -q origin
	git fetch --tags -q origin
	git rev-list --max-count=1 HEAD | xargs git reset -q --hard
	git clean -q -d -x -f;
else
	git clone -q $REPOSITORY "$DEPLOYMENT_PATH/shared/cached-copy"
	cd "$DEPLOYMENT_PATH/shared/cached-copy"
	git rev-list --max-count=1 HEAD | xargs git checkout -q -b deploy
fi

cp -RPp "$DEPLOYMENT_PATH/shared/cached-copy" "$DEPLOYMENT_PATH/releases/$CUR_TIMESTAMP"
git rev-list --max-count=1 HEAD > "$DEPLOYMENT_PATH/releases/$CUR_TIMESTAMP/REVISION"
chmod -R g+w "$DEPLOYMENT_PATH/releases/$CUR_TIMESTAMP"

rm -f "$DEPLOYMENT_PATH/current" &&  ln -s "$DEPLOYMENT_PATH/releases/$CUR_TIMESTAMP" "$DEPLOYMENT_PATH/current"
ls -1dt "$DEPLOYMENT_PATH/releases" | tail -n +$KEEP_RELEASES |  xargs rm -rf
`
*/

var deploymentScript string = `#!/usr/bin/perl

my $GoProjectPath = $ARGV[0];
my $Package = $ARGV[1];
my $App = $ARGV[2];
my $Repository = $ARGV[3];
my $KeepRelease = $ARGV[4];
my $DeploymentPath = $GoProjectPath . "/src/" . $Package;
my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst)=localtime(time);
my $CurrentTime = sprintf ("%04d%02d%02d%02d%02d%02d", $year+1900,$mon+1,$mday,$hour,$min,$sec);
my @cmd;

# update code base with remote_cache strategy
if (-d "$DeploymentPath/shared/cached-copy") {
	my $command = "cd $DeploymentPath/shared/cached-copy && "
	. "git checkout master && "
	. "git pull origin master && "

#	. "git fetch -q origin && "
#	. "git fetch --tags -q origin && "
#	. "git rev-list --max-count=1 HEAD | xargs git reset -q --hard && "

	. "git clean -q -d -x -f";

	push @cmd, $command;
} else {
	my $command = "git clone -q $Repository $DeploymentPath/shared/cached-copy && "
	. "cd $DeploymentPath/shared/cached-copy && "
	. "git rev-list --max-count=1 HEAD | xargs git checkout -q -b deploy";

	push @cmd, $command;
}

push @cmd, "cp -RPp $DeploymentPath/shared/cached-copy $DeploymentPath/releases/$CurrentTime";
push @cmd, "cd $DeploymentPath/shared/cached-copy && git rev-list --max-count=1 HEAD > $DeploymentPath/releases/$CurrentTime/REVISION";
push @cmd, "chmod -R g+w $DeploymentPath/releases/$CurrentTime";

push @cmd, "rm -f $DeploymentPath/current && ln -s $DeploymentPath/releases/$CurrentTime $DeploymentPath/current";
push @cmd, "ls -1dt $DeploymentPath/releases | tail -n +$KeepRelease | xargs rm -rf";

foreach $idx (0..$#cmd) {
    print "===========================================\n$cmd[$idx]\n";
    my $result = system($cmd[$idx]);
    print "$result\n\n";
}
exit;
`

var runScript string = `#!/usr/bin/perl

my $GoProjectPath = $ARGV[0];
my $Package = $ARGV[1];
my $App = $ARGV[2];
my $Go = "/usr/local/go/bin/go";
my $DeploymentPath = $GoProjectPath . "/src/" . $Package;
my $CurrentSrcPath = $DeploymentPath . "/current";
my @cmd;

push @cmd, "/usr/bin/tmux kill-session -t $App";
push @cmd, "/usr/bin/tmux new-session -d -s $App \"export GOPATH=$GoProjectPath && export PATH=\$PATH:$GoProjectPath/bin && cd $CurrentSrcPath && $Go build && $CurrentSrcPath/current\"";
push @cmd, "/usr/bin/tmux detach -s $App";

foreach $idx (0..$#cmd) {
    print "===========================================\n$cmd[$idx]\n";
    my $result = system($cmd[$idx]);
    print "$result\n";
}

exit;
`
