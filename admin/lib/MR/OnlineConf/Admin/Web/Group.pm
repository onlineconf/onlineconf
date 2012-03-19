package MR::OnlineConf::Admin::Web::Group;
use Mojo::Base 'MR::OnlineConf::Admin::Web::Controller';
use MR::OnlineConf::Admin::Group;

sub can_edit {
    my ($self) = @_;
    if ($self->req->method ne 'POST' && $self->req->method ne 'DELETE') {
        return 1;
    } elsif (MR::OnlineConf::Admin::Group->can_edit($self->username)) {
        return 1;
    } else {
        $self->render_access_denied();
        return;
    }
}

sub list {
    my ($self) = @_;
    my $list = MR::OnlineConf::Admin::Group->list();
    $self->render(json => [ map $_->name, @$list ]);
    return;
}

sub get {
    my ($self) = @_;
    my $group = MR::OnlineConf::Admin::Group->new($self->param('group'));
    $self->render(json => $group->users);
    return;
}

sub create {
    my ($self) = @_;
    my $group = MR::OnlineConf::Admin::Group->new($self->param('group'));
    $group->create();
    $self->render(json => { result => 'Created' });
    return;
}

sub delete {
    my ($self) = @_;
    my $group = MR::OnlineConf::Admin::Group->new($self->param('group'));
    $group->delete();
    $self->render(json => { result => 'Deleted' });
    return;
}

sub add_user {
    my ($self) = @_;
    my $group = MR::OnlineConf::Admin::Group->new($self->param('group'));
    $group->add_user($self->param('user'));
    $self->render(json => { result => 'Created' });
    return;
}

sub delete_user {
    my ($self) = @_;
    my $group = MR::OnlineConf::Admin::Group->new($self->param('group'));
    $group->delete_user($self->param('user'));
    $self->render(json => { result => 'Deleted' });
    return;
}

sub user_list {
    my ($self) = @_;
    $self->render(json => $self->users($self->param('term')));
    return;
}

sub whoami {
    my ($self) = @_;
    $self->render(json => {
        username        => $self->username,
        can_edit_groups => MR::OnlineConf::Admin::Group->can_edit($self->username) ? Mojo::JSON->true : Mojo::JSON->false,
    });
    return;
}

1;
