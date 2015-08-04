#!/usr/bin/perl

use strict;
use warnings;
use Test::More tests => 6;
use Log::Dispatch;
use MR::OnlineConf::Updater::Parameter;
use MR::OnlineConf::Updater::PerlMemory;

my $log = Log::Dispatch->new(outputs => [[ 'Screen', min_level => 'info', newline => 1 ]]);

my $tree = MR::OnlineConf::Updater::PerlMemory->new(log => $log);

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 1, name => '', path => '/', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 2, name => 'test', path => '/test', content_type => 'application/x-null', data => undef, version => 1 },
);
$tree->finalize();

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 3, name => 'value', path => '/test/value', content_type => 'text/plain', data => 3, version => 1 },
);
$tree->finalize();
is($tree->get('/test/value')->value, 3, "insert: correct value");
is($tree->get('/test/value')->content_type, 'text/plain', "insert: correct content type");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 3, name => 'value', path => '/test/value', content_type => 'text/plain', data => 4, version => 2 },
);
$tree->finalize();
is($tree->get('/test/value')->value, 4, "update value");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 3, name => 'value', path => '/test/value', content_type => 'application/json', data => '[]', version => 3 },
);
$tree->finalize();
is(ref $tree->get('/test/value')->value, 'ARRAY', "update content type and value: correct value");
is($tree->get('/test/value')->content_type, 'application/json', "update content type and value: correct content type");

$tree->delete(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 3, name => 'value', path => '/test/value', content_type => 'application/json', data => '[]', version => 4 },
);
$tree->finalize();
ok(!$tree->get('/test/value'), "delete: not exists");
