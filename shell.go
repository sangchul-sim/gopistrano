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

my $GO_PROJECT_PATH = $ARGV[0];
my $PACKAGE = $ARGV[1];
my $APP = $ARGV[2];
my @cmd;

my $bee = $GO_PROJECT_PATH . "/bin/bee";

push @cmd, "/usr/bin/tmux kill-session -t $APP";

# OK
push @cmd, "/usr/bin/tmux new-session -d -s $APP \"export GOPATH=$GO_PROJECT_PATH && export PATH=$GO_PROJECT_PATH/bin && cd $GO_PROJECT_PATH/src/$PACKAGE && /usr/local/go/bin/go install $PACKAGE && $GO_PROJECT_PATH/bin/$APP\"";
#
#push @cmd, "/usr/bin/tmux new-session -d -s $APP \"export GOPATH=/home/namu/go_project && export PATH=\$PATH:\$GOPATH/bin && cd $GO_PROJECT_PATH/src/$PACKAGE && $GO_PROJECT_PATH/bin/bee run $PACKAGE\"";

#push @cmd, "export GOPATH=/home/namu/go_project && export PATH=/home/namu/go_project/bin && cd $GO_PROJECT_PATH/src/$PACKAGE && echo $GO_PROJECT_PATH/src/$PACKAGE > result.txt && /usr/local/go/bin/go install $PACKAGE >> /tmp/build.txt && $GO_PROJECT_PATH/bin/$APP";
#push @cmd, "sleep 2";
#push @cmd, "/usr/bin/tmux detach -s $APP";

foreach $idx ( 0..$#cmd ){
    print "===========================================\n$cmd[$idx]\n";
    my $result = system($cmd[$idx]);
    print "$result\n";
}

exit;
`
