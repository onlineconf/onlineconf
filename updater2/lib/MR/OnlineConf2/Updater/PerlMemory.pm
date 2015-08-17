package MR::OnlineConf2::Updater::PerlMemory;

use Mouse;

has log => (
    is  => 'ro',
    isa => 'Log::Dispatch',
    required => 1,
);

has _root => (
    is  => 'rw',
    isa => 'MR::OnlineConf2::Updater::Parameter',
);

sub put {
    my ($self, $param) = @_;

    if (my $node = $self->_root) {
        my @path = grep length($_), split /\//, $param->path;

        while (defined(my $name = shift @path)) {
            if (@path == 0) {
                $param->name($name);
                $node->add_child($param);
                return 1;
            }

            unless ($node = $node->child($name)) {
                die sprintf "Failed to put parameter %s: no parent node found", $param->path;
            }
        }

        return 0;
    } elsif ($param->path eq '/') {
        $self->_root($param);
        return 1;
    } else {
        die sprintf "Failed to put parameter %s: no root node found", $param->path;
    }
}

sub get {
    my ($self, $path) = @_;
    my $node = $self->_root;
    my @path = split /\//, $path;

    shift @path;

    while (defined(my $name = shift @path)) {
        if ($node = $node->child($name)) {
            return $node if @path == 0;
        } else {
            return;
        }
    }

    return $node;
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
