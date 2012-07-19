package MR::OnlineConf::Admin::Web::Config;
use Mojo::Base 'MR::OnlineConf::Admin::Web::Controller';
use MR::OnlineConf::Admin::Parameter;
use File::Spec::Unix;

sub get {
    my ($self) = @_;
    my $node = MR::OnlineConf::Admin::Parameter->new($self->_path, $self->username);
    if (defined $self->param('symlink')) {
        $self->param('symlink') eq 'follow' ? $node->follow_symlink() : $node->resolve_symlink();
    }
    my $result = $self->_node($node);
    $result->{children} = [ map $self->_node($_), @{$node->children} ];
    $self->render(json => $result);
    return;
}

sub set {
    my ($self) = @_;
    my @params = sort $self->param();
    if (@params == 1 && $params[0] eq 'notification') {
        my $node = MR::OnlineConf::Admin::Parameter->new($self->_path, $self->username);
        $node->notification(scalar $self->param('notification') || undef);
        $node->update();
        $self->render(json => { result => 'NotificationChanged', %{$self->_node($node)} });
    } elsif (@params == 2 && $params[0] eq 'description' && $params[1] eq 'summary') {
        my $node = MR::OnlineConf::Admin::Parameter->new($self->_path, $self->username);
        $node->summary(scalar $self->param('summary'));
        $node->description(scalar $self->param('description'));
        $node->update();
        $self->render(json => { result => 'Renamed', %{$self->_node($node)} });
    } elsif (defined $self->param('path')) {
        die "Comment is required\n" unless defined $self->param('comment') && length $self->param('comment');
        my $node = MR::OnlineConf::Admin::Parameter->new($self->_path, $self->username);
        $node->move(
            path    => scalar $self->param('path'),
            symlink => scalar $self->param('symlink'),
            version => scalar $self->param('version'),
            comment => scalar $self->param('comment'),
        );
        $self->render(json => { result => 'Moved', %{$self->_node($node)} });
    } elsif (defined $self->param('version')) {
        die "Comment is required\n" unless defined $self->param('comment') && length $self->param('comment');
        my $node = MR::OnlineConf::Admin::Parameter->new($self->_path, $self->username);
        $node->mime(scalar $self->param('mime'));
        $node->data(scalar $self->param('data'));
        $node->update(version => scalar $self->param('version'), comment => scalar $self->param('comment'));
        $self->render(json => { result => 'Changed', %{$self->_node($node)} });
    } else {
        die "Comment is required\n" unless defined $self->param('comment') && length $self->param('comment');
        my $node = MR::OnlineConf::Admin::Parameter->new(
            path         => $self->_path,
            mime         => scalar $self->param('mime'),
            data         => scalar $self->param('data'),
            summary      => scalar $self->param('summary'),
            description  => scalar $self->param('description'),
            notification => scalar $self->param('notification'),
            username     => $self->username,
        );
        $node->create(comment => scalar $self->param('comment'));
        $self->render(json => { result => 'Created', %{$self->_node($node)} });
    }
    return;
}

sub delete {
    my ($self) = @_;
    die "Comment is required\n" unless defined $self->param('comment') && length $self->param('comment');
    my $node = MR::OnlineConf::Admin::Parameter->new($self->_path, $self->username);
    $node->delete(version => scalar $self->param('version'), comment => scalar $self->param('comment'));
    $self->render(json => { result => 'Deleted' });
    return;
}

sub search {
    my ($self) = @_;
    my $nodes = MR::OnlineConf::Admin::Parameter->search(scalar $self->param('term'), $self->username);
    $self->render(json => [ map $self->_node($_), @$nodes ]);
    return;
}

sub _path {
    my ($self) = @_;
    return File::Spec::Unix->catdir('/', $self->stash('path'));
}

sub _node {
    my ($self, $node) = @_;
    return {
        name                  => $node->name,
        path                  => $node->path,
        data                  => $node->data,
        version               => $node->version + 0,
        mtime                 => $node->mtime,
        summary               => $node->summary,
        description           => $node->description,
        mime                  => $node->mime,
        num_children          => $node->num_children + 0,
        access_modified       => $node->access_modified ? Mojo::JSON->true : Mojo::JSON->false,
        rw                    => !defined($node->rw) ? undef : $node->rw ? Mojo::JSON->true : Mojo::JSON->false,
        notification          => $node->notification,
        notification_modified => $node->notification_modified ? Mojo::JSON->true : Mojo::JSON->false,
    };
}

1;
