package main

// deployment script
var deploymentScript string = `#!/usr/bin/perl

my $User = $ARGV[0];
my $GoProjectDir = $ARGV[1];
my $Package = $ARGV[2];
my $Repository = $ARGV[3];
my $KeepRelease = $ARGV[4];
my $DeploymentDir = $GoProjectDir . "/src/" . $Package;
my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst)=localtime(time);
my $CurrentTime = sprintf ("%04d%02d%02d%02d%02d%02d", $year+1900,$mon+1,$mday,$hour,$min,$sec);
my $UtilsDir = "/home/" . $User . "/utils";
my $BackupDir = "/home/" . $User . "/backup";
my $BackupPreviousReleaseDir = $BackupDir . "/" . $CurrentTime;
my @cmd;

print "deploymentScript current time:" . $CurrentTime . "\n";

open my $fh, ">>", "/tmp/debug.txt";
print $fh "deploymentScript start time:" . $CurrentTime . "\n";

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

if (! makeDirIfNotExists($UtilsDir)) {
    print "mkdir " . $UtilsDir . " error\n";
    exit;
}

if (! makeDirIfNotExists($BackupPreviousReleaseDir, 1)) {
    print "mkdir " . $BackupPreviousReleaseDir . " error\n";
    exit;
}


# 1. backup current source
push @cmd, "cp -RPp " . $DeploymentDir . "/* " . $BackupPreviousReleaseDir;

# 2. git pull
push @cmd, "cd ". $DeploymentDir. " && "
            . "git checkout master && "
            . "git pull origin master && "
            . "git clean -q -d -x -f";

# 3. delete old backup
push @cmd, "ls -1dt " . $BackupDir . "/* | tail -n +" . $KeepRelease . " | xargs rm -rf";


foreach $idx (0..$#cmd) {
    #print "===========================================\n" . $cmd[$idx] . "\n";
    my $result = system($cmd[$idx]);
    #print $result . "\n\n";
}
print "backup dir : " . $BackupPreviousReleaseDir . "\n";


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

my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst)=localtime(time);
my $CurrentTime = sprintf ("%04d%02d%02d%02d%02d%02d", $year+1900,$mon+1,$mday,$hour,$min,$sec);
print $fh "deploymentScript end time:" . $CurrentTime . "\n\n";
close $fh;

exit;
`

// restart script
var restartScript string = `#!/usr/bin/perl

my $GoProjectDir = $ARGV[0];
my $Package = $ARGV[1];
my $App = $ARGV[2];
my $Go = "/usr/local/go/bin/go";
my $DeploymentDir = $GoProjectDir . "/src/" . $Package;
my @cmd;

my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst)=localtime(time);
my $CurrentTime = sprintf ("%04d%02d%02d%02d%02d%02d", $year+1900,$mon+1,$mday,$hour,$min,$sec);
print "runScript current time:" . $CurrentTime . "\n";

open my $fh, ">>", "/tmp/debug.txt";
print $fh "runScript start time:" . $CurrentTime . "\n";

push @cmd, "/usr/bin/tmux kill-session -t " . $App;
push @cmd, "/usr/bin/tmux new-session -d -s " . $App . " "
            . "\"export GOPATH=" . $GoProjectDir . " && "
            . "export PATH=\$PATH:" . $GoProjectDir . "/bin && "
            . "cd " . $DeploymentDir . " && "
            . $Go . " build && "
            . $DeploymentDir . "/" . $App . "\"";

foreach $idx (0..$#cmd) {
    #print "===========================================\n" . $cmd[$idx] . "\n";
    my $result = system($cmd[$idx]);
    #print "$result\n";
    #
}

my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst)=localtime(time);
my $CurrentTime = sprintf ("%04d%02d%02d%02d%02d%02d", $year+1900,$mon+1,$mday,$hour,$min,$sec);
print $fh "runScript end time:" . $CurrentTime . "\n\n";
close $fh;

exit;
`
