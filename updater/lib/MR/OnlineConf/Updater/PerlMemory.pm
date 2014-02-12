package MR::OnlineConf::Updater::PerlMemory;

use Mouse;
use IO::Interface::Simple;
use List::MoreUtils qw/uniq/;
use Net::IP::CMatch;
use Scalar::Util qw/weaken/;
use MR::OnlineConf::Updater::Transaction;
use MR::OnlineConf::Updater::Util;

my $transaction;

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

has _required_by => (
    is  => 'ro',
    isa => 'HashRef[HashRef[MR::OnlineConf::Updater::Parameter]]',
    default => sub { {} },
);

has _require_datacenter => (
    is  => 'ro',
    isa => 'HashRef',
    default => sub { {} },
);

has _require_group => (
    is  => 'ro',
    isa => 'HashRef',
    default => sub { {} },
);

sub put {
    my ($self, $param) = @_;
    $transaction ||= MR::OnlineConf::Updater::Transaction->new();
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
                    $old_p->delete_child($node);
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
    $transaction ||= MR::OnlineConf::Updater::Transaction->new();
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
                return if $node->deleted;
            }
        } else {
            return;
        }
    }
    return $node;
}

sub finalize {
    my ($self) = @_;
    $transaction ||= MR::OnlineConf::Updater::Transaction->new();
    $self->_resolve_cases();
    $self->_resolve_symlinks();
    $self->_resolve_templates();
    undef $transaction;
    return;
}

sub _has_dirty_deep {
    my ($self, $path) = @_;
    my $node = $self->get($path);
    return 0 unless $node;
    return 1 if $node->is_dirty();
    foreach my $child (values %{$node->children}) {
        return 1 if $self->_has_dirty_deep($child->path);
    }
    return 0;
}

sub _resolve_templates {
    my ($self) = @_;
    $self->_mark_dirty_templates();
    foreach my $node (MR::OnlineConf::Updater::Transaction->dirty_templates()) {
        local $self->{seen} = {};
        $self->_resolve_template($node);
    }
    return;
}

sub _mark_dirty_templates {
    my ($self) = @_;
    foreach my $node (MR::OnlineConf::Updater::Transaction->dirty()) {
        my $path = $node->path;
        my $target = $self->_required_by;
        foreach my $node (exists $target->{$path} ? (grep defined, values %{$target->{$path}}) : ()) {
            $node->dirty(1);
        }
    }
    return;
}

sub _resolve_template {
    my ($self, $template) = @_;
    $template->clear_value();
    if ($self->{seen}->{$template->id}) {
        $self->log->warning("Recursive template variable " . $template->path);
        return;
    }
    $self->{seen}->{$template->id} = 1;
    my $value = $template->value;
    $value =~ s#\$\{(.*?)\}#
        my $var = $1;
        my $replace = '';
        if ($var =~ /^\//) {
            if (my $node = $self->get($var)) {
                if (!ref $node->value) {
                    $self->log->debug(sprintf "Template variable resolved: %s: %s", $template->path, $node->path);
                    $replace = $node->value;
                    push @{$template->requires}, $node->path if $node->path ne $var && !grep { $_ eq $node->path } @{$template->requires};
                } else {
                    $self->log->warning(sprintf "Value of template variable is not text: %s: %s", $template->path, $var);
                }
            } else {
                $self->log->warning(sprintf "Template variable not resolved: %s: %s", $template->path, $var);
            }
        } elsif (defined (my $r = expand_template_macro($var))) {
            $replace = $r;
        } else {
            $self->log->warning(sprintf "Unknown template macro: %s: %s", $template->path, $var);
        }
        $replace;
    #eg;
    $self->log->debug(sprintf "Template expanded: %s: %s -> %s", $template->path, $template->value, $value);
    $template->set_value($value);
    $self->_add_to_requires($template);
    return;
}

sub _resolve_symlinks {
    my ($self) = @_;
    $self->_mark_dirty_symlinks();
    foreach my $symlink (sort { length($a->path) <=> length($b->path) } MR::OnlineConf::Updater::Transaction->dirty_symlinks()) {
        local $self->{seen} = {};
        $self->_resolve_symlink($symlink);
    }
    return;
}

sub _mark_dirty_symlinks {
    my ($self) = @_;
    foreach my $node (MR::OnlineConf::Updater::Transaction->dirty()) {
        my $path = $node->path;
        my $target = $self->_required_by;
        foreach my $node (grep defined, map values %{$target->{$_}}, grep { index($_, "$path/") == 0 } keys %$target) {
            $node->dirty(1);
        }
    }
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
        my $req = $self->_required_by->{$reqpath} ||= {};
        $req->{$node->id} = $node;
        weaken($req->{$node->id});
    }
    $self->_require_datacenter->{$node->path} = $node if $node->requires_datacenter;
    $self->_require_group->{$node->path} = $node if $node->requires_group;
    return;
}

sub _remove_from_requires {
    my ($self, $node) = @_;
    my $requires = $node->requires;
    foreach my $reqpath (@$requires) {
        my $req = $self->_required_by->{$reqpath} ||= {};
        delete $req->{$node->id};
    }
    delete $self->_require_datacenter->{$node->path};
    delete $self->_require_group->{$node->path};
    return;
}

sub _resolve_cases {
    my ($self) = @_;
    $self->_mark_dirty_cases();
    my $datacenter = $self->_which_datacenter();
    $self->log->debug(sprintf "Datacenter is %s", $datacenter ? $datacenter->name : 'unknown');
    my $groups = $self->_which_groups();
    $self->log->debug(sprintf "Groups are [%s]", join ', ', @$groups);
    foreach my $case (sort { length($a->path) <=> length($b->path) } MR::OnlineConf::Updater::Transaction->dirty_cases()) {
        $case->resolve_case($datacenter, $groups);
        $self->_add_to_requires($case);
        $self->log->debug(sprintf "Case %s resolved", $case->path);
    }
    return;
}

sub _mark_dirty_cases {
    my ($self) = @_;
    local $self->{seen} = {};
    if ($self->_has_dirty_deep('/onlineconf/datacenter')) {
        foreach my $node (values %{$self->_require_datacenter}) {
            $node->dirty(1);
        }
    }
    if ($self->_has_dirty_deep('/onlineconf/group')) {
        foreach my $node (values %{$self->_require_group}) {
            $node->dirty(1);
        }
    }
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

sub _which_groups {
    my ($self) = @_;
    my @groups;
    local $self->{seen} = {};
    if (my $groups = $self->get('/onlineconf/group')) {
        my @all_groups = grep { $_ ne 'priority' } sort keys %{$groups->children};
        my @ordered_groups;
        if (my $priority = $self->get('/onlineconf/group/priority')) {
            push @ordered_groups, map { $_ eq '*' ? @all_groups : $_ }
                split /\s*,\s*/, $priority->value if $priority->value;
        }
        push @ordered_groups, @all_groups;
        @ordered_groups = grep exists($groups->children->{$_}), uniq @ordered_groups;
        my @list;
        foreach my $name (@ordered_groups) {
            push @list, $name => $groups->children->{$name};
        }
        while (my ($name, $node) = splice(@list, 0, 2)) {
            $node = $self->get($node->path) or next;
            my $glob = $node->value;
            push @groups, $name if defined $glob && hostname_match_glob($glob);
            foreach my $subname (sort keys %{$node->children}) {
                push @list, $name => $node->children->{$subname};
            }
        }
    }
    return [ uniq @groups ];
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
