package MR::OnlineConf::Admin::Group;

use Mouse;

has name => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
);

has id => (
    is  => 'ro',
    isa => 'Int',
    lazy    => 1,
    default => sub { die sprintf "Group not exists: %s\n", $_[0]->name unless $_[0]->_row; $_[0]->_row->{ID} },
);

has users => (
    is  => 'ro',
    isa => 'ArrayRef[Str]',
    lazy_build => 1,
);

has _row => (
    is  => 'ro',
    isa => 'Maybe[HashRef]',
    lazy_build => 1,
);

around BUILDARGS => sub {
    my $orig = shift;
    my $class = shift;
    if (@_ == 1 && !ref($_[0])) {
        return $class->$orig(name => $_[0]);
    } else {
        return $class->$orig(@_);
    }
};

sub can_edit {
    my ($class, $user) = @_;
    my $list = MR::OnlineConf::Admin::Storage->select('
        SELECT *
        FROM `my_config_group` g
        JOIN `my_config_user_group` ug ON ug.`GroupID` = g.`ID`
        WHERE g.`Name` = ? AND ug.`User` = ?
    ', 'root', $user);
    return scalar @$list;
}

sub list {
    my ($class) = @_;
    return [ map MR::OnlineConf::Admin::Group->new(name => $_->{Name}, _row => $_),
        @{MR::OnlineConf::Admin::Storage->select('SELECT * FROM `my_config_group` ORDER BY `Name`')} ];
}

sub create {
    my ($self) = @_;
    MR::OnlineConf::Admin::Storage->do('INSERT INTO `my_config_group` (`Name`) VALUES (?)', $self->name);
    return;
}

sub delete {
    my ($self) = @_;
    MR::OnlineConf::Admin::Storage->do('DELETE FROM `my_config_group` WHERE `Name` = ?', $self->name);
    return;
}

sub add_user {
    my ($self, $user) = @_;
    MR::OnlineConf::Admin::Storage->do('INSERT INTO `my_config_user_group` (`GroupID`, `User`) VALUES (?, ?)', $self->id, $user);
    return;
}

sub delete_user {
    my ($self, $user) = @_;
    MR::OnlineConf::Admin::Storage->do('DELETE FROM `my_config_user_group` WHERE `GroupID` = ? AND `User` = ?', $self->id, $user);
    return;
}

sub _build_users {
    my ($self) = @_;
    return [ map $_->{User}, @{MR::OnlineConf::Admin::Storage->select('SELECT * FROM `my_config_user_group` WHERE `GroupID` = ? ORDER BY `User`', $self->id)} ];
}

sub _build__row {
    my ($self) = @_;
    return MR::OnlineConf::Admin::Storage->select('SELECT * FROM `my_config_group` WHERE `Name` = ?', $self->name)->[0];
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
