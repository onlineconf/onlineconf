package MR::OnlineConf::Storage;

# Only for compatibility with onlineconf v1. Only for OK::Config, cgi-bin/my/config and dwh-panic-drop

use Mouse;
use MR::ChangeBot::Database;
use MR::OnlineConf::Admin::Storage;
use MR::OnlineConf::Admin::Parameter;

my $USERNAME = 'onlineconf';

has database => (
    is  => 'ro',
    isa => 'HashRef',
    required => 1,
);

has notification_database => (
    is  => 'ro',
    isa => 'HashRef',
    required => 1,
);

sub BUILD {
    my ($self) = @_;
    MR::OnlineConf::Admin::Storage->new($self->database);
    MR::ChangeBot::Database->new($self->notification_database);
    return;
}

sub modules {
    return [];
}

sub canModifyGroup {
}

sub logAll {
    return [];
}

sub log {
    return [];
}

sub version {
    return 0;
}

sub getModuleIDByName {
    my ($self, $module) = @_;
    return $module;
}

sub update {
    my ($self, $module, $data, %in) = @_;
    MR::OnlineConf::Admin::Storage->transaction(sub {
        foreach my $p (values %$data) {
            $self->_create_parent($module, $p->{Key}, $in{commiter});
            my $path = $p->{Key};
            $path =~ s/\./\//g;
            $path = "/onlineconf/module/$module/$path";
            my $param = MR::OnlineConf::Admin::Parameter->new(
                path     => $path,
                mime     => 'text/plain',
                data     => $p->{Value},
                username => $in{commiter},
            );
            if ($param->exists()) {
                $param->update(comment => $in{log});
            } else {
                $param->create(comment => $in{log});
            }
        }
    });
    return;
}

sub delete {
    my ($self, $module, $data, %in) = @_;
    my @keys = ref $data eq 'HASH' ? keys %$data : @$data;
    MR::OnlineConf::Admin::Storage->transaction(sub {
        foreach my $key (@keys) {
            my $path = $key;
            $path =~ s/\./\//g;
            $path = "/onlineconf/module/$module/$path";
            my $param = MR::OnlineConf::Admin::Parameter->new($path, $in{commiter});
            $param->delete(comment => $in{log}) if $param->exists();
        }
    });
    return;
}

sub getMulti {
    my ($self, $module, $keys) = @_;
    my %keys = map { $_ => 1 } @$keys;
    my $data = $self->getAll($module);
    return { map { $_ => $data->{$_} } grep $keys{$_}, keys %$data };
}

sub getAll {
    my ($self, $module) = @_;
    my $param = MR::OnlineConf::Admin::Parameter->new("/onlineconf/module/$module", $USERNAME);
    return { map %{$self->_walk_children($_)}, @{$param->children} };
}

sub _walk_children {
    my ($self, $param) = @_;
    my %data;
    my $prefix = $param->name;
    $data{$param->name} = { Value => $param->data } unless $param->mime eq 'application/x-null';
    foreach my $child (@{$param->children}) {
        my $data = $self->_walk_children($child);
        $data{"$prefix.$_"} = $data->{$_} foreach keys %$data;
    }
    return \%data;
}

sub _create_parent {
    my ($self, $module, $key, $user) = @_;
    my @part = split /\./, $key;
    my $path = "/onlineconf/module/$module";
    while (my $part = shift @part) {
        $path .= "/$part";
        my $param = MR::OnlineConf::Admin::Parameter->new($path, $user);
        $param->create() unless $param->exists();
    }
    return;
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
