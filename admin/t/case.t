#!/usr/bin/perl

use strict;
use warnings;

use Test::More tests => 35;

use Sys::Hostname;
use MR::OnlineConf::Admin::PerlMemory;
use MR::OnlineConf::Admin::PerlMemory::Parameter;

my $hostname = hostname();

my $sort = MR::OnlineConf::Admin::PerlMemory::Parameter->new(
    ID => 1,
    Name => 'case',
    Path => '/case',
    ContentType => 'application/x-case',
    Value => '{}',
    Version => 1,
    MTime => '',
    Deleted => 0,
    groups => ['test', 'my'],
);

my @sorted = $sort->_sort_case([
    {},
    { server => '*' },
    { group => 'my' },
    { datacenter => 'varshavka' },
    { server => 'test*.{mail.ru,mydev.mail.ru}' },
    { group => 'alpha' },
    { datacenter => 'korovinka' },
    { server => 'test.{mail.ru,mydev.mail.ru}' },
    { group => 'test' },
]);

ok(
    $sorted[0]->{server} eq 'test.{mail.ru,mydev.mail.ru}'
    && $sorted[1]->{server} eq 'test*.{mail.ru,mydev.mail.ru}'
    && $sorted[2]->{server} eq '*'
    && $sorted[3]->{group} eq 'test'
    && $sorted[4]->{group} eq 'my'
    && $sorted[5]->{group} eq 'alpha'
    && $sorted[6]->{datacenter} eq 'korovinka'
    && $sorted[7]->{datacenter} eq 'varshavka'
    && $sorted[8], "case: correct order"
);

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
    host => $hostname,
    addr => ['188.93.61.150'],
    mtime => ''
);

