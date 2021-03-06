package main

// backup script
var backupScript string = `#!/usr/bin/perl
my $User = $ARGV[0];
my $GoProjectDir = $ARGV[1];
my $Package = $ARGV[2];
my $KeepRelease = $ARGV[3];
my $StartTimeStamp = time;
my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst) = localtime($StartTimeStamp);
my $StartTime = sprintf ("%04d-%02d-%02d %02d:%02d:%02d", $year+1900,$mon+1,$mday,$hour,$min,$sec);
my @cmd;
my $DeploymentDir = $GoProjectDir . "/src/" . $Package;
my $BackupDir = "/home/" . $User . "/backup";
(my $BackupTime = $StartTime) =~ s/[^\d]//g;
my $BackupPreviousReleaseDir = $BackupDir . "/" . $BackupTime;

if (! makeDirIfNotExists($BackupPreviousReleaseDir, 1)) {
    print "mkdir " . $BackupPreviousReleaseDir . " error\n";
    exit;
}

# 1. backup current source
push @cmd, "cp -RPp " . $DeploymentDir . "/* " . $BackupPreviousReleaseDir;

# 2. delete old backup
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
`

// deployment script
var deploymentScript string = `#!/usr/bin/perl

my $User = $ARGV[0];
my $GoProjectDir = $ARGV[1];
my $Package = $ARGV[2];
my $Repository = $ARGV[3];
my $DeploymentDir = $GoProjectDir . "/src/" . $Package;
my $StartTimeStamp = time;
my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst) = localtime($StartTimeStamp);
my $StartTime = sprintf ("%04d-%02d-%02d %02d:%02d:%02d", $year+1900,$mon+1,$mday,$hour,$min,$sec);
my $UtilsDir = "/home/" . $User . "/utils";
my @cmd;

open my $fh, ">>", "/tmp/debug.txt";
my $message = "deploymentScript start time: " . $StartTime . "\n";
print $message;
print $fh $message;

# 1. directory check
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

# 2. git pull
push @cmd, "cd ". $DeploymentDir. " && "
            . "git checkout master && "
            . "git pull origin master && "
            . "git clean -q -d -x -f";

foreach $idx (0..$#cmd) {
    #print "===========================================\n" . $cmd[$idx] . "\n";
    my $result = system($cmd[$idx]);
    #print $result . "\n\n";
}

sub makeDirIfNotExists() {
    my ($dir, $recursive) = @_;

    if (! -d $dir) {
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

my $FinishTimeStamp = time;
my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst) = localtime($FinishTimeStamp);
my $FinishTime = sprintf ("%04d-%02d-%02d %02d:%02d:%02d", $year+1900,$mon+1,$mday,$hour,$min,$sec);

my $DurationTimeStamp = $FinishTimeStamp - $StartTimeStamp;
my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst) = localtime($DurationTimeStamp);

$message = "deploymentScript finish time: " . $FinishTime . "\n";
print $message;
print $fh $message;

$message = "deploymentScript duration time: " . $min . "min " . $sec . "sec\n\n";
print $message;
print $fh $message;

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

my $StartTimeStamp = time;
my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst) = localtime($StartTimeStamp);
my $StartTime = sprintf ("%04d-%02d-%02d %02d:%02d:%02d", $year+1900,$mon+1,$mday,$hour,$min,$sec);

open my $fh, ">>", "/tmp/debug.txt";
my $message = "runScript start time: " . $StartTime . "\n";
print $message;
print $fh $message;

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

my $FinishTimeStamp = time;
my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst) = localtime($FinishTimeStamp);
my $FinishTime = sprintf ("%04d-%02d-%02d %02d:%02d:%02d", $year+1900,$mon+1,$mday,$hour,$min,$sec);

my $DurationTimeStamp = $FinishTimeStamp - $StartTimeStamp;
my ($sec,$min,$hour,$mday,$mon,$year,$wday,$yday,$isdst) = localtime($DurationTimeStamp);

$message = "runScript finish time: " . $FinishTime . "\n";
print $message;
print $fh $message;

$message = "runScript duration time: " . $min . "min " . $sec . "sec\n\n";
print $message;
print $fh $message;

close $fh;

exit;
`
