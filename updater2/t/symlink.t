#!/usr/bin/perl

use strict;
use warnings;
use Test::More tests => 10;
use Log::Dispatch;
use MR::OnlineConf::Updater::Parameter;
use MR::OnlineConf::Updater::PerlMemory;

my $log = Log::Dispatch->new(outputs => [[ 'Screen', min_level => 'info', newline => 1 ]]);

my $tree = MR::OnlineConf::Updater::PerlMemory->new(log => $log);

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 1, name => '', path => '/', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 2, name => 'test', path => '/test', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 3, name => 'branch1', path => '/test/branch1', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 4, name => 'branch2', path => '/test/branch2', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 5, name => 'value', path => '/test/branch1/value', content_type => 'text/plain', data => 1, version => 1 },
    { id => 6, name => 'value', path => '/test/branch2/value', content_type => 'text/plain', data => 2, version => 1 },
    { id => 7, name => 'current', path => '/test/current', content_type => 'application/x-symlink', data => '/test/branch1', version => 1 },
    { id => 8, name => 'value', path => '/test/value', content_type => 'application/x-symlink', data => '/test/current/value', version => 1 },
    { id => 9, name => 'value2', path => '/test/value2', content_type => 'application/x-symlink', data => '/test/value', version => 1 },
    { id => 10, name => 'recursive', path => '/test/recursive', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 11, name => 'r1', path => '/test/recursive/r1', content_type => 'application/x-symlink', data => '/test/recursive/r2', version => 1 },
    { id => 12, name => 'r2', path => '/test/recursive/r2', content_type => 'application/x-symlink', data => '/test/recursive/r1', version => 1 },
    { id => 13, name => 'r3', path => '/test/recursive/r3', content_type => 'application/x-symlink', data => '/test/recursive/r3', version => 1 },
);
$tree->finalize();
is($tree->get('/test/value')->value, 1, "init: /test/value has correct value");
is($tree->get('/test/value2')->value, 1, "init: /test/value2 has correct value");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 7, name => 'current', path => '/test/current', content_type => 'application/x-symlink', data => '/test/branch2', version => 2 },
);
$tree->finalize();
is($tree->get('/test/value')->value, 2, "change /test/current to another symlink: /test/value has correct value");
is($tree->get('/test/value2')->value, 2, "change /test/current to another symlink: /test/value2 has correct value");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 7, name => 'current', path => '/test/current', content_type => 'application/x-null', data => undef, version => 3 },
    { id => 14, name => 'value', path => '/test/current/value', content_type => 'text/plain', data => 3, version => 1 },
);
$tree->finalize();
is($tree->get('/test/value')->value, 3, "change /test/current to real node: /test/value has correct value");
is($tree->get('/test/value2')->value, 3, "change /test/current to real node: /test/value2 has correct value");

$tree->delete(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 14, name => 'value', path => '/test/current/value', content_type => 'text/plain', data => 3, version => 2 },
);
$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 7, name => 'current', path => '/test/current', content_type => 'application/x-symlink', data => '/test/branch1', version => 4 },
);
$tree->finalize();
is($tree->get('/test/value')->value, 1, "change /test/current to symlink: /test/value has correct value");
is($tree->get('/test/value2')->value, 1, "change /test/current to symlink: /test/value2 has correct value");

ok(!eval { $tree->get('/test/recursive/r1') } && $@ =~ /^Recursion in symlink/, "recursive symlink");
ok(!eval { $tree->get('/test/recursive/r3') } && $@ =~ /^Recursion in symlink/, "symlink on itself");