my @init = (
    { ID => 2, Name => 'test', Path => '/test', ContentType => 'application/x-null', data => undef, version => 1 },
    { ID => 3, Name => 'case1', Path => '/test/case1', ContentType => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"server":"$hostname","mime":"text/plain","value":10}]#, version => 1 },
    { ID => 4, Name => 'case2', Path => '/test/case2', ContentType => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"text/plain","value":15}]#, version => 1 },
    { ID => 5, Name => 'case3', Path => '/test/case3', ContentType => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"server":"xxx","mime":"text/plain","value":10}]#, version => 1 },
    { ID => 6, Name => 'value', Path => '/test/value', ContentType => 'text/plain', data => '6', version => 1 },
    { ID => 7, Name => 'case4', Path => '/test/case4', ContentType => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/value"}]#, version => 1 },
    { ID => 8, Name => 'case5', Path => '/test/case5', ContentType => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/case4"}]#, version => 1 },
    { ID => 9, Name => 'branch1', Path => '/test/branch1', ContentType => 'application/x-null', data => undef, version => 1 },
    { ID => 10, Name => 'value', Path => '/test/branch1/value', ContentType => 'text/plain', data => '16', version => 1 },
    { ID => 11, Name => 'branch2', Path => '/test/branch2', ContentType => 'application/x-null', data => undef, version => 1 },
    { ID => 12, Name => 'value', Path => '/test/branch2/value', ContentType => 'text/plain', data => '18', version => 1 },
    { ID => 13, Name => 'current', Path => '/test/current', ContentType => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1"}]#, version => 1 },
    { ID => 14, Name => 'mod', Path => '/test/mod', ContentType => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"$hostname","mime":"application/x-symlink","value":"/test/branch2/value"}]#, version => 1 },
    { ID => 15, Name => 'curmod', Path => '/test/curmod', ContentType => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1"},{"server":"$hostname","mime":"application/x-symlink","value":"/test/branch2"}]#, version => 1 },
    { ID => 16, Name => 'case6', Path => '/test/case6', ContentType => 'application/x-case', data => qq#[{"server":"xxx","mime":"text/plain","value":5},{"datacenter":"xxx","mime":"text/plain","value":6}]#, version => 1 },
    { ID => 17, Name => 'case7', Path => '/test/case7', ContentType => 'application/x-case', data => qq#[{"server":"$hostname","mime":"application/x-case","value":"[{\\"datacenter\\":\\"test-dc\\",\\"mime\\":\\"text/plain\\",\\"value\\":\\"10\\"}]"}]#, version => 1 },
    { ID => 18, Name => 'branch3', Path => '/test/branch3', ContentType => 'application/x-null', data => undef, version => 1 },
    { ID => 19, Name => 'value', Path => '/test/branch3/value', ContentType => 'text/plain', data => '20', version => 1 },
    { ID => 101, Name => 'onlineconf', Path => '/onlineconf', ContentType => 'application/x-null', data => undef, version => 1 },
    { ID => 102, Name => 'datacenter', Path => '/onlineconf/datacenter', ContentType => 'application/x-symlink', data => '/infrastructure/datacenter', version => 1 },
    { ID => 103, Name => 'infrastructure', Path => '/infrastructure', ContentType => 'application/x-null', data => undef, version => 1 },
    { ID => 104, Name => 'datacenter', Path => '/infrastructure/datacenter', ContentType => 'application/x-null', data => undef, version => 1 },
    { ID => 105, Name => 'test-dc', Path => '/infrastructure/datacenter/test-dc', ContentType => 'text/plain', data => '188.93.61.0/24', version => 1 },
    { ID => 106, Name => 'false-dc', Path => '/infrastructure/datacenter/false-dc', ContentType => 'text/plain', data => '188.93.11.0/24', version => 1 },
    { ID => 107, Name => 'group', Path => '/onlineconf/group', ContentType => 'application/x-symlink', data => '/infrastructure/group', version => 1 },
    { ID => 108, Name => 'group', Path => '/infrastructure/group', ContentType => 'application/x-null', data => undef, version => 1 },
    { ID => 109, Name => 'priority', Path => '/infrastructure/group/priority', ContentType => 'text/plain', data => 'stat,*', version => 1 },
    { ID => 110, Name => 'alei', Path => '/infrastructure/group/alei', ContentType => 'text/plain', data => 'alei*.{mail.ru,mydev.mail.ru}', version => 1 },
    { ID => 111, Name => 'stat', Path => '/infrastructure/group/stat', ContentType => 'text/plain', data => 'alei{22,33,44,50}.{mail.ru,mydev.mail.ru}', version => 1 },
    { ID => 112, Name => 'alpha', Path => '/infrastructure/group/alpha', ContentType => 'text/plain', data => 'myalpha*.i.mail.ru', version => 1 },
    { ID => 113, Name => 'top', Path => '/infrastructure/group/top', ContentType => 'application/x-null', data => undef, version => 1 },
    { ID => 114, Name => 'subtop', Path => '/infrastructure/group/subtop', ContentType => 'text/plain', data => 'alei*.{mail.ru,mydev.mail.ru}', version => 1 },
    { ID => 115, Name => 'sub', Path => '/infrastructure/group/top/sub', ContentType => 'application/x-symlink', data => '/infrastructure/group/subtop', version => 1 },
    { ID => 52, Name => 'atest', Path => '/atest', ContentType => 'application/x-null', data => undef, version => 1 },
    { ID => 54, Name => 'case2', Path => '/atest/case2', ContentType => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"text/plain","value":25}]#, version => 1 },
    { ID => 55, Name => 'case3', Path => '/atest/case3', ContentType => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"group":"alei","mime":"text/plain","value":25}]#, version => 1 },
);

foreach my $item (@init) {
    $item->{MTime} = '';
    $item->{Value} = delete $item->{data};
    $item->{Deleted} = 0;
    $item->{Version} = delete $item->{version};

    $tree->put(
        MR::OnlineConf::Admin::PerlMemory::Parameter->new($item)
    );
}

$tree->serialize();

is($tree->get('/test/case1')->value, 10, "when server: /test/case1 has correct value");
is($tree->get('/test/case2')->value, 15, "when datacenter: /test/case2 has correct value");
is($tree->get('/test/case3')->value, 5, "default: /test/case3 has correct value");
is($tree->get('/test/case4')->value, 6, "when datacenter: symlink in /test/case4 has correct value");
is($tree->get('/test/case5')->value, 6, "when datacenter: symlink in /test/case5 has correct value");
is($tree->get('/test/case6')->value, undef, "no such case");
is($tree->get('/test/case7')->value, 10, "deep datacenter case");
is($tree->get('/atest/case2')->value, 25, "when datacenter: /test/acase2 has correct value");
is($tree->get('/atest/case3')->value, 25, "when group: /test/acase3 has correct value");
is($tree->get('/onlineconf/datacenter/test-dc')->value, '188.93.61.0/24', "/onlinconf/datacenter/test-dc has correct value");
is($tree->get('/test/mod')->value, 18, "/test/mod use server");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 6, Name => 'value', Path => '/test/value', ContentType => 'text/plain', Value => '8', Version => 2 },
);
$tree->serialize();
is($tree->get('/test/case4')->value, 8, "when datacenter: change symlink target value");
is($tree->get('/test/case5')->value, 8, "when datacenter: change indirect symlink target value");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 4, Name => 'case2', Path => '/test/case2', ContentType => 'application/x-case', Value => qq#[{"mime":"text/plain","value":5},{"datacenter":"false-dc","mime":"text/plain","value":15}]#, Version => 2 },
);
$tree->serialize();
is($tree->get('/test/case2')->value, 5, "when datacenter: change to default");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 4, Name => 'case2', Path => '/test/case2', ContentType => 'application/x-case', Value => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"text/plain","value":15}]#, Version => 3 },
);
$tree->serialize();
is($tree->get('/test/case2')->value, 15, "when datacenter: change to datacenter");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 105, Name => 'test-dc', Path => '/infrastructure/datacenter/test-dc', ContentType => 'text/plain', Value => '188.93.60.0/24', Version => 2 },
);
$tree->serialize();
is($tree->get('/test/case2')->value, 5, "when datacenter: change datacenter definition");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 105, Name => 'test-dc', Path => '/infrastructure/datacenter/test-dc', ContentType => 'application/json', Value => '["188.93.61.0/24"]', Version => 3 },
);
$tree->serialize();
is($tree->get('/test/case2')->value, 15, "when datacenter: change datacenter definition back");

