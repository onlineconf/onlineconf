#!/usr/bin/perl

use strict;
use warnings;
use Test::More tests => 35;
use Log::Dispatch;
use Sys::Hostname;
use MR::OnlineConf::Updater::Parameter;
use MR::OnlineConf::Updater::PerlMemory;

my $hostname = hostname();

my @case = (
    {},
    { server => '*' },
    { group => 'my' },
    { datacenter => 'varshavka' },
    { server => 'test*.mail.ru' },
    { group => 'alpha' },
    { datacenter => 'korovinka' },
    { server => 'test.mail.ru' },
    { group => 'test' },
);

my @groups = ('test', 'my');

my @sorted = MR::OnlineConf::Updater::Parameter->_sort_case(\@case, \@groups);
ok($sorted[0]->{server} eq 'test.mail.ru' && $sorted[1]->{server} eq 'test*.mail.ru' && $sorted[2]->{server} eq '*'
    && $sorted[3]->{group} eq 'test' && $sorted[4]->{group} eq 'my' && $sorted[5]->{group} eq 'alpha'
    && $sorted[6]->{datacenter} eq 'korovinka' && $sorted[7]->{datacenter} eq 'varshavka' && $sorted[8], "case: correct order");

my $log = Log::Dispatch->new(outputs => [[ 'Screen', min_level => 'info', newline => 1 ]]);

