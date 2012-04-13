package MR::OnlineConf::Admin::Parameter;

use Mouse;
use File::Spec::Unix;
use JSON;
use YAML;
use MR::OnlineConf::Admin::Version;
use MR::OnlineConf::Admin::Access;

my $ASTERISK = join ', ', map "t.`$_`", qw/ID Name ParentID Path Value ContentType Summary Description Version MTime Deleted/;

has path => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
    writer   => '_path',
);

has username => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
);

has id => (
    is  => 'ro',
    isa => 'Int',
    lazy    => 1,
    default => sub { die sprintf "Node not exists: %s\n", $_[0]->path unless $_[0]->_row; $_[0]->_row->{ID} },
    clearer => 'clear_id',
);

has name => (
    is  => 'ro',
    isa => 'Str',
    lazy    => 1,
    default => sub { $_[0]->_row ? $_[0]->_row->{Name} : (File::Spec::Unix->splitdir($_[0]->path))[-1] },
    clearer => 'clear_name',
);

has data => (
    is  => 'rw',
    isa => 'Maybe[Str]',
    lazy      => 1,
    default   => sub { $_[0]->_row && defined($_[0]->rw) ? $_[0]->_row->{Value} : undef },
    predicate => 'has_data',
    clearer   => 'clear_data',
);

has version => (
    is  => 'ro',
    isa => 'Str',
    lazy    => 1,
    default => sub { die sprintf "Node not exists: %s\n", $_[0]->path unless $_[0]->_row; $_[0]->_row->{Version} },
    clearer => 'clear_version',
);

has mtime => (
    is  => 'ro',
    isa => 'Str',
    lazy    => 1,
    default => sub { die sprintf "Node not exists: %s\n", $_[0]->path unless $_[0]->_row; $_[0]->_row->{MTime} },
    clearer => 'clear_mtime',
);

has summary => (
    is  => 'rw',
    isa => 'Str',
    lazy      => 1,
    default   => sub { $_[0]->_row ? $_[0]->_row->{Summary} : '' },
    predicate => 'has_summary',
    clearer   => 'clear_summary',
);

has description => (
    is  => 'rw',
    isa => 'Str',
    lazy      => 1,
    default   => sub { $_[0]->_row ? $_[0]->_row->{Description} : '' },
    predicate => 'has_description',
    clearer   => 'clear_description',
);

has mime => (
    is  => 'rw',
    isa => 'Str',
    lazy      => 1,
    default   => sub { $_[0]->_row ? $_[0]->_row->{ContentType} : 'application/x-null' },
    predicate => 'has_mime',
    clearer   => 'clear_mime',
);

has deleted => (
    is  => 'ro',
    isa => 'Bool',
    lazy    => 1,
    default => sub { $_[0]->_row ? $_[0]->_row->{Deleted} : undef },
    clearer => 'clear_deleted',
);

has rw => (
    is  => 'ro',
    isa => 'Bool',
    lazy    => 1,
    default => sub { $_[0]->_row ? $_[0]->_row->{RW} : undef },
    clearer => 'clear_rw',
);

has access_modified => (
    is  => 'ro',
    isa => 'Bool',
    lazy    => 1,
    default => sub { $_[0]->_row ? $_[0]->_row->{AccessModified} : undef },
    clearer => 'clear_access_modified',
);

has notification => (
    is  => 'rw',
    isa => 'Maybe[Str]',
    lazy    => 1,
    default => sub { $_[0]->_row ? $_[0]->_row->{Notification} : undef },
    trigger => sub { $_[0]->_notification_changed(1) },
    clearer => 'clear_notification',
);

has notification_modified => (
    is  => 'ro',
    isa => 'Bool',
    lazy    => 1,
    default => sub { $_[0]->_row ? $_[0]->_row->{NotificationModified} : undef },
    clearer => 'clear_notification_modified',
);

has num_children => (
    is  => 'ro',
    isa => 'Int',
    lazy    => 1,
    default => sub { die sprintf "Node not exists: %s\n", $_[0]->path unless $_[0]->_row; $_[0]->_row->{NumChildren} },
    clearer => 'clear_num_children',
);

has children => (
    is  => 'ro',
    isa => 'ArrayRef[MR::OnlineConf::Admin::Parameter]',
    lazy_build => 1,
    clearer => 'clear_children',
);

has versions => (
    is  => 'ro',
    isa => 'ArrayRef[MR::OnlineConf::Admin::Version]',
    lazy    => 1,
    default => sub { MR::OnlineConf::Admin::Version->select_by_node($_[0]) },
    clearer => 'clear_versions',
);

