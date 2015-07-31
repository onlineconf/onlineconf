package MR::OnlineConf::Updater::Admin;

use Mouse;

# External modules
use CBOR::XS;
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

has mtime => (
    is => 'rw',
    isa => 'Str',
    default => '1',
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

has version => (
    is => 'ro',
    isa => 'Str',
    required => 1,
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

has hostname => (
    is => 'ro',
    isa => 'Str',
    lazy => 1,
    default => sub {
        Sys::Hostname::hostname()
    },
);

has lwp => (
    is => 'ro',
    isa => 'LWP::UserAgent',
    lazy => 1,
    default => sub {
        my ($self) = @_;
        my $ua = LWP::UserAgent->new();

        $ua->default_header("X-OnlineConf-Client-Host" => $self->hostname);
        $ua->default_header("X-OnlineConf-Client-Version" => $self->version);

        $ua->default_header("Authorization" => "Basic " . MIME::Base64::encode_base64(
            $self->username . ':' . $self->password
        ));

        $ua->add_handler(
            request_prepare => sub {
                my ($req, $ua, $h) = @_;
                $req->header('X-OnlineConf-Client-Mtime' => $self->mtime);
            }
        );

        return $ua;
    }
);

sub get_config {
    my ($self) = @_;
    my $res = $self->lwp->get($self->address . 'client/config');

    return [] unless $res;

    warn $res->status_line;

    if ($res->headers->header('X-OnlineConf-Admin-Last-Modified') gt $self->mtime) {
        $self->mtime(
            $res->headers->header('X-OnlineConf-Admin-Last-Modified')
        );
    }

    if ($res->is_success) {
        my $data;

        eval {
            $data = CBOR::XS::decode_cbor(
                $res->content
            );
        };

        if ($@) {
            warn "WebAPI CBOR ERROR: $@";
        }

        return $data || [];
    }

    return [];
}

sub post_activity {
    my ($self) = @_;

    return $self->lwp->post(
        $self->address . 'client/activity', {
            host => $self->hostname,
            mtime => $self->mtime,
            version => $self->version,
        }
    );
}

1;