my $tree = MR::OnlineConf::Updater::PerlMemory->new(log => $log);

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 1, name => '', path => '/', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 2, name => 'test', path => '/test', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 3, name => 'case1', path => '/test/case1', content_type => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"server":"$hostname","mime":"text/plain","value":10}]#, version => 1 },
    { id => 4, name => 'case2', path => '/test/case2', content_type => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"text/plain","value":15}]#, version => 1 },
    { id => 5, name => 'case3', path => '/test/case3', content_type => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"server":"xxx","mime":"text/plain","value":10}]#, version => 1 },
    { id => 6, name => 'value', path => '/test/value', content_type => 'text/plain', data => '6', version => 1 },
    { id => 7, name => 'case4', path => '/test/case4', content_type => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/value"}]#, version => 1 },
    { id => 8, name => 'case5', path => '/test/case5', content_type => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/case4"}]#, version => 1 },
    { id => 9, name => 'branch1', path => '/test/branch1', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 10, name => 'value', path => '/test/branch1/value', content_type => 'text/plain', data => '16', version => 1 },
    { id => 11, name => 'branch2', path => '/test/branch2', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 12, name => 'value', path => '/test/branch2/value', content_type => 'text/plain', data => '18', version => 1 },
    { id => 13, name => 'current', path => '/test/current', content_type => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1"}]#, version => 1 },
    { id => 14, name => 'mod', path => '/test/mod', content_type => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"$hostname","mime":"application/x-symlink","value":"/test/branch2/value"}]#, version => 1 },
    { id => 15, name => 'curmod', path => '/test/curmod', content_type => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1"},{"server":"$hostname","mime":"application/x-symlink","value":"/test/branch2"}]#, version => 1 },
    { id => 16, name => 'case6', path => '/test/case6', content_type => 'application/x-case', data => qq#[{"server":"xxx","mime":"text/plain","value":5},{"datacenter":"xxx","mime":"text/plain","value":6}]#, version => 1 },
    { id => 17, name => 'case7', path => '/test/case7', content_type => 'application/x-case', data => qq#[{"server":"$hostname","mime":"application/x-case","value":"[{\\"datacenter\\":\\"test-dc\\",\\"mime\\":\\"text/plain\\",\\"value\\":\\"10\\"}]"}]#, version => 1 },
    { id => 18, name => 'branch3', path => '/test/branch3', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 19, name => 'value', path => '/test/branch3/value', content_type => 'text/plain', data => '20', version => 1 },
    { id => 101, name => 'onlineconf', path => '/onlineconf', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 102, name => 'datacenter', path => '/onlineconf/datacenter', content_type => 'application/x-symlink', data => '/infrastructure/datacenter', version => 1 },
    { id => 103, name => 'infrastructure', path => '/infrastructure', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 104, name => 'datacenter', path => '/infrastructure/datacenter', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 105, name => 'test-dc', path => '/infrastructure/datacenter/test-dc', content_type => 'text/plain', data => '188.93.61.0/24', version => 1 },
    { id => 106, name => 'false-dc', path => '/infrastructure/datacenter/false-dc', content_type => 'text/plain', data => '188.93.11.0/24', version => 1 },
    { id => 107, name => 'group', path => '/onlineconf/group', content_type => 'application/x-symlink', data => '/infrastructure/group', version => 1 },
    { id => 108, name => 'group', path => '/infrastructure/group', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 109, name => 'priority', path => '/infrastructure/group/priority', content_type => 'text/plain', data => 'stat,*', version => 1 },
    { id => 110, name => 'alei', path => '/infrastructure/group/alei', content_type => 'text/plain', data => 'alei*.mail.ru', version => 1 },
    { id => 111, name => 'stat', path => '/infrastructure/group/stat', content_type => 'text/plain', data => 'alei{22,33,44}.mail.ru', version => 1 },
    { id => 112, name => 'alpha', path => '/infrastructure/group/alpha', content_type => 'text/plain', data => 'myalpha*.i.mail.ru', version => 1 },
    { id => 113, name => 'top', path => '/infrastructure/group/top', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 114, name => 'subtop', path => '/infrastructure/group/subtop', content_type => 'text/plain', data => 'alei*.mail.ru', version => 1 },
    { id => 115, name => 'sub', path => '/infrastructure/group/top/sub', content_type => 'application/x-symlink', data => '/infrastructure/group/subtop', version => 1 },
    { id => 52, name => 'atest', path => '/atest', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 54, name => 'case2', path => '/atest/case2', content_type => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"text/plain","value":25}]#, version => 1 },
    { id => 55, name => 'case3', path => '/atest/case3', content_type => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"group":"alei","mime":"text/plain","value":25}]#, version => 1 },
);
$tree->finalize();
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

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 6, name => 'value', path => '/test/value', content_type => 'text/plain', data => '8', version => 2 },
);
$tree->finalize();
is($tree->get('/test/case4')->value, 8, "when datacenter: change symlink target value");
is($tree->get('/test/case5')->value, 8, "when datacenter: change indirect symlink target value");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 4, name => 'case2', path => '/test/case2', content_type => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"false-dc","mime":"text/plain","value":15}]#, version => 2 },
);
$tree->finalize();
is($tree->get('/test/case2')->value, 5, "when datacenter: change to default");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 4, name => 'case2', path => '/test/case2', content_type => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"text/plain","value":15}]#, version => 3 },
);
$tree->finalize();
is($tree->get('/test/case2')->value, 15, "when datacenter: change to datacenter");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 105, name => 'test-dc', path => '/infrastructure/datacenter/test-dc', content_type => 'text/plain', data => '188.93.60.0/24', version => 2 },
);
$tree->finalize();
is($tree->get('/test/case2')->value, 5, "when datacenter: change datacenter definition");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 105, name => 'test-dc', path => '/infrastructure/datacenter/test-dc', content_type => 'application/json', data => '["188.93.61.0/24"]', version => 3 },
);
$tree->finalize();
is($tree->get('/test/case2')->value, 15, "when datacenter: change datacenter definition back");

is($tree->get('/test/current/value')->value, 16, "/test/current/value has correct value on init");
$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 13, name => 'current', path => '/test/current', content_type => 'application/x-case', data => qq#[{"mime":"text/plain","value":5},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch2"}]#, version => 2 },
);
$tree->finalize();
is($tree->get('/test/current/value')->value, 18, "/test/current/value has correct value after /test/current modification");

