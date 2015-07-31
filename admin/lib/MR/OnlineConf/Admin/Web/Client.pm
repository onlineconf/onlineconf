package MR::OnlineConf::Admin::Web::Client;

use Mojo::Base 'MR::OnlineConf::Admin::Web::Controller';

# External modules
use Socket;
use CBOR::XS;

# Internal modules
use MR::OnlineConf::Admin::Storage;
use MR::OnlineConf::Admin::PerlMemory;

my $tree;

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

    my $host = $self->req->headers->header('X-OnlineConf-Client-Host');
    my $mtime = $self->req->headers->header('X-OnlineConf-Client-Mtime');
    my $version = $self->req->headers->header('X-OnlineConf-Client-Version');

    unless ($host) {
        $self->app->log->warn("Unknow client host $ip");
        $self->render(text => '', status => 400);
        return 0;
    }

    unless ($mtime) {
        $self->app->log->warn("Unknow client mtime $ip");
        $self->render(text => '', status => 400);
        return 0;
    }

    unless ($version) {
        $self->app->log->warn("Unknow client version $ip");
        $self->render(text => '', status => 400);
        return 0;
    }

    $aliases  ||= '';
    $hostname ||= '';

    $self->stash(ip => $ip);
    $self->stash(host => $host);
    $self->stash(mtime => $mtime);
    $self->stash(version => $version);

    return 1 if $host eq $hostname;
    return 1 if grep { $host eq $_ } split ' ', $aliases;

    $self->render(text => '', status => 400);
    $self->app->log->warn("Hacker $ip $host $hostname $aliases");

    return 0;
}

sub config {
    my ($self) = @_;
    my $ip = $self->stash('ip');
    my $host = $self->stash('host');
    my $mtime = $self->stash('mtime');

    $tree ||= MR::OnlineConf::Admin::PerlMemory->new();
    $tree->refresh();

    $self->res->headers->header(
        'X-OnlineConf-Admin-Last-Modified' => $tree->mtime
    );

    if ($mtime ge $tree->mtime) {
        return $self->render(
            text => '',
            status => 304,
        );
    }

    $tree->host($host);
    $tree->addr([$ip]);

    $self->render(
        data => CBOR::XS::encode_cbor(
            $tree->serialize($mtime)
        )
    );
}

sub activity {
    my ($self) = @_;
    my $host = $self->stash('host');
    my $mtime = $self->stash('mtime');
    my $version = $self->stash('version');

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
