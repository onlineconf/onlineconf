package MR::OnlineConf::Admin::Web::Updater;

use Mojo::Base 'MR::OnlineConf::Admin::Web::Controller';

# External modules
use Socket;
use CBOR::XS;

# Internal modules
use MR::OnlineConf::Admin::Storage;
use MR::OnlineConf::Admin::PerlMemory;

my $tree = MR::OnlineConf::Admin::PerlMemory->new();

sub validate {
    my ($self) = @_;
    my $ip = $self->req->env->{HTTP_X_REAL_IP} || $self->req->env->{REMOTE_ADDR};

    unless ($ip) {
        $self->app->log->warn('Remote ip not found');
        $self->render(text => '', status => 400);
        return 0;
    }

    my ($hostname, $aliases) = gethostbyaddr(inet_aton($ip), AF_INET);

    unless ($hostname || $aliases) {
        $self->app->log->warn("Cant lookup hostname for $ip");
        $self->render(text => '', status => 400);
        return 0;
    }

    my $host = $self->param('host');

    $aliases  ||= '';
    $hostname ||= '';

    $self->stash(ip => $ip);

    return 1 if $host eq $hostname;
    return 1 if grep { $host eq $_ } split ' ', $aliases;

    $self->render(text => '', status => 400);
    $self->app->log->warn("Hacker $ip $host $hostname $aliases");

    return 0;
}

sub mtime {
    my ($self) = @_;
    my $list = MR::OnlineConf::Admin::Storage->select(qq[
        SELECT
            MAX(`MTime`) AS `MTime`
        FROM
            `my_config_tree_log`
    ]);

    $list ||= [{}];

    $self->render(json => $list->[0]);
}

sub config {
    my ($self) = @_;
    my $ip = $self->stash('ip');
    my $host = $self->param('host');
    my $mtime = $self->param('mtime');
    my $reselect = $self->param('reselect');

    $tree->refresh();
    $tree->host($host);
    $tree->addr([$ip]);

    $self->render(
        data => CBOR::XS::encode_cbor(
            $tree->serialize()
        )
    );
}

sub activity {
    my ($self) = @_;
    my $host = $self->param('host');
    my $mtime = $self->param('mtime');
    my $version = $self->param('version');

    MR::OnlineConf::Admin::Storage->do(qq[
            REPLACE INTO
                `my_config_activity` (`Host`, `Time`, `Online`, `Package`)
            VALUES
                (?, ?, now(), ?)
        ],
        $host,
        $mtime || 0,
        $version
    );

    $self->render(json => {success => 1});
}

1;
