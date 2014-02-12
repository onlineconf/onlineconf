package MR::OnlineConf::Updater::Parameter;

use Mouse;
use YAML;
use JSON;
use File::Spec::Unix;
use Sys::Hostname ();
use Text::Glob;
use MR::OnlineConf::Updater::Transaction;

has id => (
    is  => 'ro',
    isa => 'Int',
    required => 1,
);

has name => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
    writer   => '_name',
);

has path => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
    writer   => '_path',
);

has version => (
    is  => 'ro',
    isa => 'Int',
    required => 1,
    writer   => '_version',
);

has content_type => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
    writer   => '_content_type',
);

has data => (
    is  => 'ro',
    isa => 'Maybe[Str]',
    required => 1,
    writer   => '_data',
    trigger  => sub { $_[0]->clear() },
);

has value => (
    is  => 'ro',
    isa => 'Maybe[Str|HashRef|ArrayRef]',
    lazy    => 1,
    writer  => 'set_value',
    clearer => 'clear_value',
    default => sub {
        my ($self) = @_;
        my $data;
        my $type = $self->content_type;
        if ($type eq 'application/x-case') {
            $type = $self->_case_content_type;
            $data = $self->_case_data;
        } else {
            $data = $self->data;
        }
        if ($type eq 'application/json') {
            return JSON::from_json($data);
        } elsif ($type eq 'application/x-yaml') {
            return YAML::Load($data);
        } else {
            return $data;
        }
    },
);

has children => (
    is  => 'ro',
    isa => 'HashRef',
    default => sub { {} },
);

has is_case => (
    is  => 'ro',
    isa => 'Bool',
    lazy   => 1,
    default => sub { $_[0]->content_type eq 'application/x-case' },
    clearer => 'clear_is_case',
);

has is_null => (
    is  => 'ro',
    isa => 'Bool',
    lazy   => 1,
    default => sub { $_[0]->content_type eq 'application/x-null' || $_[0]->is_case && $_[0]->_case_content_type && $_[0]->_case_content_type eq 'application/x-null' },
    clearer => 'clear_is_null',
);

has is_template => (
    is  => 'ro',
    isa => 'Bool',
    lazy   => 1,
    default => sub { $_[0]->content_type eq 'application/x-template' || $_[0]->is_case && $_[0]->_case_content_type && $_[0]->_case_content_type eq 'application/x-template' },
    clearer => 'clear_is_template',
);

has is_symlink => (
    is  => 'ro',
    isa => 'Bool',
    lazy   => 1,
    default => sub { $_[0]->content_type eq 'application/x-symlink' || $_[0]->is_case && $_[0]->_case_content_type && $_[0]->_case_content_type eq 'application/x-symlink' },
    clearer => 'clear_is_symlink',
);

has symlink_target => (
    is  => 'rw',
    isa => 'MR::OnlineConf::Updater::Parameter',
    weak_ref => 1,
    clearer  => 'clear_symlink_target',
);

has requires => (
    is  => 'ro',
    isa => 'ArrayRef[Str]',
    lazy    => 1,
    clearer => '_clear_requires',
    default => sub { [ $_[0]->is_symlink ? ($_[0]->value) : $_[0]->is_template ? ( $_[0]->value =~ /\$\{(\/.*?)\}/g ) : () ] },
);

has dirty => (
    is  => 'rw',
    isa => 'Bool',
    default => 0,
    trigger => sub { MR::OnlineConf::Updater::Transaction->add_to_dirty($_[0]) if $_[1] },
);

has dirty_children => (
    is  => 'ro',
    isa => 'Bool',
    default => 0,
    writer  => '_dirty_children',
    trigger => sub { MR::OnlineConf::Updater::Transaction->add_to_dirty($_[0]) if $_[1] },
);

has _case_content_type => (
    is  => 'rw',
    isa => 'Str',
    clearer => '_clear_case_content_type',
);

has _case_data => (
    is  => 'rw',
    isa => 'Maybe[Str]',
    clearer => '_clear_case_data',
);

has requires_datacenter => (
    is  => 'ro',
    isa => 'Bool',
    lazy    => 1,
    default => sub { $_[0]->_has_case_of_type('datacenter') },
    clearer => '_clear_requires_datacenter',
);

has requires_group => (
    is  => 'ro',
    isa => 'Bool',
    lazy    => 1,
    default => sub { $_[0]->_has_case_of_type('group') },
    clearer => '_clear_requires_group',
);

has deleted => (
    is  => 'rw',
    isa => 'Bool',
    default => 0,
);

sub child {
    my ($self, $name) = @_;
    return $self->children->{$name};
}

sub add_child {
    my ($self, $child) = @_;
    $self->children->{$child->name} = $child;
    $self->_dirty_children(1);
    $child->dirty(1);
    return;
}

sub delete_child {
    my ($self, $child) = @_;
    delete $self->children->{$child->name};
    $self->_dirty_children(1);
    $child->dirty(1);
    $child->deleted(1);
    return;
}

