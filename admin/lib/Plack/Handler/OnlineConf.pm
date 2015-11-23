package Plack::Handler::OnlineConf;

use strict;
use warnings;

use base 'Starman::Server';
 
sub run {
    my($self, $app) = @_;
    my $path = $ENV{ONLINECONF_CONFIG_PATH} || '/usr/local/etc/';
    my $config = YAML::LoadFile($path . 'onlineconf-admin.yaml');

    return $self->SUPER::run($app, {workers => $config->{workers} || 5});
}
 
1;
