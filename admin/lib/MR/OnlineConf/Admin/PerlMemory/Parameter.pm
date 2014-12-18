package MR::OnlineConf::Admin::PerlMemory::Parameter;

use Mouse;

# External modules
use JSON::XS;
use Text::Glob;

my $clear;
my %types = (
    'text/plain' => 'is_plain',
    'application/json' => 'is_json',
    'application/x-yaml' => 'is_yaml',
    'application/x-list' => 'is_list',
    'application/x-case' => 'is_case',
    'application/x-null' => 'is_null',
    'application/x-server' => 'is_server', # Список пар ip:port
    'application/x-symlink' => 'is_symlink',
    'application/x-template' => 'is_template', #imgsmail.${hostname}
);

while (my ($type, $attribute) = each %types) {
    $clear .= '$self->' . "clear_$attribute();\n";

    __PACKAGE__->meta->add_attribute($attribute,
        is  => 'rw',
        isa => 'Bool',
        lazy => 1,
        default => sub { $_[0]->ContentType eq $type },
        clearer => "clear_$attribute",
    );
}

__PACKAGE__->meta->add_method(
    clear_types => eval "
        sub {
            my (\$self) = \@_;
            $clear
        }
    "
);

has ID => (
    is  => 'ro',
    isa => 'Int',
    required => 1,
);

has Name => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
);

has Path => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
);

has MTime => (
    is => 'ro',
    isa => 'Str',
    writer => '_MTime',
    required => 1
);

has Value => (
    is  => 'ro',
    isa => 'Maybe[Str]',
    writer => '_Value',
    required => 1,
);

has Deleted => (
    is => 'ro',
    isa => 'Bool',
    required => 1,
);

has Version => (
    is  => 'ro',
    isa => 'Int',
    writer => '_Version',
    required => 1,
);

has ContentType => (
    is  => 'ro',
    isa => 'Str',
    writer => '_ContentType',
    required => 1,
);

has JSONParser => (
    is => 'ro',
    isa => 'JSON::XS',
    lazy => 1,
    default => sub {
        return JSON::XS->new->utf8(0);
    }
);

has value => (
    is  => 'ro',
    isa => 'Maybe[Str|HashRef|ArrayRef]',
    lazy => 1,
    writer => 'set_value',
    clearer => 'clear_value',
    default => sub {
        my ($self) = @_;
        my $data = $self->Value;

        if ($self->is_case) {
            $data = $self->case_data;
        }

        return unless defined $data;

        if ($self->is_json) {
            return $data;
        } elsif ($self->is_yaml) {
            return $data;
        } elsif ($self->is_template) {
            my $host = $self->host;
            my $addr = $self->addr;

            $data =~ s#\$\{(.*?)\}#
                my $str = $1;
                my $org = $1;

                if ($str eq 'hostname') {
                    unless ($str = $host) {
                        $str = '${' . $org . '}';
                    }
                } elsif ($str eq 'hostname -s' || $str eq 'short_hostname') {
                    unless (($str) = $host =~ /^(\w+)\./) {
                        $str = '${' . $org . '}';
                    }
                } elsif ($str eq 'hostname -i' || $str eq 'ip') {
                    unless ($str = join ' ', @$addr) {
                        $str = '${' . $org . '}';
                    }
                }

                $str;
            #eg;
        }

        $data =~ s/\n/\\n/g;
        $data =~ s/\r/\\r/g;

        return $data;
    },
);

has host => (
    is => 'rw',
    isa => 'Str',
    lazy => 1,
    default => sub { '' }
);

has addr => (
    is => 'rw',
    isa => 'ArrayRef',
    lazy => 1,
    default => sub { [] },
);

has groups => (
    is => 'rw',
    isa => 'ArrayRef',
);

has children => (
    is  => 'ro',
    isa => 'HashRef',
    default => sub { {} },
);

