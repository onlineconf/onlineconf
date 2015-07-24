#!/usr/bin/perl

use strict;
use warnings;

use Test::More tests => 10;

use MR::OnlineConf::Admin::PerlMemory;
use MR::OnlineConf::Admin::PerlMemory::Parameter;

my $tree = MR::OnlineConf::Admin::PerlMemory->new(
    list => [{
        ID => 1,
        Name => '',
        Path => '/',
        ContentType => 'application/x-null',
        Value => undef,
        Version => 1,
        MTime => '',
        Deleted => 0,
    }],
    mtime => ''
);

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 1, Name => '', Path => '/', ContentType => 'application/x-null', Value => undef, Version => 1 },
    { Deleted => 0, MTime => '', ID => 2, Name => 'test', Path => '/test', ContentType => 'application/x-null', Value => undef, Version => 1 },
    { Deleted => 0, MTime => '', ID => 3, Name => 'branch1', Path => '/test/branch1', ContentType => 'application/x-null', Value => undef, Version => 1 },
    { Deleted => 0, MTime => '', ID => 4, Name => 'branch2', Path => '/test/branch2', ContentType => 'application/x-null', Value => undef, Version => 1 },
    { Deleted => 0, MTime => '', ID => 5, Name => 'value', Path => '/test/branch1/value', ContentType => 'text/plain', Value => 1, Version => 1 },
    { Deleted => 0, MTime => '', ID => 6, Name => 'value', Path => '/test/branch2/value', ContentType => 'text/plain', Value => 2, Version => 1 },
    { Deleted => 0, MTime => '', ID => 7, Name => 'current', Path => '/test/current', ContentType => 'application/x-symlink', Value => '/test/branch1', Version => 1 },
    { Deleted => 0, MTime => '', ID => 8, Name => 'value', Path => '/test/value', ContentType => 'application/x-symlink', Value => '/test/current/value', Version => 1 },
    { Deleted => 0, MTime => '', ID => 9, Name => 'value2', Path => '/test/value2', ContentType => 'application/x-symlink', Value => '/test/value', Version => 1 },
    { Deleted => 0, MTime => '', ID => 10, Name => 'recursive', Path => '/test/recursive', ContentType => 'application/x-null', Value => undef, Version => 1 },
    { Deleted => 0, MTime => '', ID => 11, Name => 'r1', Path => '/test/recursive/r1', ContentType => 'application/x-symlink', Value => '/test/recursive/r2', Version => 1 },
    { Deleted => 0, MTime => '', ID => 12, Name => 'r2', Path => '/test/recursive/r2', ContentType => 'application/x-symlink', Value => '/test/recursive/r1', Version => 1 },
    { Deleted => 0, MTime => '', ID => 13, Name => 'r3', Path => '/test/recursive/r3', ContentType => 'application/x-symlink', Value => '/test/recursive/r3', Version => 1 },
);

$tree->serialize();
is($tree->get('/test/value')->value, 1, "init: /test/value has correct value");
is($tree->get('/test/value2')->value, 1, "init: /test/value2 has correct value");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 7, Name => 'current', Path => '/test/current', ContentType => 'application/x-symlink', Value => '/test/branch2', Version => 2 },
);

$tree->serialize();
is($tree->get('/test/value')->value, 2, "change /test/current to another symlink: /test/value has correct value");
is($tree->get('/test/value2')->value, 2, "change /test/current to another symlink: /test/value2 has correct value");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 7, Name => 'current', Path => '/test/current', ContentType => 'application/x-null', Value => undef, Version => 3 },
    { Deleted => 0, MTime => '', ID => 14, Name => 'value', Path => '/test/current/value', ContentType => 'text/plain', Value => 3, Version => 1 },
);

$tree->serialize();
is($tree->get('/test/value')->value, 3, "change /test/current to real node: /test/value has correct value");
is($tree->get('/test/value2')->value, 3, "change /test/current to real node: /test/value2 has correct value");

$tree->delete(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 1, MTime => '', ID => 14, Name => 'value', Path => '/test/current/value', ContentType => 'text/plain', Value => 3, Version => 2 },
);

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 7, Name => 'current', Path => '/test/current', ContentType => 'application/x-symlink', Value => '/test/branch1', Version => 4 },
);

$tree->serialize();
is($tree->get('/test/value')->value, 1, "change /test/current to symlink: /test/value has correct value");
is($tree->get('/test/value2')->value, 1, "change /test/current to symlink: /test/value2 has correct value");

ok(!defined($tree->get('/test/recursive/r1')), "recursive symlink");
ok(!defined($tree->get('/test/recursive/r3')), "symlink on itself");
