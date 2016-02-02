#!/usr/bin/perl

use strict;
use warnings;

use Test::More;

use Sys::Hostname;
use MR::OnlineConf::Admin::PerlMemory;

my $host = hostname();

my @match_glob = (
    'alei*.mydev.mail.ru',
    'alei*.mydev.*.*',
    'alei*.{mydev.mail.ru}',
    'alei*.{mydev.mail.ru,mail.ru}',
);

my @not_match_glob = (
    'alei*.mydev.*',
);

plan tests => scalar @match_glob + scalar @not_match_glob;

foreach my $glob (@match_glob) {
    ok(MR::OnlineConf::Admin::PerlMemory::hostname_match_glob($glob, $host), "match $glob - $host");
}

foreach my $glob (@not_match_glob) {
    ok(!MR::OnlineConf::Admin::PerlMemory::hostname_match_glob($glob, $host), "not match $glob - $host");
}