has access => (
    is  => 'ro',
    isa => 'ArrayRef[MR::OnlineConf::Admin::Access]',
    lazy_build => 1,
);

has _local_options => (
    is  => 'ro',
    isa => 'HashRef',
    default => sub { {} },
);

has _row => (
    is  => 'rw',
    isa => 'Maybe[HashRef]',
    lazy    => 1,
    default => sub {
        my ($self) = @_;
        my $deleted = $self->_local_options->{load_deleted} ? '' : 'AND NOT `Deleted`';
        my $for_update = $self->_local_options->{for_update} ? 'FOR UPDATE' : '';
        return MR::OnlineConf::Admin::Storage->select("
            SELECT $ASTERISK,
                (SELECT count(*) FROM `my_config_tree` c WHERE c.`ParentID` = t.`ID` AND NOT c.`Deleted`) AS `NumChildren`,
                (SELECT count(*) <> 0 FROM `my_config_tree_group` g WHERE g.`NodeID` = t.`ID`) AS `AccessModified`,
                `my_config_tree_access`(t.`ID`, ?) AS `RW`,
                `my_config_tree_notification`(t.`ID`) AS `Notification`,
                t.`Notification` IS NOT NULL AS `NotificationModified`
            FROM `my_config_tree` t
            WHERE `Path` = ? $deleted
            $for_update
        ", $self->username, $self->path)->[0];
    },
    clearer => '_clear_row',
);

has _notification_changed => (
    is  => 'rw',
    isa => 'Bool',
    default => 0,
    clearer => '_clear_notification_changed',
);

around BUILDARGS => sub {
    my $orig = shift;
    my $class = shift;
    if (@_ == 2 && $_[0] =~ /^\//) {
        return $class->$orig(path => $_[0], username => $_[1]);
    } else {
        return $class->$orig(@_);
    }
};

sub search {
    my ($class, $term, $username) = @_;
    $term =~ s/([%_])/\\$1/g;
    $term = "%$term%";
    return [
        map MR::OnlineConf::Admin::Parameter->new(path => $_->{Path}, username => $username, _row => $_),
            @{MR::OnlineConf::Admin::Storage->select("
                SELECT * FROM (
                    SELECT $ASTERISK,
                        (SELECT count(*) FROM `my_config_tree` c WHERE c.`ParentID` = t.`ID` AND NOT c.`Deleted`) AS `NumChildren`,
                        (SELECT count(*) <> 0 FROM `my_config_tree_group` g WHERE g.`NodeID` = t.`ID`) AS `AccessModified`,
                        `my_config_tree_access`(t.`ID`, ?) AS `RW`,
                        `my_config_tree_notification`(t.`ID`) AS `Notification`,
                        t.`Notification` IS NOT NULL AS `NotificationModified`
                    FROM `my_config_tree` t
                    WHERE NOT `Deleted`
                    AND (`Name` COLLATE ascii_general_ci LIKE ? OR `Value` COLLATE utf8_general_ci LIKE ? OR `Summary` LIKE ? OR `Description` LIKE ?)
                    ORDER BY `Path`
                ) x
                WHERE `RW` IS NOT NULL
                OR (`Name` COLLATE ascii_general_ci LIKE ? OR `Summary` LIKE ? OR `Description` LIKE ?)
            ", $username, map $term, (1 .. 7))}
    ];
}

sub exists {
    my ($self) = @_;
    return defined $self->_row;
}

sub create {
    my ($self, %in) = @_;
    $self->validate();
    MR::OnlineConf::Admin::Storage->transaction(sub {
        my $parent_path = $self->path;
        $parent_path =~ s/\/[^\/]+$//;
        $parent_path ||= '/';
        my $parent = MR::OnlineConf::Admin::Parameter->new($parent_path, $self->username);
        die "Parent node $parent_path not exists\n" unless $parent->exists();
        die "Access denied\n" unless $parent->rw;
        my $current = MR::OnlineConf::Admin::Parameter->new($self->path, $self->username);
        {
            local $current->_local_options->{for_update} = 1;
            local $current->_local_options->{load_deleted} = 1;
            $current->_row;
        }
        if ($current->exists()) {
            die sprintf "Node %s already exists\n", $current->path unless $current->deleted;
            MR::OnlineConf::Admin::Storage->do('UPDATE `my_config_tree` SET `Value` = ?, `Summary` = ?, `Description` = ?, `ContentType` = ?, `Notification` = ?, `Version` = `Version` + 1, MTime = now(), `Deleted` = false WHERE `Path` = ?',
                $self->data, $self->summary, $self->description, $self->mime, $self->notification, $self->path);
        } else {
            $self->_row(undef);
            MR::OnlineConf::Admin::Storage->do('INSERT INTO `my_config_tree` (`ParentID`, `Name`, `Value`, `Summary`, `Description`, `ContentType`, `Notification`) VALUES (?, ?, ?, ?, ?, ?, ?)',
                $parent->id, $self->name, $self->data, $self->summary, $self->description, $self->mime, $self->notification);
        }
        $self->clear();
        $self->_log(comment => $in{comment});
    });
    return;
}

sub update {
    my ($self, %in) = @_;
    $self->validate();
    MR::OnlineConf::Admin::Storage->transaction(sub {
        my %changed;
        {
            my $current = MR::OnlineConf::Admin::Parameter->new($self->path, $self->username);
            local $current->_local_options->{for_update} = 1;
            die sprintf "Node %s not exists\n", $current->path unless $current->exists();
            die sprintf "Version not match: %s != %s\n", $in{version}, $current->version if defined $in{version} && $in{version} != $current->version;
            die "Access denied\n" unless $current->rw;
            $changed{Value} = $self->data if $self->has_data() && ((defined $self->data xor defined $current->data) || $self->data ne $current->data);
            $changed{ContentType} = $self->mime if $self->has_mime() && $self->mime ne $current->mime;
            $changed{Summary} = $self->summary if $self->has_summary() && ((defined $self->summary xor defined $current->summary) || $self->summary ne $current->summary);
            $changed{Description} = $self->description if $self->has_description() && ((defined $self->description xor defined $current->description) || $self->description ne $current->description);
            $changed{Notification} = $self->notification if $self->_notification_changed;
        }
        return unless keys %changed;
        my $inc_version = exists $changed{Value} || exists $changed{ContentType};
        MR::OnlineConf::Admin::Storage->do('UPDATE `my_config_tree` SET ' . join(', ', map("`$_` = ?", keys %changed), $inc_version ? ("`Version` = `Version` + 1", "MTime = now()") : ()) . ' WHERE `Path` = ?', values %changed, $self->path);
        $self->clear();
        $self->_log(comment => $in{comment}) if $inc_version;
    });
    return;
}

sub delete {
    my ($self, %in) = @_;
    MR::OnlineConf::Admin::Storage->transaction(sub {
        {
            $self->clear();
            local $self->_local_options->{for_update} = 1;
            die sprintf "Node %s not exists\n", $self->path unless $self->exists();
            die sprintf "Version not match: %s != %s\n", $in{version}, $self->version if defined $in{version} && $in{version} != $self->version;
            die "Access denied\n" unless $self->rw;
        }
        MR::OnlineConf::Admin::Storage->do('UPDATE `my_config_tree` SET `Deleted` = true, `Version` = `Version` + 1, MTime = now() WHERE `ID` = ?', $self->id);
        $self->delete_all_access();
        $self->clear();
        {
            local $self->_local_options->{load_deleted} = 1;
            $self->_log(comment => $in{comment});
        }
    });
    return;
}

sub move {
    my ($self, %in) = @_;
    MR::OnlineConf::Admin::Storage->transaction(sub {
        my $comment = sprintf "Moved from %s.", $self->path;
        $comment .= ' ' . $in{comment} if defined $in{comment};
        my $old_path = $self->path;
        {
            my $current = MR::OnlineConf::Admin::Parameter->new($self->path, $self->username);
            local $current->_local_options->{for_update} = 1;
            die sprintf "Node %s not exists\n", $current->path unless $current->exists();
            die sprintf "Version not match: %s != %s\n", $in{version}, $current->version if defined $in{version} && $in{version} != $current->version;
            die "Access denied\n" unless $current->rw;
        }
        die "Node $in{path} already exists\n" if MR::OnlineConf::Admin::Parameter->new($in{path}, $self->username)->exists();
        my $parent_path = $in{path};
        $parent_path =~ s/\/([^\/]+)$//;
        $parent_path ||= '/';
        my $name = $1;
        my $parent = MR::OnlineConf::Admin::Parameter->new($parent_path, $self->username);
        die "Parent node $parent_path not exists\n" unless $parent->exists();
        die "Access denied\n" unless $parent->rw;
        MR::OnlineConf::Admin::Storage->do('UPDATE `my_config_tree` SET `ParentID` = ?, `Name` = ?, `Version` = `Version` + 1, `MTime` = now() WHERE `Path` = ?', $parent->id, $name, $self->path);
        $self->_update_children_path();
        $self->clear();
        $self->_path($in{path});
        $self->_log(comment => $comment);
        if ($in{symlink}) {
            my $symlink = MR::OnlineConf::Admin::Parameter->new(
                path     => $old_path,
                mime     => 'application/x-symlink',
                data     => $self->path,
                username => $self->username,
            );
            my $comment = sprintf "Moved to %s.", $self->path;
            $comment .= ' ' . $in{comment} if defined $in{comment};
            $symlink->create(comment => $comment);
        }
    });
    return;
}

sub set_access {
    my ($self, $group, $rw) = @_;
    die "Access denied\n" unless $self->rw || MR::OnlineConf::Admin::Access->can_edit($self->username);
    MR::OnlineConf::Admin::Access->set($self->id, $group, $rw);
    $self->clear_access();
    return;
}

sub delete_access {
    my ($self, $group) = @_;
    die "Access denied\n" unless $self->rw || MR::OnlineConf::Admin::Access->can_edit($self->username);
    MR::OnlineConf::Admin::Access->delete($self->id, $group);
    $self->clear_access();
    return;
}

sub delete_all_access {
    my ($self, $group) = @_;
    die "Access denied\n" unless $self->rw || MR::OnlineConf::Admin::Access->can_edit($self->username);
    MR::OnlineConf::Admin::Access->delete_all($self->id);
    $self->clear_access();
    return;
}

sub clear {
    my ($self) = @_;
    foreach my $attr ($self->meta->get_all_attributes()) {
        if (my $clearer = $attr->clearer) {
            $self->$clearer();
        }
    }
    return;
}

sub validate {
    my ($self) = @_;
    $self->_validate($self->mime, $self->data);
    die "Only root user can disable notifications\n"
        if $self->_notification_changed && defined $self->notification && $self->notification eq 'none'
        && !MR::OnlineConf::Admin::Group->can_edit($self->username);
    return;
}

sub _build_children {
    my ($self) = @_;
    return [
        map MR::OnlineConf::Admin::Parameter->new(path => $_->{Path}, username => $self->username, _row => $_),
            @{MR::OnlineConf::Admin::Storage->select("
                SELECT $ASTERISK,
                    (SELECT count(*) FROM `my_config_tree` c WHERE c.`ParentID` = t.`ID` AND NOT c.`Deleted`) AS `NumChildren`,
                    (SELECT count(*) <> 0 FROM `my_config_tree_group` g WHERE g.`NodeID` = t.`ID`) AS `AccessModified`,
                    `my_config_tree_access`(t.`ID`, ?) AS `RW`,
                    `my_config_tree_notification`(t.`ID`) AS `Notification`,
                    t.`Notification` IS NOT NULL AS `NotificationModified`
                FROM `my_config_tree` t WHERE `ParentID` = ? AND NOT `Deleted`
                ORDER BY `Name`
            ", $self->username, $self->id)}
    ];
}

sub _build_access {
    my ($self) = @_;
    return MR::OnlineConf::Admin::Access->select_by_node_id($self->id);
}

sub _log {
    my ($self, %in) = @_;
    my $version = MR::OnlineConf::Admin::Version->new(
        node_id  => $self->id,
        path     => $self->path,
        mime     => $self->mime,
        data     => $self->data,
        version  => $self->version,
        mtime    => $self->mtime,
        deleted  => $self->deleted,
        username => $self->username,
        author   => $self->username,
        comment  => $in{comment},
    );
    $version->log(notification => $self->notification);
    return;
}

sub _validate {
    my ($self, $mime, $data) = @_;
    if ($mime eq 'application/x-null') {
        die "Data is not NULL\n" if defined $data;
    } elsif ($mime eq 'application/json') {
        eval { JSON::from_json($data); 1 } or die "Invalid JSON: $@";
    } elsif ($mime eq 'application/x-yaml') {
        eval { YAML::Load($data); 1 } or die "Invalid YAML: $@";
    } elsif ($mime eq 'application/x-symlink') {
        die "Invalid symlink: path not found\n"
            unless MR::OnlineConf::Admin::Parameter->new($data, $self->username)->exists();
    } elsif ($mime eq 'application/x-case') {
        my $cases;
        eval { $cases = JSON::from_json($data); 1 } or die "Invalid JSON: $@";
        foreach my $case (@$cases) {
            $self->_validate($case->{mime}, $case->{value});
        }
    } else {
        return;
    }
}

sub _update_children_path {
    my ($self) = @_;
    $self->clear_children();
    MR::OnlineConf::Admin::Storage->do("UPDATE `my_config_tree` SET Path = NULL WHERE `ParentID` = ?", $self->id);
    foreach my $child (@{$self->children}) {
        $child->_update_children_path();
    }
    return;
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