is($tree->get('/test/current/value')->value, 16, "/test/current/value has correct value on init");
$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 13, Name => 'current', Path => '/test/current', ContentType => 'application/x-case', Value => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch2"}]#, Version => 2 },
);
$tree->serialize();
is($tree->get('/test/current/value')->value, 18, "/test/current/value has correct value after /test/current modification");

is($tree->get('/test/mod')->value, 18, "/test/mod use server");
$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 14, Name => 'mod', Path => '/test/mod', ContentType => 'application/x-case', Value => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2/value"}]#, Version => 2 },
);
$tree->serialize();
is($tree->get('/test/mod')->value, 16, "/test/mod use datacenter");
$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 14, Name => 'mod', Path => '/test/mod', ContentType => 'application/x-case', Value => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"xxx","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2/value"}]#, Version => 3 },
);
$tree->serialize();
is($tree->get('/test/mod')->value, 8, "/test/mod use default");
$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 14, Name => 'mod', Path => '/test/mod', ContentType => 'application/x-case', Value => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2/value"}]#, Version => 4 },
);
$tree->serialize();
is($tree->get('/test/mod')->value, 16, "/test/mod use datacenter back");
$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 14, Name => 'mod', Path => '/test/mod', ContentType => 'application/x-case', Value => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"$hostname","mime":"application/x-symlink","value":"/test/branch2/value"}]#, Version => 5 },
);
$tree->serialize();
is($tree->get('/test/mod')->value, 18, "/test/mod use server back");

is($tree->get('/test/curmod/value')->value, 18, "/test/curmod/value use server");
$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 15, Name => 'curmod', Path => '/test/curmod', ContentType => 'application/x-case', Value => qq#[{"mime":"application/x-symlink","value":"/test"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2"}]#, Version => 2 },
);
$tree->serialize();
is($tree->get('/test/curmod/value')->value, 16, "/test/curmod/value use datacenter");
$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 15, Name => 'curmod', Path => '/test/curmod', ContentType => 'application/x-case', Value => qq#[{"mime":"application/x-symlink","value":"/test"},{"datacenter":"xxx","mime":"application/x-symlink","value":"/test/branch1"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2"}]#, Version => 3 },
);
$tree->serialize();
is($tree->get('/test/curmod/value')->value, 8, "/test/curmod/value use default");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 14, Name => 'mod', Path => '/test/mod', ContentType => 'application/x-case', Value => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2/value"},{"group":"alei","mime":"application/x-symlink","value":"/test/branch3/value"}]#, Version => 6 },
);
$tree->serialize();
is($tree->get('/test/mod')->value, 20, "/test/mod use group");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 14, Name => 'mod', Path => '/test/mod', ContentType => 'application/x-case', Value => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"alei*.{mail.ru,mydev.mail.ru}","mime":"application/x-symlink","value":"/test/branch2/value"},{"group":"alei","mime":"application/x-symlink","value":"/test/branch3/value"}]#, Version => 7 },
);
$tree->serialize();
is($tree->get('/test/mod')->value, 18, "/test/mod use server");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 14, Name => 'mod', Path => '/test/mod', ContentType => 'application/x-case', Value => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"group":"stat","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2/value"},{"group":"alei","mime":"application/x-symlink","value":"/test/branch3/value"}]#, Version => 8 },
);
$tree->serialize();
is($tree->get('/test/mod')->value, 16, "/test/mod use another group");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 111, Name => 'stat', Path => '/infrastructure/group/stat', ContentType => 'text/plain', Value => 'stat*.mail.ru', Version => 2 },
);
$tree->serialize();
is($tree->get('/test/mod')->value, 20, "when group: change group definition");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 111, Name => 'stat', Path => '/infrastructure/group/stat', ContentType => 'text/plain', Value => 'alei*.{mail.ru,mydev.mail.ru}', Version => 3 },
);
$tree->serialize();
is($tree->get('/test/mod')->value, 16, "when group: change group definition back");

$tree->delete(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 1, MTime => '', ID => 111, Name => 'stat', Path => '/infrastructure/group/stat', ContentType => 'text/plain', Value => 'alei*.{mail.ru,mydev.mail.ru}', Version => 4 },
);
$tree->serialize();
is($tree->get('/test/mod')->value, 20, "when group: remove group definition");

$tree->put(MR::OnlineConf::Admin::PerlMemory::Parameter->new($_)) foreach (
    {Deleted => 0, MTime => '', ID => 111, Name => 'stat', Path => '/infrastructure/group/stat', ContentType => 'text/plain', Value => 'alei*.{mail.ru,mydev.mail.ru}', Version => 5 },
);
$tree->serialize();
is($tree->get('/test/mod')->value, 16, "when group: add group definition");