sub update {
    my ($self, $new) = @_;
    $self->_data($new->data);
    $self->_content_type($new->content_type);
    $self->_version($new->version);
    if ($new->path ne $self->path) {
        $self->_path($new->path);
        $self->_name($new->name);
        $self->_update_children_path();
    }
    $self->dirty(1);
    return;
}

sub real_node {
    my ($self) = @_;
    my %seen;
    my $node = $self;
    while ($node->is_symlink) {
        return if $seen{$node->id};
        $seen{$node->id} = 1;
        $node = $node->symlink_target;
        return unless $node;
        return if $node->deleted;
    }
    return $node;
}

sub parent_path {
    my ($self) = @_;
    my $path = $self->path;
    $path =~ s/\/[^\/]+$//;
    $path = '/' if $path eq '';
    return $path;
}

sub clear {
    my ($self) = @_;
    $self->clear_value();
    $self->clear_is_case();
    $self->clear_is_null();
    $self->clear_is_template();
    $self->clear_is_symlink();
    $self->clear_symlink_target();
    $self->_clear_requires();
    $self->_clear_requires_datacenter();
    $self->_clear_requires_group();
    $self->_clear_case_content_type();
    $self->_clear_case_data();
    return;
}

sub is_dirty {
    my ($self) = @_;
    return $self->dirty || $self->dirty_children;
}

sub clear_dirty {
    my ($self) = @_;
    $self->dirty(0);
    $self->_dirty_children(0);
    return;
}

sub resolve_case {
    my ($self, $datacenter, $groups) = @_;
    $self->clear();
    $self->_resolve_case($self->data, $datacenter, $groups);
    return;
}

sub _resolve_case {
    my ($self, $data, $datacenter, $groups) = @_;
    $data = JSON::from_json($data);
    my %groups = map { $_ => 1 } @$groups;
    foreach my $case ($self->_sort_case($data, $groups)) {
        if (exists $case->{server}) {
            if (Text::Glob::match_glob($case->{server}, Sys::Hostname::hostname())) {
                $self->_apply_case($case, $datacenter, $groups);
                return;
            }
        } elsif (exists $case->{group}) {
            if ($groups{$case->{group}}) {
                $self->_apply_case($case, $datacenter, $groups);
                return;
            }
        } elsif (exists $case->{datacenter}) {
            if ($datacenter && $case->{datacenter} eq $datacenter->name) {
                $self->_apply_case($case, $datacenter, $groups);
                return;
            }
        } else {
            $self->_apply_case($case, $datacenter, $groups);
            return;
        }
    }
    $self->_apply_case({ mime => 'application/x-null', value => undef });
    return;
}

sub _sort_case {
    my ($self, $data, $groups) = @_;
    my %group_idx = map { $groups->[$_] => $_ } (0 .. $#$groups);
    return sort {
        if (exists $a->{server} && exists $b->{server}) {
            my $as = $a->{server};
            my $bs = $b->{server};
            my $ar = $as =~ s/(?:[\*\?]+|\{.*?\}|\[.*?\])//g;
            my $br = $bs =~ s/(?:[\*\?]+|\{.*?\}|\[.*?\])//g;
            return length($bs) <=> length($as) || $ar <=> $br || length($b->{server}) <=> length($a->{server});
        } elsif (exists $a->{group} && exists $b->{group}) {
            return exists $group_idx{$a->{group}}
                ? exists $group_idx{$b->{group}} ? $group_idx{$a->{group}} <=> $group_idx{$b->{group}} : -1
                : exists $group_idx{$b->{group}} ? 1 : $a->{group} cmp $b->{group};
        } elsif (exists $a->{datacenter} && exists $b->{datacenter}) {
            return $a->{datacenter} cmp $b->{datacenter};
        } else {
            return exists $a->{server} ? -1 : exists $b->{server} ? 1
                : exists $a->{group} ? -1 : exists $b->{group} ? 1
                : exists $a->{datacenter} ? -1 : exists $b->{datacenter} ? 1
                : 0;
        }
    } @$data;
}

sub _apply_case {
    my ($self, $case, $datacenter, $groups) = @_;
    if ($case->{mime} eq 'application/x-case') {
        $self->_resolve_case($case->{value}, $datacenter, $groups);
    } else {
        $self->_case_content_type($case->{mime});
        $self->_case_data($case->{value});
    }
    return;
}

sub _update_children_path {
    my ($self) = @_;
    foreach my $child (values %{$self->children}) {
        $child->_path(File::Spec::Unix->catfile($self->path, $child->name));
        $child->_update_children_path();
        $child->dirty(1);
    }
    return;
}

sub _has_case_of_type {
    my ($self, $type) = @_;
    return 0 unless $self->is_case;
    return ref($self)->_data_has_case_of_type($self->data, $type);
}

sub _data_has_case_of_type {
    my ($class, $data, $type) = @_;
    $data = JSON::from_json($data);
    foreach my $case (@$data) {
        return 1 if exists $case->{$type}
            || $case->{mime} eq 'application/x-case' && $class->_data_has_case_of_type($case->{value}, $type);
    }
    return 0;
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
