package MR::OnlineConf::Admin::Version;

use utf8;
use Mouse;
use MR::ChangeBot::Notification;

has node_id => (
    is  => 'ro',
    isa => 'Int',
    required => 1,
);

has path => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
);

has mime => (
    is  => 'ro',
    isa => 'Str',
    lazy   => 1,
    default => sub { die sprintf "Version %s of %s not exists\n", $_[0]->version, $_[0]->node_id unless $_[0]->_row; $_[0]->_row->{ContentType} },
);

has data => (
    is  => 'ro',
    isa => 'Maybe[Str]',
    lazy   => 1,
    default => sub { die sprintf "Version %s of %s not exists\n", $_[0]->version, $_[0]->node_id unless $_[0]->_row; defined($_[0]->rw) ? $_[0]->_row->{Value} : undef },
);

has version => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
);

has author => (
    is  => 'ro',
    isa => 'Str',
    lazy    => 1,
    default => sub { die sprintf "Version %s of %s not exists\n", $_[0]->version, $_[0]->node_id unless $_[0]->_row; $_[0]->_row->{Author} },
);

has mtime => (
    is  => 'ro',
    isa => 'Str',
    lazy    => 1,
    default => sub { die sprintf "Version %s of %s not exists\n", $_[0]->version, $_[0]->node_id unless $_[0]->_row; $_[0]->_row->{MTime} },
);

has deleted => (
    is  => 'ro',
    isa => 'Bool',
    lazy    => 1,
    default => sub { die sprintf "Version %s of %s not exists\n", $_[0]->version, $_[0]->node_id unless $_[0]->_row; $_[0]->_row->{Deleted} },
);

has comment => (
    is  => 'ro',
    isa => 'Maybe[Str]',
    lazy    => 1,
    default => sub { die sprintf "Version %s of %s not exists\n", $_[0]->version, $_[0]->node_id unless $_[0]->_row; $_[0]->_row->{Comment} },
);

has username => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
);

has rw => (
    is  => 'ro',
    isa => 'Bool',
    lazy    => 1,
    default => sub { die sprintf "Version %s of %s not exists\n", $_[0]->version, $_[0]->node_id unless $_[0]->_row; $_[0]->_row->{RW} },
);

has _row => (
    is  => 'ro',
    isa => 'Maybe[HashRef]',
    lazy    => 1,
    default => sub {
        my ($self) = @_;
        return MR::OnlineConf::Admin::Storage->select(
            'SELECT *, `my_config_tree_access`(`NodeID`, ?) AS `RW` FROM `my_config_tree_log` WHERE `NodeID` = ? AND `Version` = ?',
            $self->username, $self->node_id, $self->version
        )->[0];
    },
    clearer => '_clear_row',
);

sub select_by_node {
    my ($class, $node) = @_;
    return [ map MR::OnlineConf::Admin::Version->new(node_id => $_->{NodeID}, path => $node->path, version => $_->{Version}, username => $node->username, rw => $_->{RW}, _row => $_),
        @{MR::OnlineConf::Admin::Storage->select('SELECT *, `my_config_tree_access`(`NodeID`, ?) AS `RW` FROM `my_config_tree_log` WHERE `NodeID` = ? ORDER BY `Version` DESC', $node->username, $node->id)} ];
}

sub select {
    my ($class, %in) = @_;
    my (@condition, @bind);
    if (my $author = $in{author}) {
        push @condition, 'l.`Author` = ?';
        push @bind, $author;
    }
    if (my $branch = $in{branch}) {
        $branch =~ s/([%_])/\\$1/g;
        push @condition, 't.`Path` LIKE ?';
        push @bind, "$branch%";
    }
    if (my $from = $in{from}) {
        push @condition, 'l.`MTime` >= ?';
        push @bind, $from;
    }
    if (my $till = $in{till}) {
        push @condition, $till =~ /^\d{4}-\d{2}-\d{2}$/ ? 'l.`MTime` < ? + interval 1 day' : 'l.`MTime` < ?';
        push @bind, $till;
    }
    push @condition, "t.`Path` <> '/onlineconf/selftest/update-time'";
    my $where = @condition ? ('WHERE ' . join(' AND ', @condition)): '';
    return [
        map MR::OnlineConf::Admin::Version->new(node_id => $_->{NodeID}, path => $_->{Path}, version => $_->{Version}, username => $in{username}, rw => $_->{RW}, _row => $_),
            @{MR::OnlineConf::Admin::Storage->select("
                SELECT l.*, t.`Path`, `my_config_tree_access`(t.`ID`, ?) AS `RW`
                FROM `my_config_tree_log` l JOIN `my_config_tree` t ON t.`ID` = l.`NodeID`
                $where
                ORDER BY l.`MTime` DESC
                LIMIT 50
            ", $in{username}, @bind)}
    ];
}

sub log {
    my ($self) = @_;
    MR::OnlineConf::Admin::Storage->do('INSERT INTO `my_config_tree_log` (`NodeID`, `Version`, `ContentType`, `Value`, `Author`, `MTime`, `Comment`, `Deleted`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)',
        $self->node_id, $self->version, $self->mime, $self->data, $self->author, $self->mtime, $self->comment, $self->deleted ? 1 : 0);
    my $message = sprintf "%s %s параметр %s.", $self->author, $self->deleted ? "удалил" : "изменил", $self->path;
    MR::ChangeBot::Notification->new(origin => 'onlineconf', message => $message)->create();
    return;
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
