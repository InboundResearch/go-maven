#! /usr/bin/env perl

use strict;
use warnings FATAL => 'all';

# this is a maven wrapper intended to solve the problem that release builds don't actually deploy
# to the local nexus server using the maven release plugin

# global variables
our $mavenCommand = "mvn --quiet";
our $goPropertiesFileName = "go.properties";
our ($releaseBuildType, $snapshotBuildType) = ("", "-SNAPSHOT");

# function to do what 'chomp' should (but doesn't)
sub trim { my $s = shift; $s =~ s/^\s+|\s+$//g; return $s; };

sub setMavenVersion {
    my $newVersion = shift;
    my $newBuildType = shift;
    my $newVersionCommand = "$mavenCommand versions:set -DnewVersion=$newVersion$newBuildType -DgenerateBackupPoms=false -DprocessDependencies=false --non-recursive";
    #print STDERR "Exec ($versionCommand)\n";
    print STDERR "Setting build version ($newVersion$newBuildType)\n";
    system ($newVersionCommand) && die "Couldn't set version.\n";
}

sub checkin {
    my $message = shift;
    print STDERR "Check-in ($message).\n";
    system ("git add --all . && git commit -m 'go git ($message)' && git push origin HEAD;");
}

sub execute {
    my ($task, $command) = @_;
    print STDERR "Execute task ($task)\n";
    system ($command) && die ("($task) FAILED\n");
}

# get the version from maven, do this before checking options
my $mvnVersionCommand = "$mavenCommand -Dexec.executable='echo' -Dexec.args='\${project.version}' --non-recursive exec:exec";
my @mvnVersionCommandOutput = `$mvnVersionCommand`;
my $version = trim ($mvnVersionCommandOutput[0]);
$version =~ s/-SNAPSHOT$//;
print STDERR "Build at version ($version)\n";

# allowed options are: [--verbose]? [--clean]? [--notest]? [--git]? [build* | validate | package | install | deploy | release]
# * build is the default command
my $shouldClean = 0;
my $shouldTest = 1;
my $shouldCheckin = 0;
my $task = "build";
my %tasks; $tasks{$_} = $_ for ("build", "validate", "package", "install", "deploy", "release");
foreach (@ARGV) {
    my $arg = lc ($_);
    if ($arg eq "--verbose")  { $mavenCommand =~ s/ --quiet//; }
    elsif ($arg eq "--clean")  { $shouldClean = 1; }
    elsif ($arg eq "--notest") { $shouldTest = 0; }
    elsif ($arg eq "--git") { $shouldCheckin = 1; }
    elsif (exists $tasks{$arg}) { $task = $arg; }
    else { die "Unknown task ($arg).\n"; }
}

# figure out how to fulfill the task
if ($task eq "release") {
    # will be 0 if there are no changes...
    system ("git diff --quiet HEAD;") && die ("Please commit all changes before performing a release.\n");

    # ask the user to supply the new release version (default to the current version sans "SNAPSHOT"
    print "What is the release version (default [$version]): ";
    my $input = <STDIN>; $input = trim ($input);
    if (length ($input) > 0) { $version = $input; }

    # ask the user to supply the next development version (default to a dot-release)
    my ($major, $minor, $dot) = split (/\./, $version);
    my $nextDevelopmentVersion = "$major.$minor." . ($dot + 1);
    print "What will the new development version be (default [$nextDevelopmentVersion]): ";
    $input = <STDIN>; $input = trim ($input);
    $nextDevelopmentVersion = (length ($input) > 0) ? $input : $nextDevelopmentVersion;

    # configure testing by default, belittle the user if they want to skip it
    my $command = $mavenCommand;
    if ($shouldTest == 0) {
        print "WARNING - release without test, type 'y' to confirm (default [n]):";
        $input = <STDIN>; $input = lc (trim ($input));
        if ($input eq "y") {
            $command = "$command -Dmaven.test.skip=true";
        }
    }

    # set the version, and execute the release deployment build
    setMavenVersion($version, $releaseBuildType);
    execute ($task, "$command clean deploy");
    checkin("$version");
    print STDERR "Tag release ($version).\n";
    system ("git tag -a 'Release-$version' -m 'Release-$version';");

    # update the version to the development version and check it in
    setMavenVersion($nextDevelopmentVersion, $snapshotBuildType);
    checkin("$nextDevelopmentVersion");
} else {
    my $command = ($shouldClean == 1) ? "$mavenCommand clean" : "$mavenCommand";
    if ($task eq "build") {
        $command = ($shouldTest == 0) ? "$command compile" : "$command test";
    } else {
        $command = "$command $task";
        if ($shouldTest == 0) { $command = "$command -Dmaven.test.skip=true"; }
    }
    execute ($task, $command);
    if ($shouldCheckin) { checkin("CHECKPOINT - $version$snapshotBuildType"); }
}
