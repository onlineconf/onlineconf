package MR::OnlineConf::Admin::Monitoring;

use Mouse;

has host => (
    is  => 'ro',
    isa => 'Str',
    required => 1,
);

has mtime => (
    is  => 'ro',
    isa => 'Str',
    lazy    => 1,
    default => sub { die sprintf "Activity not exists: %s\n", $_[0]->host unless $_[0]->_row; $_[0]->_row->{Time} },
);

has online => (
    is  => 'ro',
    isa => 'Str',
    lazy    => 1,
    default => sub { die sprintf "Activity not exists: %s\n", $_[0]->host unless $_[0]->_row; $_[0]->_row->{Online} },
);

has package => (
    is  => 'ro',
    isa => 'Maybe[Str]',
    lazy    => 1,
    default => sub { die sprintf "Activity not exists: %s\n", $_[0]->host unless $_[0]->_row; $_[0]->_row->{Package} },
);

has mtime_alert => (
    is  => 'ro',
    isa => 'Bool',
    lazy    => 1,
    default => sub { die sprintf "Activity not exists: %s\n", $_[0]->host unless $_[0]->_row; $_[0]->_row->{TimeAlert} },
);

has online_alert => (
    is  => 'ro',
    isa => 'Bool',
    lazy    => 1,
    default => sub { die sprintf "Activity not exists: %s\n", $_[0]->host unless $_[0]->_row; $_[0]->_row->{OnlineAlert} },
);

has _row => (
    is  => 'ro',
    isa => 'HashRef',
    required => 1,
);

around BUILDARGS => sub {
    my $orig = shift;
    my $class = shift;
    if (@_ == 1 && !ref($_[0])) {
        return $class->$orig(host => $_[0]);
    } else {
        return $class->$orig(@_);
    }
};

my %SORT = (host => 'Host', mtime => 'Time', online => 'Online', package => 'Package');

sub list {
    my ($class, %in) = @_;
    my $sort = $SORT{$in{sort} || 'host'} || 'Host';
    my $mtime = MR::OnlineConf::Admin::Storage->select('SELECT MAX(`MTime`) AS `MTime` FROM `my_config_tree_log`')->[0]->{MTime};
    return [ map MR::OnlineConf::Admin::Monitoring->new(host => $_->{Host}, _row => $_),
        @{MR::OnlineConf::Admin::Storage->select("
            SELECT *, `Time` <> ? AND ? < now() - INTERVAL 30 MINUTE AS `TimeAlert`, `Online` < now() - INTERVAL 30 MINUTE AS `OnlineAlert`
            FROM `my_config_activity` ORDER BY `$sort`
        ", $mtime, $mtime)} ];
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
