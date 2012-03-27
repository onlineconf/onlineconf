package MR::OnlineConf::Admin::Web::Log;
use Mojo::Base 'MR::OnlineConf::Admin::Web::Controller';
use MR::OnlineConf::Admin::Parameter;
use MR::OnlineConf::Admin::Version;
use File::Spec::Unix;

sub list {
    my ($self) = @_;
    my $node = MR::OnlineConf::Admin::Parameter->new($self->_path, $self->username);
    my $versions = $node->versions;
    $self->render(json => [ map $self->_version($_), @$versions ]);
    return;
}

sub global {
    my ($self) = @_;
    my $versions = MR::OnlineConf::Admin::Version->select(
        username => $self->username,
        author   => scalar $self->param('author'),
        branch   => scalar $self->param('branch'),
        from     => scalar $self->param('from'),
        till     => scalar $self->param('till'),
    );
    $self->render(json => [ map $self->_version($_), @$versions ]);
    return;
}

sub _path {
    my ($self) = @_;
    return File::Spec::Unix->catdir('/', $self->stash('path'));
}

sub _version {
    my ($self, $version) = @_;
    return {
        path    => $version->path,
        version => $version->version,
        mime    => $version->mime,
        data    => $version->data,
        author  => $version->author,
        mtime   => $version->mtime,
        deleted => $version->deleted ? Mojo::JSON->true : Mojo::JSON->false,
        comment => $version->comment,
        rw      => !defined($version->rw) ? undef : $version->rw ? Mojo::JSON->true : Mojo::JSON->false,
    };
}

1;
