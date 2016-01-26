#!/usr/bin/perl

use strict;
use warnings;

use Test::More tests => 16;

use Socket;
use Sys::Hostname;
use MR::OnlineConf::Admin::PerlMemory;
use MR::OnlineConf::Admin::PerlMemory::Parameter;

my ($short_hostname, $ip) = map {
    my $h = `$_`; chomp $h; $h
} (
    'hostname -s', 'hostname -i'
);

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
    { Deleted => 0, MTime => '', ID => 4, Name => 'template1', Path => '/test/template1', ContentType => 'application/x-template', Value => '${hostname}#${short_hostname}#${ip}#${/test/value}', Version => 1 },
    { Deleted => 0, MTime => '', ID => 5, Name => 'template2', Path => '/test/template2', ContentType => 'application/x-template', Value => '${hostname}#${short_hostname}#${ip}#${/test/symlink}', Version => 1 },
    { Deleted => 0, MTime => '', ID => 6, Name => 'template3', Path => '/test/template3', ContentType => 'application/x-template', Value => '${hostname}#${short_hostname}#${ip}#${/symlink/value}', Version => 1 },
    { Deleted => 0, MTime => '', ID => 7, Name => 'template4', Path => '/test/template4', ContentType => 'application/x-template', Value => '${hostname}#${short_hostname}#${ip}#${/symlink/symlink}', Version => 1 },
    { Deleted => 0, MTime => '', ID => 51, Name => 'symlink', Path => '/test/symlink', ContentType => 'application/x-symlink', Value => '/test/value', Version => 1 },
    { Deleted => 0, MTime => '', ID => 52, Name => 'symlink', Path => '/symlink', ContentType => 'application/x-symlink', Value => '/test', Version => 1 },
);

$tree->serialize();
is($tree->get('/test/template1')->value, "$host#$short_hostname#$ip#3", "insert template: correct value 1");
is($tree->get('/test/template2')->value, "$host#$short_hostname#$ip#3", "insert template: correct value 2");
is($tree->get('/test/template3')->value, "$host#$short_hostname#$ip#3", "insert template: correct value 3");
is($tree->get('/test/template4')->value, "$host#$short_hostname#$ip#3", "insert template: correct value 4");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 3, Name => 'value', Path => '/test/value', ContentType => 'text/plain', Value => 4, Version => 2 },
);

$tree->serialize();
is($tree->get('/test/template1')->value, "$host#$short_hostname#$ip#4", "update value: template is updated too 1");
is($tree->get('/test/template2')->value, "$host#$short_hostname#$ip#4", "update value: template is updated too 2");
is($tree->get('/test/template3')->value, "$host#$short_hostname#$ip#4", "update value: template is updated too 3");
is($tree->get('/test/template4')->value, "$host#$short_hostname#$ip#4", "update value: template is updated too 4");

$tree->delete(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 1, MTime => '', ID => 3, Name => 'value', Path => '/test/value', ContentType => 'text/plain', Value => 4, Version => 3 },
);

$tree->serialize();
is($tree->get('/test/template1')->value, "$host#$short_hostname#$ip#", "delete value: template is updated 1");
is($tree->get('/test/template2')->value, "$host#$short_hostname#$ip#", "delete value: template is updated 2");
is($tree->get('/test/template3')->value, "$host#$short_hostname#$ip#", "delete value: template is updated 3");
is($tree->get('/test/template4')->value, "$host#$short_hostname#$ip#", "delete value: template is updated 4");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    { Deleted => 0, MTime => '', ID => 3, Name => 'value', Path => '/test/value', ContentType => 'text/plain', Value => 5, Version => 4 },
);

$tree->serialize();
is($tree->get('/test/template1')->value, "$host#$short_hostname#$ip#5", "add value: template is updated 1");
is($tree->get('/test/template2')->value, "$host#$short_hostname#$ip#5", "add value: template is updated 2");
is($tree->get('/test/template3')->value, "$host#$short_hostname#$ip#5", "add value: template is updated 3");
is($tree->get('/test/template4')->value, "$host#$short_hostname#$ip#5", "add value: template is updated 4");
