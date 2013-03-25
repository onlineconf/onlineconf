package MR::OnlineConf::Updater::PerlMemory;

use Mouse;
use IO::Interface::Simple;
use List::MoreUtils qw/uniq/;
use Net::IP::CMatch;
use Scalar::Util qw/weaken/;
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

has _dirty_cases => (
    is  => 'ro',
    isa => 'ArrayRef[MR::OnlineConf::Updater::Parameter]',
    default => sub { [] },
);

has _dirty_symlinks => (
    is  => 'ro',
    isa => 'ArrayRef[MR::OnlineConf::Updater::Parameter]',
    default => sub { [] },
);

has _required_by => (
    is  => 'ro',
    isa => 'HashRef[ArrayRef[MR::OnlineConf::Updater::Parameter]]',
    default => sub { {} },
);

sub put {
    my ($self, $param) = @_;
    if (my $node = $self->_index->{$param->id}) {
        if ($node->version < $param->version) {
            $self->_remove_from_requires($node);
            if ($node->path eq $param->path) {
                my $old_version = $node->version;
                $node->update($param);
                $self->log->debug(sprintf "Parameter %s was updated from version %s to %s", $node->path, $old_version, $node->version);
                $self->_add_to_requires($node);
            } else {
                if (my $old_p = $self->get($node->parent_path)) {
                    delete $old_p->children->{$node->name};
                }
                if (my $new_p = $self->get($param->parent_path)) {
                    my $old_version = $node->version;
                    my $old_path = $node->path;
                    $node->update($param);
                    $new_p->add_child($node);
                    $self->log->debug(sprintf "Parameter %s was updated from version %s to %s and moved from %s", $node->path, $old_version, $node->version, $old_path);
                    $self->_add_to_requires($node);
                }
            }
        }
        return 1;
    } elsif ($node = $self->_root) {
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
                $self->_add_to_requires($param);
                return 1;
            } else {
                die sprintf "Failed to put parameter %s: no parent node found", $param->path;
            }
        }
        if ($node->path eq $param->path) {
            if ($node->version < $param->version) {
                $self->_remove_from_requires($node);
                my $old_version = $node->version;
                $node->update($param);
                $self->log->debug(sprintf "Parameter %s was updated from version %s to %s", $node->path, $old_version, $node->version);
                $self->_add_to_requires($node);
                return 1;
            }
        } elsif ($param->path eq $parent->path . '/' . $param->name) {
            if ($param->is_symlink && $node->path eq $param->value) {
                return 0;
            } else {
                $parent->add_child($param);
                $self->_index->{$param->id} = $param;
                $self->_add_to_requires($param);
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

sub delete {
    my ($self, $param) = @_;
    $self->_remove_from_requires($param);
    delete $self->_index->{$param->id};
    my @path = split /\//, $param->path;
    shift @path;
    my $node = $self->_root;
    while (defined(my $name = shift @path)) {
        if (my $child = $node->child($name)) {
            if (@path == 0) {
                $node->delete_child($child);
                return 1;
            } else {
                $node = $child;
            }
        } else {
            return 0;
        }
    }
    return 0;
}

sub get {
    my ($self, $path, $no_resolve) = @_;
    my @path = split /\//, $path;
    shift @path;
    my $node = $self->_root;
    while (defined(my $name = shift @path)) {
        if ($node = $node->child($name)) {
            return $node if $no_resolve && @path == 0;
            my %seen;
            while ($node->is_symlink) {
                die "Recursion in symlink" if $seen{$node->id};
                $seen{$node->id} = 1;
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
    $self->_resolve_cases();
    $self->_resolve_symlinks();
    return;
}

sub _resolve_symlinks {
    my ($self) = @_;
    my $symlinks = $self->_dirty_symlinks;
    foreach my $symlink (sort { length($a->path) <=> length($b->path) } uniq @$symlinks) {
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
        $self->log->debug(sprintf "Symlink resolved: %s -> %s", $symlink->path, $node->path);
        $symlink->symlink_target($node);
        return;
    }
    $self->log->warning(sprintf "Symlink not resolved: %s -> %s", $symlink->path, $symlink->value);
    return;
}

sub _add_to_requires {
    my ($self, $node) = @_;
    my $requires = $node->requires;
    foreach my $reqpath (@$requires) {
        my $list = $self->_required_by->{$reqpath} ||= [];
        @$list = grep $_, uniq @$list, $node;
        weaken($_) foreach @$list;
    }
    $self->_resolve_depends($node);
    push @{$self->_dirty_cases}, $node if $node->is_case;
    push @{$self->_dirty_symlinks}, $node if $node->is_symlink;
    return;
}

sub _remove_from_requires {
    my ($self, $node) = @_;
    my $requires = $node->requires;
    foreach my $reqpath (@$requires) {
        my $list = $self->_required_by->{$reqpath} ||= [];
        @$list = grep { $_ && $_->id != $node->id } uniq @$list;
        weaken($_) foreach @$list;
    }
    $self->_resolve_depends($node);
    return;
}

sub _resolve_depends {
    my ($self, $node) = @_;
    $self->_resolve_symlink_depends($node) if $node->is_symlink;
    return;
}

sub _resolve_symlink_depends {
    my ($self, $node) = @_;
    my $path = $node->path;
    my $target = $self->_required_by;
    foreach my $symlink (map @{$target->{$_}}, grep { index($_, "$path/") == 0 } keys %$target) {
        $symlink->clear_symlink_target();
        push @{$self->_dirty_symlinks}, $symlink if $symlink->is_symlink;
    }
    return;
}

sub _resolve_cases {
    my ($self) = @_;
    my $datacenter = $self->_which_datacenter();
    $self->log->debug(sprintf "Datacenter is %s", $datacenter ? $datacenter->name : 'unknown');
    my $cases = $self->_dirty_cases;
    foreach my $case (sort { length($a->path) <=> length($b->path) } uniq @$cases) {
        $case->resolve_case($datacenter);
        $self->_add_to_requires($case);
        $self->log->debug(sprintf "Case %s resolved", $case->path);
    }
    @$cases = ();
    $self->_resolve_symlinks();
    return;
}

sub _which_datacenter {
    my ($self) = @_;
    my @addrs = map $_->address, grep { $_->is_running && !$_->is_loopback } IO::Interface::Simple->interfaces();
    local $self->{seen} = {};
    if (my $datacenters = $self->get('/onlineconf/datacenter')) {
        foreach my $dc (values %{$datacenters->children}) {
            my @masks = ref $dc->value eq 'ARRAY' ? @{$dc->value} : grep $_, split /(?:,|\s+)/, $dc->value;
            foreach my $addr (@addrs) {
                return $dc if match_ip($addr, @masks);
            }
        }
    }
    return;
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
