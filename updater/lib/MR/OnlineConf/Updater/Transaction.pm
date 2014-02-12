package MR::OnlineConf::Updater::Transaction;

use Mouse;
use Scalar::Util qw/weaken/;

has _dirty => (
    is  => 'ro',
    isa => 'HashRef',
    default => sub { {} },
);

my $current;

sub current {
    return $current;
}

sub BUILD {
    my ($self) = @_;
    die "Current transaction already exists" if $current;
    $current = $self;
    weaken($current);
    return;
}

sub DEMOLISH {
    my ($self) = @_;
    $_->clear_dirty() foreach values %{$self->_dirty};
    return;
}

sub add_to_dirty {
    my ($class, $node) = @_;
    my $self = $class->current();
    $self->_dirty->{$node->path} = $node;
    return;
}

sub dirty {
    my $self = shift->current();
    return values %{$self->_dirty};
}

sub dirty_cases {
    my $self = shift->current();
    return grep $_->is_case, values %{$self->_dirty};
}

sub dirty_symlinks {
    my $self = shift->current();
    return grep $_->is_symlink, values %{$self->_dirty};
}

sub dirty_templates {
    my $self = shift->current();
    return grep $_->is_template, values %{$self->_dirty};
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
