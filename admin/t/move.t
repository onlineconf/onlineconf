#!/usr/bin/perl

use strict;
use warnings;

use Test::More tests => 24;

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
);

$tree->serialize();

# reName

SKIP: {
    skip "WTF reName?", 4;

    $tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
        { Deleted => 0, MTime => '', ID => 3, Name => '1', Path => '/test/1', ContentType => 'application/x-null', Value => undef, Version => 1 },
        { Deleted => 0, MTime => '', ID => 4, Name => 'before', Path => '/test/1/before', ContentType => 'text/plain', Value => 3, Version => 1 },
        { Deleted => 0, MTime => '', ID => 5, Name => 'value', Path => '/test/1/before/value', ContentType => 'text/plain', Value => 7, Version => 1 },
    );

    $tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
        { Deleted => 0, MTime => '', ID => 4, Name => 'after', Path => '/test/1/after', ContentType => 'text/plain', Value => 3, Version => 2 },
    );

    $tree->serialize();
    ok(!$tree->get('/test/1/before'), "reName: /test/1/before not exists");
    ok(!$tree->get('/test/1/before/value'), "reName: /test/1/before/value not exists");
    is($tree->get('/test/1/after')->value, 3, "reName: /test/1/after has correct value");
    is($tree->get('/test/1/after/value')->value, 7, "reName: /test/1/after/value has correct value");
}

# reName with symlink

SKIP: {
    skip "WTF reName with symlink", 4;

    $tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
        { Deleted => 0, MTime => '', ID => 6, Name => '2', Path => '/test/2', ContentType => 'application/x-null', Value => undef, Version => 1 },
        { Deleted => 0, MTime => '', ID => 7, Name => 'before', Path => '/test/2/before', ContentType => 'text/plain', Value => 4, Version => 1 },
        { Deleted => 0, MTime => '', ID => 8, Name => 'value', Path => '/test/2/before/value', ContentType => 'text/plain', Value => 8, Version => 1 },
    );

    $tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
        { Deleted => 0, MTime => '', ID => 7, Name => 'after', Path => '/test/2/after', ContentType => 'text/plain', Value => 4, Version => 2 },
        { Deleted => 0, MTime => '', ID => 101, Name => 'before', Path => '/test/2/before', ContentType => 'application/x-symlink', Value => '/test/2/after', Version => 2 },
    );

    $tree->serialize();
    is($tree->get('/test/2/before')->value, 4, "reName with symlink: /test/2/before has correct value");
    is($tree->get('/test/2/before/value')->value, 8, "reName with symlink: /test/2/before/value has correct value");
    is($tree->get('/test/2/after')->value, 4, "reName with symlink: /test/2/after has correct value");
    is($tree->get('/test/2/after/value')->value, 8, "reName with symlink: /test/2/after/value has correct value");
}

# Move

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 9, Name => '3', Path => '/test/3', ContentType => 'application/x-null', Value => undef, Version => 1 },
    { Deleted => 0, MTime => '', ID => 10, Name => 'src', Path => '/test/3/src', ContentType => 'application/x-null', Value => undef, Version => 1 },
    { Deleted => 0, MTime => '', ID => 11, Name => 'dir', Path => '/test/3/src/dir', ContentType => 'text/plain', Value => 5, Version => 1 },
    { Deleted => 0, MTime => '', ID => 12, Name => 'value', Path => '/test/3/src/dir/value', ContentType => 'text/plain', Value => 6, Version => 1 },
    { Deleted => 0, MTime => '', ID => 13, Name => 'dst', Path => '/test/3/dst', ContentType => 'application/x-null', Value => undef, Version => 1 },
);

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 11, Name => 'dir', Path => '/test/3/dst/dir', ContentType => 'text/plain', Value => 5, Version => 2 },
    { Deleted => 0, MTime => '', ID => 12, Name => 'value', Path => '/test/3/dst/dir/value', ContentType => 'text/plain', Value => 6, Version => 1 },
);

$tree->serialize();
ok(!$tree->get('/test/3/src/dir'), "move: /test/3/src/dir not exists");
ok(!$tree->get('/test/3/src/dir/value'), "move: /test/3/src/dir/value not exists");
is($tree->get('/test/3/dst/dir')->value, 5, "move: /test/3/dst/dir has correct value");
is($tree->get('/test/3/dst/dir/value')->value, 6, "move: /test/3/dst/dir/value has correct value");

