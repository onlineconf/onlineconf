package MR::OnlineConf::Updater::Parameter;

use Mouse;
use YAML;
use JSON;
use File::Spec::Unix;
use Sys::Hostname ();
use Text::Glob;

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
    clearer => 'clear_requires',
    default => sub { [ $_[0]->is_symlink ? ($_[0]->value) : (), keys %{$_[0]->_case_requires} ] },
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

has _case_requires => (
    is  => 'rw',
    isa => 'HashRef',
    lazy    => 1,
    default => sub { {} },
    clearer => '_clear_case_requires',
);

sub child {
    my ($self, $name) = @_;
    return $self->children->{$name};
}

sub add_child {
    my ($self, $child) = @_;
    $self->children->{$child->name} = $child;
    $child->_resolve_case($child->data) if $child->content_type eq 'application/x-case';
    return;
}

sub delete_child {
    my ($self, $child) = @_;
    delete $self->children->{$child->name};
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
    $self->_resolve_case($self->data) if $self->content_type eq 'application/x-case';
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
    $self->clear_is_symlink();
    $self->clear_symlink_target();
    $self->clear_requires();
    $self->_clear_case_content_type();
    $self->_clear_case_data();
    $self->_clear_case_requires();
    return;
}

sub resolve_case {
    my ($self, $datacenter) = @_;
    $self->clear();
    $self->_resolve_case($self->data, $datacenter);
    return;
}

sub _resolve_case {
    my ($self, $data, $datacenter) = @_;
    $data = JSON::from_json($data);
    foreach my $case (sort _sort @$data) {
        if (exists $case->{server}) {
            if (Text::Glob::match_glob($case->{server}, Sys::Hostname::hostname())) {
                $self->_apply_case($case);
                return;
            }
        } elsif (exists $case->{datacenter}) {
            if (@_ == 3) {
                if ($datacenter && $case->{datacenter} eq $datacenter->name) {
                    $self->_case_requires->{$datacenter->path} = 1;
                    $self->_apply_case($case);
                    return;
                }
            } else {
                $self->_case_requires->{"/onlineconf/datacenter/$case->{datacenter}"} = 1;
            }
        } else {
            $self->_apply_case($case);
            return;
        }
    }
    $self->_apply_case({ mime => 'application/x-null', value => undef });
    return;
}

sub _apply_case {
    my ($self, $case) = @_;
    if ($case->{mime} eq 'application/x-case') {
        $self->_resolve_case($case->{value});
    } else {
        $self->_case_content_type($case->{mime});
        $self->_case_data($case->{value});
    }
    return;
}

sub _sort {
    if (exists $a->{server} && exists $b->{server}) {
        my $as = $a->{server};
        my $bs = $b->{server};
        my $ar = $as =~ s/(?:[\*\?]+|\{.*?\}|\[.*?\])//g;
        my $br = $bs =~ s/(?:[\*\?]+|\{.*?\}|\[.*?\])//g;
        return length($bs) <=> length($as) || $ar <=> $br || length($b->{server}) <=> length($a->{server});
    } elsif (exists $a->{datacenter} && exists $b->{datacenter}) {
        return $a->{datacenter} cmp $b->{datacenter};
    } else {
        return exists $a->{server} ? -1 : exists $b->{server} ? 1
            : exists $a->{datacenter} ? -1 : exists $b->{datacenter} ? 1
            : 0;
    }
}

sub _update_children_path {
    my ($self) = @_;
    foreach my $child (values %{$self->children}) {
        $child->_path(File::Spec::Unix->catfile($self->path, $child->name));
        $child->_update_children_path();
    }
    return;
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
