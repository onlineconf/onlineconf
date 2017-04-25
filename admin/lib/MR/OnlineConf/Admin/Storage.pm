package MR::OnlineConf::Admin::Storage;

use Mouse;

extends 'MR::DBI::NoOnlineConf';

around BUILDARGS => sub {
    my $orig = shift;
    my $class = shift;
    my $args = @_ == 1 ? $_[0] : { @_ };
    $args->{database} ||= $args->{base};
    $args->{utf8} ||= 'utf8mb4';
    return $class->$orig($args);
};

sub singleton {
    return 1;
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