# Move with symlink

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 14, Name => '4', Path => '/test/4', ContentType => 'application/x-null', Value => undef, Version => 1 },
    { Deleted => 0, MTime => '', ID => 15, Name => 'src', Path => '/test/4/src', ContentType => 'application/x-null', Value => undef, Version => 1 },
    { Deleted => 0, MTime => '', ID => 16, Name => 'dir', Path => '/test/4/src/dir', ContentType => 'text/plain', Value => 11, Version => 1 },
    { Deleted => 0, MTime => '', ID => 17, Name => 'value', Path => '/test/4/src/dir/value', ContentType => 'text/plain', Value => 12, Version => 1 },
    { Deleted => 0, MTime => '', ID => 18, Name => 'dst', Path => '/test/4/dst', ContentType => 'application/x-null', Value => undef, Version => 1 },
);

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 102, Name => 'dir', Path => '/test/4/src/dir', ContentType => 'application/x-symlink', Value => '/test/4/dst/dir', Version => 2 },
    { Deleted => 0, MTime => '', ID => 16, Name => 'dir', Path => '/test/4/dst/dir', ContentType => 'text/plain', Value => 11, Version => 2 },
    { Deleted => 0, MTime => '', ID => 17, Name => 'value', Path => '/test/4/dst/dir/value', ContentType => 'text/plain', Value => 12, Version => 1 },
);

$tree->serialize();
is($tree->get('/test/4/src/dir')->value, 11, "move with symlink: /test/4/src/dir has correct value");
is($tree->get('/test/4/src/dir/value')->value, 12, "move with symlink: /test/4/src/dir/value has correct value");
is($tree->get('/test/4/dst/dir')->value, 11, "move with symlink: /test/4/dst/dir has correct value");
is($tree->get('/test/4/dst/dir/value')->value, 12, "move with symlink: /test/4/dst/dir/value has correct value");

# move-n-reName

SKIP: {
    skip "WTF move-n-reName", 4;

    $tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
        { Deleted => 0, MTime => '', ID => 19, Name => '5', Path => '/test/5', ContentType => 'application/x-null', Value => undef, Version => 1 },
        { Deleted => 0, MTime => '', ID => 20, Name => 'src', Path => '/test/5/src', ContentType => 'application/x-null', Value => undef, Version => 1 },
        { Deleted => 0, MTime => '', ID => 21, Name => 'before', Path => '/test/5/src/before', ContentType => 'text/plain', Value => 23, Version => 1 },
        { Deleted => 0, MTime => '', ID => 22, Name => 'value', Path => '/test/5/src/before/value', ContentType => 'text/plain', Value => 24, Version => 1 },
        { Deleted => 0, MTime => '', ID => 23, Name => 'dst', Path => '/test/5/dst', ContentType => 'application/x-null', Value => undef, Version => 1 },
    );

    $tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
        { Deleted => 0, MTime => '', ID => 21, Name => 'after', Path => '/test/5/dst/after', ContentType => 'text/plain', Value => 23, Version => 2 },
    );

    $tree->serialize();
    ok(!$tree->get('/test/5/src/before'), "move-n-reName: /test/5/src/before not exists");
    ok(!$tree->get('/test/5/src/before/value'), "move-n-reName: /test/5/src/before/value not exists");
    is($tree->get('/test/5/dst/after')->value, 23, "move-n-reName: /test/5/dst/after has correct value");
    is($tree->get('/test/5/dst/after/value')->value, 24, "move-n-reName: /test/5/dst/after/value has correct value");
}

# move-n-reName with symlink

SKIP: {
    skip "WTF move-n-reName with symlink", 4;

    $tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
        { Deleted => 0, MTime => '', ID => 24, Name => '6', Path => '/test/6', ContentType => 'application/x-null', Value => undef, Version => 1 },
        { Deleted => 0, MTime => '', ID => 25, Name => 'src', Path => '/test/6/src', ContentType => 'application/x-null', Value => undef, Version => 1 },
        { Deleted => 0, MTime => '', ID => 26, Name => 'before', Path => '/test/6/src/before', ContentType => 'text/plain', Value => 31, Version => 1 },
        { Deleted => 0, MTime => '', ID => 27, Name => 'value', Path => '/test/6/src/before/value', ContentType => 'text/plain', Value => 32, Version => 1 },
        { Deleted => 0, MTime => '', ID => 28, Name => 'dst', Path => '/test/6/dst', ContentType => 'application/x-null', Value => undef, Version => 1 },
    );

    $tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
        { Deleted => 0, MTime => '', ID => 26, Name => 'after', Path => '/test/6/dst/after', ContentType => 'text/plain', Value => 31, Version => 2 },
        { Deleted => 0, MTime => '', ID => 103, Name => 'before', Path => '/test/6/src/before', ContentType => 'application/x-symlink', Value => '/test/6/dst/after', Version => 2 },
    );

    $tree->serialize();
    is($tree->get('/test/6/src/before')->value, 31, "move-n-reName with symlink: /test/6/src/before has correct value");
    is($tree->get('/test/6/src/before/value')->value, 32, "move-n-reName with symlink: /test/6/src/before/value has correct value");
    is($tree->get('/test/6/dst/after')->value, 31, "move-n-reName with symlink: /test/6/dst/after has correct value");
    is($tree->get('/test/6/dst/after/value')->value, 32, "move-n-reName with symlink: /test/6/dst/after/value has correct value");
}
