package main

// deployment script
var deploymentScript string = `#!/usr/bin/perl

my $User = $ARGV[0];
my $GoProjectDir = $ARGV[1];
my $Package = $ARGV[2];
my $Repository = $ARGV[3];
my $KeepRelease = $ARGV[4];

my $GoProjectSourceDir = $GoProjectDir . "/src";
my $DeploymentDir = $GoProjectSourceDir . "/" . $Package;
my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst)=localtime(time);
my $CurrentTime = sprintf ("%04d%02d%02d%02d%02d%02d", $year+1900,$mon+1,$mday,$hour,$min,$sec);
my @cmd;

my $UtilDir = "/home/" . $User . "/utils";
my $BackupDir = "/home/" . $User . "/backup";
my $BackupReleaseDir = $BackupDir . "/releases";
my $BackupCurrentReleaseDir = $BackupDir . "/releases/" . $CurrentTime;


# 0. directory check
if (! -d $DeploymentDir) {
    if (! makeDirIfNotExists($GoProjectDir)) {
        print "mkdir " . $GoProjectDir . " error\n";
        exit;
    }

	push @cmd, "export PATH=\$PATH:/usr/local/go/bin/ && "
                . "export GOPATH=\$HOME/go_project && "
                . "export PATH=\$PATH:\$GOPATH/bin && "
	            . "cd " . $GoProjectDir. " && "
                . "go get github.com/astaxie/beego && "
                . "go get github.com/beego/bee && "
                . "go get " . $Package;
}

if (! makeDirIfNotExists($UtilDir)) {
    print "mkdir " . $UtilDir . " error\n";
    exit;
}

if (! makeDirIfNotExists($BackupDir)) {
    print "mkdir " . $BackupDir . " error\n";
    exit;
}

if (! makeDirIfNotExists($BackupReleaseDir)) {
    print "mkdir " . $BackupReleaseDir . " error\n";
    exit;
}

if (! makeDirIfNotExists($BackupCurrentReleaseDir)) {
    print "mkdir " . $BackupCurrentReleaseDir . " error\n";
    exit;
}


# 1. backup current source
push @cmd, "cp -RPp " . $DeploymentDir . "/* " . $BackupCurrentReleaseDir;

# 2. git pull
push @cmd, "cd ". $DeploymentDir. " && "
            . "git checkout master && "
            . "git pull origin master && "
            . "git clean -q -d -x -f";

# 3. delete old backup
push @cmd, "ls -1dt " . $BackupReleaseDir . "/* | tail -n +" . $KeepRelease . " | xargs rm -rf";


foreach $idx (0..$#cmd) {
    #print "===========================================\n" . $cmd[$idx] . "\n";
    my $result = system($cmd[$idx]);
    #print $result . "\n\n";
}
print "backup dir : " . $BackupCurrentReleaseDir . "\n";


sub makeDirIfNotExists() {
    my ($dir, $recursive) = @_;

    if (! -d $dir) {
#        return mkdir $dir, 0755;
        if ($recursive) {
            system("mkdir -p " . $dir);
        } else {
            system("mkdir " . $dir);
        }

        return 1;
    } else {
        return 1;
    }

    return 0;
}

exit;
`

// run script
var runScript string = `#!/usr/bin/perl

my $GoProjectDir = $ARGV[0];
my $Package = $ARGV[1];
my $App = $ARGV[2];
my $Go = "/usr/local/go/bin/go";
my $DeploymentDir = $GoProjectDir . "/src/" . $Package;
my @cmd;

push @cmd, "/usr/bin/tmux kill-session -t $App";
push @cmd, "/usr/bin/tmux new-session -d -s $App \"export GOPATH=$GoProjectDir && "
            . "export PATH=\$PATH:$GoProjectDir/bin && "
            . "cd $DeploymentDir && $Go build && "
            . "$DeploymentDir/$App\"";

foreach $idx (0..$#cmd) {
    #print "===========================================\n$cmd[$idx]\n";
    my $result = system($cmd[$idx]);
    #print "$result\n";
    #
}

exit;
`
