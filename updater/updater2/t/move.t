#!/usr/bin/perl

use strict;
use warnings;
use Test::More tests => 24;
use Log::Dispatch;
use MR::OnlineConf::Updater::Parameter;
use MR::OnlineConf::Updater::PerlMemory;

my $log = Log::Dispatch->new(outputs => [[ 'Screen', min_level => 'info', newline => 1 ]]);

my $tree = MR::OnlineConf::Updater::PerlMemory->new(log => $log);

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 1, name => '', path => '/', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 2, name => 'test', path => '/test', content_type => 'application/x-null', data => undef, version => 1 },

    { id => 3, name => '1', path => '/test/1', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 4, name => 'before', path => '/test/1/before', content_type => 'text/plain', data => 3, version => 1 },
    { id => 5, name => 'value', path => '/test/1/before/value', content_type => 'text/plain', data => 7, version => 1 },

    { id => 6, name => '2', path => '/test/2', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 7, name => 'before', path => '/test/2/before', content_type => 'text/plain', data => 4, version => 1 },
    { id => 8, name => 'value', path => '/test/2/before/value', content_type => 'text/plain', data => 8, version => 1 },

    { id => 9, name => '3', path => '/test/3', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 10, name => 'src', path => '/test/3/src', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 11, name => 'dir', path => '/test/3/src/dir', content_type => 'text/plain', data => 5, version => 1 },
    { id => 12, name => 'value', path => '/test/3/src/dir/value', content_type => 'text/plain', data => 6, version => 1 },
    { id => 13, name => 'dst', path => '/test/3/dst', content_type => 'application/x-null', data => undef, version => 1 },

    { id => 14, name => '4', path => '/test/4', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 15, name => 'src', path => '/test/4/src', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 16, name => 'dir', path => '/test/4/src/dir', content_type => 'text/plain', data => 11, version => 1 },
    { id => 17, name => 'value', path => '/test/4/src/dir/value', content_type => 'text/plain', data => 12, version => 1 },
    { id => 18, name => 'dst', path => '/test/4/dst', content_type => 'application/x-null', data => undef, version => 1 },

    { id => 19, name => '5', path => '/test/5', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 20, name => 'src', path => '/test/5/src', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 21, name => 'before', path => '/test/5/src/before', content_type => 'text/plain', data => 23, version => 1 },
    { id => 22, name => 'value', path => '/test/5/src/before/value', content_type => 'text/plain', data => 24, version => 1 },
    { id => 23, name => 'dst', path => '/test/5/dst', content_type => 'application/x-null', data => undef, version => 1 },

    { id => 24, name => '6', path => '/test/6', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 25, name => 'src', path => '/test/6/src', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 26, name => 'before', path => '/test/6/src/before', content_type => 'text/plain', data => 31, version => 1 },
    { id => 27, name => 'value', path => '/test/6/src/before/value', content_type => 'text/plain', data => 32, version => 1 },
    { id => 28, name => 'dst', path => '/test/6/dst', content_type => 'application/x-null', data => undef, version => 1 },
);
$tree->finalize();


$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 4, name => 'after', path => '/test/1/after', content_type => 'text/plain', data => 3, version => 2 },
);
$tree->finalize();
ok(!$tree->get('/test/1/before'), "rename: /test/1/before not exists");
ok(!$tree->get('/test/1/before/value'), "rename: /test/1/before/value not exists");
is($tree->get('/test/1/after')->value, 3, "rename: /test/1/after has correct value");
is($tree->get('/test/1/after/value')->value, 7, "rename: /test/1/after/value has correct value");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 7, name => 'after', path => '/test/2/after', content_type => 'text/plain', data => 4, version => 2 },
    { id => 101, name => 'before', path => '/test/2/before', content_type => 'application/x-symlink', data => '/test/2/after', version => 2 },
);
$tree->finalize();
is($tree->get('/test/2/before')->value, 4, "rename with symlink: /test/2/before has correct value");
is($tree->get('/test/2/before/value')->value, 8, "rename with symlink: /test/2/before/value has correct value");
is($tree->get('/test/2/after')->value, 4, "rename with symlink: /test/2/after has correct value");
is($tree->get('/test/2/after/value')->value, 8, "rename with symlink: /test/2/after/value has correct value");


$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 11, name => 'dir', path => '/test/3/dst/dir', content_type => 'text/plain', data => 5, version => 2 },
);
$tree->finalize();
ok(!$tree->get('/test/3/src/dir'), "move: /test/3/src/dir not exists");
ok(!$tree->get('/test/3/src/dir/value'), "move: /test/3/src/dir/value not exists");
is($tree->get('/test/3/dst/dir')->value, 5, "move: /test/3/dst/dir has correct value");
is($tree->get('/test/3/dst/dir/value')->value, 6, "move: /test/3/dst/dir/value has correct value");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 16, name => 'dir', path => '/test/4/dst/dir', content_type => 'text/plain', data => 11, version => 2 },
    { id => 102, name => 'dir', path => '/test/4/src/dir', content_type => 'application/x-symlink', data => '/test/4/dst/dir', version => 2 },
);
$tree->finalize();
is($tree->get('/test/4/src/dir')->value, 11, "move with symlink: /test/4/src/dir has correct value");
is($tree->get('/test/4/src/dir/value')->value, 12, "move with symlink: /test/4/src/dir/value has correct value");
is($tree->get('/test/4/dst/dir')->value, 11, "move with symlink: /test/4/dst/dir has correct value");
is($tree->get('/test/4/dst/dir/value')->value, 12, "move with symlink: /test/4/dst/dir/value has correct value");


$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 21, name => 'after', path => '/test/5/dst/after', content_type => 'text/plain', data => 23, version => 2 },
);
$tree->finalize();
ok(!$tree->get('/test/5/src/before'), "move-n-rename: /test/5/src/before not exists");
ok(!$tree->get('/test/5/src/before/value'), "move-n-rename: /test/5/src/before/value not exists");
is($tree->get('/test/5/dst/after')->value, 23, "move-n-rename: /test/5/dst/after has correct value");
is($tree->get('/test/5/dst/after/value')->value, 24, "move-n-rename: /test/5/dst/after/value has correct value");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 26, name => 'after', path => '/test/6/dst/after', content_type => 'text/plain', data => 31, version => 2 },
    { id => 103, name => 'before', path => '/test/6/src/before', content_type => 'application/x-symlink', data => '/test/6/dst/after', version => 2 },
);
$tree->finalize();
is($tree->get('/test/6/src/before')->value, 31, "move-n-rename with symlink: /test/6/src/before has correct value");
is($tree->get('/test/6/src/before/value')->value, 32, "move-n-rename with symlink: /test/6/src/before/value has correct value");
is($tree->get('/test/6/dst/after')->value, 31, "move-n-rename with symlink: /test/6/dst/after has correct value");
is($tree->get('/test/6/dst/after/value')->value, 32, "move-n-rename with symlink: /test/6/dst/after/value has correct value");
