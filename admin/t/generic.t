#!/usr/bin/perl

use strict;
use warnings;

use Test::More tests => 7;

use Socket;
use Sys::Hostname;
use MR::OnlineConf::Admin::PerlMemory;
use MR::OnlineConf::Admin::PerlMemory::Parameter;

my $host = hostname();
my $addr = Socket::inet_ntoa(scalar(gethostbyname($host)));
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
    host => $host,
    addr => [$addr],
    mtime => ''
);

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 1, Name => '', Path => '/', ContentType => 'application/x-null', Value => undef, Version => 1 },
    { Deleted => 0, MTime => '', ID => 2, Name => 'test', Path => '/test', ContentType => 'application/x-null', Value => undef, Version => 1 },
);
$tree->serialize();

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 3, Name => 'value', Path => '/test/value', ContentType => 'text/plain', Value => 3, Version => 1 },
);
$tree->serialize();
is($tree->get('/test/value')->value, 3, "insert: correct value");
is($tree->get('/test/value')->ContentType, 'text/plain', "insert: correct content type");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 3, Name => 'value', Path => '/test/value', ContentType => 'text/plain', Value => 4, Version => 2 },
);
$tree->serialize();
is($tree->get('/test/value')->value, 4, "update value");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 3, Name => 'value', Path => '/test/value', ContentType => 'application/json', Value => '[]', Version => 3 },
);
$tree->serialize();
is($tree->get('/test/value')->value, '[]', "update content type and value: correct value");
is($tree->get('/test/value')->ContentType, 'application/json', "update content type and value: correct content type");

$tree->delete(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 1, MTime => '', ID => 3, Name => 'value', Path => '/test/value', ContentType => 'application/json', Value => '[]', Version => 4 },
);
$tree->serialize();
ok(!$tree->get('/test/value'), "delete: not exists");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 3, Name => 'value', Path => '/test/value', ContentType => 'application/x-null', Value => '', Version => 5 },
    { Deleted => 0, MTime => '', ID => 4, Name => '1', Path => '/test/value/1', ContentType => 'text/plain', Value => '1', Version => 1 },
    { Deleted => 0, MTime => '', ID => 4, Name => '1', Path => '/test/value/1', ContentType => 'application/x-case', Value => qq#[{"mime":"text/plain","value":2},{"server":"$host","mime":"text/plain","value":3}]#, Version => 2 },
);
$tree->serialize();
is($tree->get('/test/value/1')->value, 3, '/test/value/1 change type from text to case: value');