is($tree->get('/test/mod')->value, 18, "/test/mod use server");
$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 14, name => 'mod', path => '/test/mod', content_type => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2/value"}]#, version => 2 },
);
$tree->finalize();
is($tree->get('/test/mod')->value, 16, "/test/mod use datacenter");
$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 14, name => 'mod', path => '/test/mod', content_type => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"xxx","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2/value"}]#, version => 3 },
);
$tree->finalize();
is($tree->get('/test/mod')->value, 8, "/test/mod use default");
$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 14, name => 'mod', path => '/test/mod', content_type => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2/value"}]#, version => 4 },
);
$tree->finalize();
is($tree->get('/test/mod')->value, 16, "/test/mod use datacenter back");
$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 14, name => 'mod', path => '/test/mod', content_type => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"$hostname","mime":"application/x-symlink","value":"/test/branch2/value"}]#, version => 5 },
);
$tree->finalize();
is($tree->get('/test/mod')->value, 18, "/test/mod use server back");

is($tree->get('/test/curmod/value')->value, 18, "/test/curmod/value use server");
$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 15, name => 'curmod', path => '/test/curmod', content_type => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2"}]#, version => 2 },
);
$tree->finalize();
is($tree->get('/test/curmod/value')->value, 16, "/test/curmod/value use datacenter");
$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 15, name => 'curmod', path => '/test/curmod', content_type => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test"},{"datacenter":"xxx","mime":"application/x-symlink","value":"/test/branch1"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2"}]#, version => 3 },
);
$tree->finalize();
is($tree->get('/test/curmod/value')->value, 8, "/test/curmod/value use default");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 14, name => 'mod', path => '/test/mod', content_type => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2/value"},{"group":"alei","mime":"application/x-symlink","value":"/test/branch3/value"}]#, version => 6 },
);
$tree->finalize();
is($tree->get('/test/mod')->value, 20, "/test/mod use group");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 14, name => 'mod', path => '/test/mod', content_type => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"datacenter":"test-dc","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"alei*.mail.ru","mime":"application/x-symlink","value":"/test/branch2/value"},{"group":"alei","mime":"application/x-symlink","value":"/test/branch3/value"}]#, version => 7 },
);
$tree->finalize();
is($tree->get('/test/mod')->value, 18, "/test/mod use server");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 14, name => 'mod', path => '/test/mod', content_type => 'application/x-case', data => qq#[{"mime":"application/x-symlink","value":"/test/value"},{"group":"stat","mime":"application/x-symlink","value":"/test/branch1/value"},{"server":"xxx","mime":"application/x-symlink","value":"/test/branch2/value"},{"group":"alei","mime":"application/x-symlink","value":"/test/branch3/value"}]#, version => 8 },
);
$tree->finalize();
is($tree->get('/test/mod')->value, 16, "/test/mod use another group");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 111, name => 'stat', path => '/infrastructure/group/stat', content_type => 'text/plain', data => 'stat*.mail.ru', version => 2 },
);
$tree->finalize();
is($tree->get('/test/mod')->value, 20, "when group: change group definition");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 111, name => 'stat', path => '/infrastructure/group/stat', content_type => 'text/plain', data => 'alei*.mail.ru', version => 3 },
);
$tree->finalize();
is($tree->get('/test/mod')->value, 16, "when group: change group definition back");

$tree->delete(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 111, name => 'stat', path => '/infrastructure/group/stat', content_type => 'text/plain', data => 'alei*.mail.ru', version => 4 },
);
$tree->finalize();
is($tree->get('/test/mod')->value, 20, "when group: remove group definition");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 111, name => 'stat', path => '/infrastructure/group/stat', content_type => 'text/plain', data => 'alei*.mail.ru', version => 5 },
);
$tree->finalize();
is($tree->get('/test/mod')->value, 16, "when group: add group definition");
