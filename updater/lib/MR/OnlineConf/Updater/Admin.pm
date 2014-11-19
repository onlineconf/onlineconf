package MR::OnlineConf::Updater::Admin;

use Mouse;

# External modules
use JSON::XS;
use MIME::Base64;
use Sys::Hostname;
use LWP::UserAgent;

has host => (
    is => 'ro',
    isa => 'Str',
    required => 1,
);

has port => (
    is => 'ro',
    isa => 'Int',
    required => 1,
);

has address => (
    is => 'ro',
    isa => 'Str',
    lazy => 1,
    default => sub {
        my ($self) = @_;

        return 'http://' . $self->host . ':' . $self->port . '/';
    }
);

has username => (
    is => 'ro',
    isa => 'Str',
    required => 1,
);

has password => (
    is => 'ro',
    isa => 'Str',
    required => 1,
);

has lwp => (
    is => 'ro',
    isa => 'LWP::UserAgent',
    lazy => 1,
    default => sub {
        my ($self) = @_;
        my $ua = LWP::UserAgent->new();

        $ua->default_header("X-Updater-Host" => Sys::Hostname::hostname());

        $ua->default_header("Authorization" => "Basic " . MIME::Base64::encode_base64(
            $self->username . ':' . $self->password
        ));

        return $ua;
    }
);

sub get_mtime {
    my ($self) = @_;
    my $res = _decode_res(
        $self->lwp->get($self->address . 'updater/mtime')
    );

    return $res ? $res->{MTime} : undef;
}

sub get_config {
    my ($self, $mtime, $reselect) = @_;

    if ($mtime && $reselect) {
        return _decode_res(
            $self->lwp->get($self->address . "updater/config/$mtime/$reselect")
        );
    }

    return _decode_res(
        $self->lwp->get($self->address . 'updater/config')
    );
}

sub post_activity {
    my ($self, $host, $mtime, $version) = @_;

    $self->lwp->post(
        $self->address . 'updater/activity', {
            host => $host,
            mtime => $mtime,
            version => $version,
        }
    );
}

# Internal methods

sub _decode_res {
    my ($res) = @_;

    return unless $res;

    if ($res->is_success) {
        my $data;

        eval {
            $data = JSON::XS::decode_json(
                $res->decoded_content
            );
        };

        if ($@) {
            warn "WebAPI JSON ERROR: $@";
        }

        return $data;
    }

    warn $res->status_line;

    return;
}

1;
