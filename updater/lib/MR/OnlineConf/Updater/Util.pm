package MR::OnlineConf::Updater::Util;

use Mouse;
use Mouse::Exporter;
use Sys::Hostname;
use Text::Glob;

Mouse::Exporter->setup_import_methods(
    as_is => ['hostname_match_glob', 'expand_template_macro'],
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

sub expand_template_macro ($) {
    my $macro = shift;
    if ($macro eq 'hostname') {
        return hostname();
    } elsif ($macro eq 'hostname -s' || $macro eq 'short_hostname') {
        return hostname_s();
    } elsif ($macro eq 'hostname -i' || $macro eq 'ip') {
        return hostname_i();
    }
    return;
}

{
    my ($hostname_s, $hostname_i);
    sub hostname_s () { $hostname_s ||= do { my $h = `hostname -s`; chomp $h; $h } }
    sub hostname_i () { $hostname_i ||= do { my $h = `hostname -i`; chomp $h; $h } }
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
