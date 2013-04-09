package MR::OnlineConf::Updater::Util;

use Mouse;
use Mouse::Exporter;
use Sys::Hostname;
use Text::Glob;

Mouse::Exporter->setup_import_methods(
    as_is => ['hostname_match_glob'],
);

sub match_glob_strict_wildcard_dot ($@) {
    my $glob = shift;
    local $Text::Glob::strict_leading_dot = 0;
    local $Text::Glob::strict_wildcard_slash = 1;
    my $re_str = Text::Glob::glob_to_regex_string($glob);
    $re_str =~ s/\Q[^\/]\E/[^\\.]/g;
    my $re = qr/^$re_str$/;
    return grep { $_ =~ $re } @_;
}

sub hostname_match_glob ($) {
    return match_glob_strict_wildcard_dot(shift, hostname());
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
