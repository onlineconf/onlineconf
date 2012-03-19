package MR::OnlineConf::Admin::Access;

use Mouse;
extends 'MR::OnlineConf::Admin::Group';

has node_id => (
    is  => 'ro',
    isa => 'Int',
    required => 1,
);

has overridden => (
    is  => 'ro',
    isa => 'Bool',
    lazy    => 1,
    default => sub { die sprintf "Access not exists: %s\n", $_[0]->name unless $_[0]->_row; $_[0]->_row->{Overridden} },
);

has rw => (
    is  => 'ro',
    isa => 'Bool',
    lazy    => 1,
    default => sub { die sprintf "Access not exists: %s\n", $_[0]->name unless $_[0]->_row; $_[0]->_row->{RW} },
);

sub select_by_node_id {
    my ($class, $node_id) = @_;
    return [
        map MR::OnlineConf::Admin::Access->new(name => $_->{Name}, node_id => $node_id, _row => $_),
            @{MR::OnlineConf::Admin::Storage->select('
                SELECT `ID`, `Name`,
                    IF(tg.`GroupID` IS NOT NULL, tg.`RW`, `my_config_tree_group_access`(?, g.`ID`)) AS `RW`,
                    tg.`GroupID` IS NOT NULL AS `Overridden`
                FROM `my_config_group` g
                LEFT JOIN `my_config_tree_group` tg ON tg.`GroupID` = g.`ID` AND tg.`NodeID` = ?
                ORDER BY g.`Name`
            ', $node_id, $node_id)}
    ];
}

sub set {
    my ($class, $node_id, $group, $rw) = @_;
    MR::OnlineConf::Admin::Storage->do('
        INSERT INTO `my_config_tree_group` (`NodeID`, `GroupID`, `RW`)
        VALUES (?, (SELECT `ID` FROM `my_config_group` WHERE `Name` = ?), ?)
        ON DUPLICATE KEY UPDATE `RW` = VALUES(`RW`)
    ', $node_id, $group, !defined $rw ? undef : $rw ? 1 : 0);
    return;
}

sub delete {
    my ($class, $node_id, $group) = @_;
    MR::OnlineConf::Admin::Storage->do('DELETE FROM `my_config_tree_group` WHERE `NodeID` = ? AND `GroupID` = (SELECT `ID` FROM `my_config_group` WHERE `Name` = ?)', $node_id, $group);
    return;
}

sub delete_all {
    my ($class, $node_id) = @_;
    MR::OnlineConf::Admin::Storage->do('DELETE FROM `my_config_tree_group` WHERE `NodeID` = ?', $node_id);
    return;
}

sub _build__row {
    my ($self) = @_;
    return MR::OnlineConf::Admin::Storage->select('
        SELECT `ID`, `Name`,
            IF(tg.`GroupID` IS NOT NULL, tg.`RW`, `my_config_tree_group_access`(?, g.`ID`)) AS `RW`,
            tg.`GroupID` IS NOT NULL AS `Overridden`
        FROM `my_config_group` g
        LEFT JOIN `my_config_tree_group` tg ON tg.`GroupID` = g.`ID` AND tg.`NodeID` = ?
        WHERE g.`Name` = ?
    ', $self->node_id, $self->node_id, $self->name)->[0];
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
