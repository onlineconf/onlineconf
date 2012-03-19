package MR::OnlineConf::Admin::Web::Monitoring;
use Mojo::Base 'MR::OnlineConf::Admin::Web::Controller';
use MR::OnlineConf::Admin::Monitoring;

sub list {
    my ($self) = @_;
    my $list = MR::OnlineConf::Admin::Monitoring->list();
    $self->render(json => [ map $self->_host($_), @$list ]);
    return;
}

sub _host {
    my ($self, $host) = @_;
    return {
        host         => $host->host,
        mtime        => $host->mtime,
        online       => $host->online,
        package      => $host->package,
        mtime_alert  => $host->mtime_alert ? Mojo::JSON->true : Mojo::JSON->false,
        online_alert => $host->online_alert ? Mojo::JSON->true : Mojo::JSON->false,
    };
}

1;
