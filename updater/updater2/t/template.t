#!/usr/bin/perl

use strict;
use warnings;
use Test::More tests => 16;
use Log::Dispatch;
use MR::OnlineConf::Updater::Parameter;
use MR::OnlineConf::Updater::PerlMemory;

my ($hostname, $short_hostname, $ip) = map { my $h = `$_`; chomp $h; $h } ('hostname', 'hostname -s', 'hostname -i');

my $log = Log::Dispatch->new(outputs => [[ 'Screen', min_level => 'info', newline => 1 ]]);

my $tree = MR::OnlineConf::Updater::PerlMemory->new(log => $log);

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 1, name => '', path => '/', content_type => 'application/x-null', data => undef, version => 1 },
    { id => 2, name => 'test', path => '/test', content_type => 'application/x-null', data => undef, version => 1 },
);
$tree->finalize();

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 3, name => 'value', path => '/test/value', content_type => 'text/plain', data => 3, version => 1 },
    { id => 4, name => 'template1', path => '/test/template1', content_type => 'application/x-template', data => '${hostname}#${short_hostname}#${ip}#${/test/value}', version => 1 },
    { id => 5, name => 'template2', path => '/test/template2', content_type => 'application/x-template', data => '${hostname}#${short_hostname}#${ip}#${/test/symlink}', version => 1 },
    { id => 6, name => 'template3', path => '/test/template3', content_type => 'application/x-template', data => '${hostname}#${short_hostname}#${ip}#${/symlink/value}', version => 1 },
    { id => 7, name => 'template4', path => '/test/template4', content_type => 'application/x-template', data => '${hostname}#${short_hostname}#${ip}#${/symlink/symlink}', version => 1 },
    { id => 51, name => 'symlink', path => '/test/symlink', content_type => 'application/x-symlink', data => '/test/value', version => 1 },
    { id => 52, name => 'symlink', path => '/symlink', content_type => 'application/x-symlink', data => '/test', version => 1 },
);
$tree->finalize();
is($tree->get('/test/template1')->value, "$hostname#$short_hostname#$ip#3", "insert template: correct value 1");
is($tree->get('/test/template2')->value, "$hostname#$short_hostname#$ip#3", "insert template: correct value 2");
is($tree->get('/test/template3')->value, "$hostname#$short_hostname#$ip#3", "insert template: correct value 3");
is($tree->get('/test/template4')->value, "$hostname#$short_hostname#$ip#3", "insert template: correct value 4");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 3, name => 'value', path => '/test/value', content_type => 'text/plain', data => 4, version => 2 },
);
$tree->finalize();
is($tree->get('/test/template1')->value, "$hostname#$short_hostname#$ip#4", "update value: template is updated too 1");
is($tree->get('/test/template2')->value, "$hostname#$short_hostname#$ip#4", "update value: template is updated too 2");
is($tree->get('/test/template3')->value, "$hostname#$short_hostname#$ip#4", "update value: template is updated too 3");
is($tree->get('/test/template4')->value, "$hostname#$short_hostname#$ip#4", "update value: template is updated too 4");

$tree->delete(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 3, name => 'value', path => '/test/value', content_type => 'text/plain', data => 4, version => 3 },
);
$tree->finalize();
is($tree->get('/test/template1')->value, "$hostname#$short_hostname#$ip#", "delete value: template is updated 1");
is($tree->get('/test/template2')->value, "$hostname#$short_hostname#$ip#", "delete value: template is updated 2");
is($tree->get('/test/template3')->value, "$hostname#$short_hostname#$ip#", "delete value: template is updated 3");
is($tree->get('/test/template4')->value, "$hostname#$short_hostname#$ip#", "delete value: template is updated 4");

$tree->put(MR::OnlineConf::Updater::Parameter->new($_)) foreach (
    { id => 3, name => 'value', path => '/test/value', content_type => 'text/plain', data => 5, version => 4 },
);
$tree->finalize();
is($tree->get('/test/template1')->value, "$hostname#$short_hostname#$ip#5", "add value: template is updated 1");
is($tree->get('/test/template2')->value, "$hostname#$short_hostname#$ip#5", "add value: template is updated 2");
is($tree->get('/test/template3')->value, "$hostname#$short_hostname#$ip#5", "add value: template is updated 3");
is($tree->get('/test/template4')->value, "$hostname#$short_hostname#$ip#5", "add value: template is updated 4");
