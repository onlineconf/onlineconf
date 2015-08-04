package MR::OnlineConf2::Updater::Parameter;

use Mouse;
use YAML;
use JSON;
use File::Spec::Unix;
use Sys::Hostname ();
use Text::Glob;
use MR::OnlineConf2::Updater::Transaction;

has id => (
    is  => 'ro',
    isa => 'Int',
    required => 1,
);

has name => (
    is  => 'rw',
    isa => 'Str',
    required => 1,
);

has path => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
);

has data => (
    is  => 'ro',
    isa => 'Maybe[Str]',
    required => 1,
);

has value => (
    is  => 'ro',
    isa => 'Maybe[Str|HashRef|ArrayRef]',
    lazy    => 1,
    default => sub {
        my ($self) = @_;
        my $type = $self->content_type;

        if ($type eq 'application/json') {
            return JSON::from_json($self->data);
        } elsif ($type eq 'application/x-yaml') {
            return YAML::Load($self->data);
        }

        return $self->data;
    },
);

has is_null => (
    is  => 'ro',
    isa => 'Bool',
    lazy   => 1,
    default => sub { $_[0]->content_type eq 'application/x-null' },
);

has is_case => (
    is => 'ro',
    isa => 'Bool',
    lazy => 1,
    default => sub { $_[0]->content_type eq 'application/x-case' },
);

has children => (
    is  => 'ro',
    isa => 'HashRef',
    default => sub { {} },
);

has version => (
    is  => 'ro',
    isa => 'Int',
    required => 1,
);

has content_type => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
);

sub child {
    my ($self, $name) = @_;
    return $self->children->{$name};
}

sub add_child {
    my ($self, $child) = @_;
    $self->children->{$child->name} = $child;
    return;
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