has case_data => (
    is  => 'rw',
    isa => 'Maybe[Str]',
    clearer => 'clear_case_data',
);

has datacenter => (
    is => 'rw',
    isa => 'MR::OnlineConf::Admin::PerlMemory::Parameter',
);

has groups_hash => (
    is => 'ro',
    isa => 'HashRef',
    lazy => 1,
    clearer => 'clear_groups_hash',
    default => sub {
        return {
            map { $_ => 1 } @{
                $_[0]->groups
            }
        };
    }
);

has unpacked_cases => (
    is => 'ro',
    isa => 'ArrayRef',
    lazy => 1,
    clearer => 'clear_unpacked_cases',
    default => sub {
        my ($self) = @_;
        return $self->_unpack_case($self->Value);
    }
);

has symlink_target => (
    is  => 'rw',
    isa => 'MR::OnlineConf::Admin::PerlMemory::Parameter',
    clearer => 'clear_symlink_target',
    weak_ref => 1,
);

sub add_child {
    my ($self, $child) = @_;
    $self->children->{$child->Name} = $child;
    return;
}

sub delete_child {
    my ($self, $child) = @_;
    delete $self->children->{$child->Name};
    return;
}

sub clear {
    my ($self) = @_;

    $self->clear_types();
    $self->clear_value();
    $self->clear_case_data();
    $self->clear_groups_hash();
    $self->clear_unpacked_cases();
    $self->clear_symlink_target();
}

sub clear_case_before_resolve {
    my ($self) = @_;

    $self->clear_types();
    $self->clear_value();
    $self->clear_case_data();
    $self->clear_groups_hash();
    $self->clear_symlink_target();
}

sub clear_symlink_before_resolve {
    my ($self) = @_;

    $self->clear_symlink_target();
}

sub clear_template_before_resolve {
    my ($self) = @_;

    $self->clear_value();
}

sub resolve_case {
    my ($self) = @_;
    $self->_resolve_case($self->unpacked_cases);
    return;
}

sub _resolve_case {
    my ($self, $data) = @_;
    my $host = $self->host;
    my $groups = $self->groups;
    my $datacenter = $self->datacenter;
    my $groups_hash = $self->groups_hash;

    foreach my $case (@$data) {
        my $bingo;

        if (exists $case->{server}) {
            if (hostname_match_glob($case->{server}, $host)) {
                $bingo++;
            }
        } elsif (exists $case->{group}) {
            if ($groups_hash->{$case->{group}}) {
                $bingo++;
            }
        } elsif (exists $case->{datacenter}) {
            if ($datacenter && $case->{datacenter} eq $datacenter->Name) {
                $bingo++;
            }
        } else {
            $bingo++;
        }

        if ($bingo) {
            my $mime = $case->{mime};

            if ($mime eq 'application/x-case') {
                $self->_resolve_case($case->{value});
            } else {
                if (my $attribute = $types{$mime}) {
                    $self->$attribute(1);
                }

                $self->case_data($case->{value});
            }

            return;
        }
    }

    $self->is_null(1);
    $self->case_data(undef);

    return;
}

sub _sort_case {
    my ($self, $data) = @_;
    my $groups = $self->groups;
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

sub _unpack_case {
    my ($self, $data) = @_;

    $data = $self->JSONParser->decode($data);

    foreach my $case (@$data) {
        if ($case->{mime} eq 'application/x-case') {
            $case->{value} = $self->_unpack_case(
                $case->{value}
            );
        }
    }

    return [ $self->_sort_case($data) ];
}

my %glob_to_regex_cache;

sub hostname_match_glob {
    my $glob = shift;
    my $re;

    unless ($re = $glob_to_regex_cache{$glob}) {
        $glob_to_regex_cache{$glob} = $re = Text::Glob::glob_to_regex($glob);
    }

    return $_[0] =~ $re;
}

no Mouse;

__PACKAGE__->meta->make_immutable();

1;
