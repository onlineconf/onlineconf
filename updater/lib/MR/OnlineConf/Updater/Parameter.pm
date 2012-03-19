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
    trigger  => sub { $_[0]->clear_value(); $_[0]->clear_is_symlink(); $_[0]->symlink_target(undef) },
);

has value => (
    is  => 'ro',
    isa => 'Maybe[Str|HashRef|ArrayRef]',
    lazy    => 1,
    clearer => 'clear_value',
    default => sub {
        my ($self) = @_;
        my $type = $self->content_type;
        if ($type eq 'application/json') {
            return JSON::from_json($self->data);
        } elsif ($type eq 'application/x-yaml') {
            return YAML::Load($self->data);
        } else {
            return $self->data;
        }
    },
);

has children => (
    is  => 'ro',
    isa => 'HashRef',
    default => sub { {} },
);

has is_symlink => (
    is  => 'ro',
    isa => 'Bool',
    lazy   => 1,
    default => sub { $_[0]->content_type eq 'application/x-symlink' },
    clearer => 'clear_is_symlink',
);

has symlink_target => (
    is  => 'rw',
    isa => 'Maybe[MR::OnlineConf::Updater::Parameter]',
);

sub BUILD {
    my ($self) = @_;
    $self->_resolve_case($self->data) if $self->content_type eq 'application/x-case';
    return;
}

sub child {
    my ($self, $name) = @_;
    return $self->children->{$name};
}

sub add_child {
    my ($self, $child) = @_;
    $self->children->{$child->name} = $child;
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

sub _resolve_case {
    my ($self, $data) = @_;
    $data = JSON::from_json($data);
    foreach my $case (sort _sort @$data) {
        if (Text::Glob::match_glob($case->{server}, Sys::Hostname::hostname())) {
            if ($case->{mime} eq 'application/x-case') {
                $self->_resolve_case($case->{value});
            } else {
                $self->_content_type($case->{mime});
                $self->_data($case->{value});
            }
            return;
        }
    }
    $self->_content_type('application/x-null');
    $self->_data(undef);
    return;
}

sub _sort {
    my $as = $a->{server};
    my $bs = $b->{server};
    my $ar = $as =~ s/(?:[\*\?]+|\{.*?\}|\[.*?\])//g;
    my $br = $bs =~ s/(?:[\*\?]+|\{.*?\}|\[.*?\])//g;
    return length($bs) <=> length($as) || $ar <=> $br || length($b->{server}) <=> length($a->{server});
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
