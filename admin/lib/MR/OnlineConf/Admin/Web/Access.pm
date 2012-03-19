package MR::OnlineConf::Admin::Web::Access;
use Mojo::Base 'MR::OnlineConf::Admin::Web::Controller';
use MR::OnlineConf::Admin::Parameter;
use MR::OnlineConf::Admin::Version;
use File::Spec::Unix;

sub list {
    my ($self) = @_;
    my $node = MR::OnlineConf::Admin::Parameter->new($self->_path, $self->username);
    $self->render(json => [ map $self->_access($_), @{$node->access} ]);
    return;
}

sub set {
    my ($self) = @_;
    my $node = MR::OnlineConf::Admin::Parameter->new($self->_path, $self->username);
    my $rw = $self->param('rw');
    $rw = $rw eq 'true' ? 1 : $rw eq 'false' ? 0 : undef;
    $node->set_access($self->param('group'), $rw);
    $self->render(json => { result => 'Changed' });
    return;
}

sub delete {
    my ($self) = @_;
    my $node = MR::OnlineConf::Admin::Parameter->new($self->_path, $self->username);
    $node->delete_access($self->param('group'));
    my @access = grep { $_->name eq $self->param('group') } @{$node->access};
    my $access = @access ? $self->_access($access[0]) : {};
    $self->render(json => { result => 'Changed', %$access });
    return
}

sub _path {
    my ($self) = @_;
    return File::Spec::Unix->catdir('/', $self->stash('path'));
}

sub _access {
    my ($self, $access) = @_;
    return {
        group      => $access->name,
        overridden => $access->overridden ? Mojo::JSON->true : Mojo::JSON->false,
        rw         => !defined($access->rw) ? undef : $access->rw ? Mojo::JSON->true : Mojo::JSON->false,
    }
}

1;
