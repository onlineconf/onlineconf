package MR::OnlineConf::Updater::PerlMemory;

use Mouse;
use Sys::Hostname ();
use Text::Glob;

has log => (
    is  => 'ro',
    isa => 'Log::Dispatch',
    required => 1,
);

has _root => (
    is  => 'rw',
    isa => 'MR::OnlineConf::Updater::Parameter',
);

has _index => (
    is  => 'rw',
    isa => 'HashRef',
    default => sub { {} },
);

has _symlinks => (
    is  => 'ro',
    isa => 'ArrayRef[MR::OnlineConf::Updater::Parameter]',
    default => sub { [] },
);

sub put {
    my ($self, $param) = @_;
    if (my $node = $self->_index->{$param->id}) {
        if ($node->version < $param->version) {
            if ($node->path eq $param->path) {
                my $old_version = $node->version;
                $node->update($param);
                $self->log->debug(sprintf "Parameter %s was updated from version %s to %s", $node->path, $old_version, $node->version);
                push @{$self->_symlinks}, $node if $node->is_symlink;
                return 1;
            } else {
                my $old_pp = $node->parent_path;
                my $new_pp = $param->parent_path;
                if ($old_pp ne $new_pp) {
                    if (my $old_p = $self->get($old_pp)) {
                        delete $old_p->children->{$node->name};
                    }
                    if (my $new_p = $self->get($new_pp)) {
                        my $old_version = $node->version;
                        my $old_path = $node->path;
                        $node->update($param);
                        $new_p->add_child($node);
                        $self->log->debug(sprintf "Parameter %s was updated from version %s to %s and moved from %s", $node->path, $old_version, $node->version, $old_path);
                        push @{$self->_symlinks}, $node if $node->is_symlink;
                        return 1;
                    }
                }
            }
        }
    }
    if (my $node = $self->_root) {
        my @path = split /\//, $param->path;
        shift @path;
        my $parent;
        while (defined(my $name = shift @path)) {
            if (my $next = $node->child($name)) {
                $parent = $node;
                $node = $next;
            } elsif (@path == 0) {
                $node->add_child($param);
                $self->_index->{$param->id} = $param;
                push @{$self->_symlinks}, $param if $param->is_symlink;
                return 1;
            } else {
                die sprintf "Failed to put parameter %s: no parent node found", $param->path;
            }
        }
        if ($node->path eq $param->path) {
            if ($node->version < $param->version) {
                my $old_version = $node->version;
                $node->update($param);
                $self->log->debug(sprintf "Parameter %s was updated from version %s to %s", $node->path, $old_version, $node->version);
                push @{$self->_symlinks}, $node if $node->is_symlink;
                return 1;
            }
        } elsif ($param->path eq $parent->path . '/' . $param->name) {
            if ($param->is_symlink && $node->path eq $param->value) {
                return 0;
            } else {
                $parent->add_child($param);
                $self->_index->{$param->id} = $param;
                push @{$self->_symlinks}, $param if $param->is_symlink;
                return 1;
            }
        } else {
            die sprintf "Failed to put parameter %s: no valid path found", $param->path;
        }
        return 0;
    } elsif ($param->path eq '/') {
        $self->_root($param);
        $self->_index->{$param->id} = $param;
        return 1;
    } else {
        die sprintf "Failed to put parameter %s: no root node found", $param->path;
    }
}

sub get {
    my ($self, $path, $no_follow) = @_;
    my @path = split /\//, $path;
    shift @path;
    my $node = $self->_root;
    while (defined(my $name = shift @path)) {
        if ($node = $node->child($name)) {
            return $node if $no_follow && @path == 0;
            while ($node->is_symlink) {
                if (!$node->symlink_target && exists $self->{seen}) {
                    $self->_resolve_symlink($node);
                }
                $node = $node->symlink_target;
                return unless $node;
            }
        } else {
            return;
        }
    }
    return $node;
}

sub finalize {
    my ($self) = @_;
    $self->_resolve_symlinks();
    return;
}

sub _resolve_symlinks {
    my ($self) = @_;
    my $symlinks = $self->_symlinks;
    foreach my $symlink (sort { length($a->path) <=> length($b->path) } @$symlinks) {
        local $self->{seen} = {};
        $self->_resolve_symlink($symlink);
    }
    @$symlinks = ();
    return;
}

sub _resolve_symlink {
    my ($self, $symlink) = @_;
    if ($self->{seen}->{$symlink->id}) {
        $self->log->warning("Recursive symlink " . $symlink->path);
        return;
    }
    $self->{seen}->{$symlink->id} = 1;
    if (my $node = $self->get($symlink->value, 1)) {
        $self->{seen}->{$node->id} = 1;
        $self->log->debug(sprintf "Symlink resolved: %s -> %s", $symlink->path, $symlink->value);
        $symlink->symlink_target($node);
        return;
    }
    $self->log->warning(sprintf "Symlink not resolved: %s -> %s", $symlink->path, $symlink->value);
    return;
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
