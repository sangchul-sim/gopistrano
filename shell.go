package main

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

var runScript string = `#!/usr/bin/perl

my $go_project_path = $ARGV[0];
my $package = $ARGV[1];
my $app = $ARGV[2];
my $go = $go_project_path . "/usr/local/go/bin/go";
my @cmd;

push @cmd, "/usr/bin/tmux kill-session -t $app";
push @cmd, "/usr/bin/tmux new-session -d -s $app \"export GOPATH=$go_project_path && export PATH=$go_project_path/bin && cd $go_project_path/src/$package && /usr/local/go/bin/go install $package && $go_project_path/bin/$app\"";
push @cmd, "/usr/bin/tmux detach -s $app";

foreach $idx ( 0..$#cmd ){
    print "===========================================\n$cmd[$idx]\n";
    my $result = system($cmd[$idx]);
    print "$result\n";
}

exit;
`
