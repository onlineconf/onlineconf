package MR::OnlineConf::Const;

use strict;
use Exporter;
use vars qw/@EXPORT_OK %EXPORT_TAGS @CONST %CONST %ERRORS @ERRORS/;
use base qw/Exporter/;

$EXPORT_TAGS{all}      = \@EXPORT_OK;
$EXPORT_TAGS{const}    = \@CONST;
$EXPORT_TAGS{errors}   = \@ERRORS;

sub constantiate($$;@) {
    my ($const, @arrs) = @_;
    my $pkg = (caller)[0];
    my @a;    
    eval qq{ sub $pkg\::$_ () { $const->{$_} }; push \@a, q{$_}; 'TRUE'.$_; } or die $@
        for keys %$const;
    push @$_, @a for @arrs;
}

BEGIN {
    %CONST = (
        MY_CONFIG_JSON_FLAG             => 0x00000001,
        MY_CONFIG_DELETED_FLAG          => 0x00000002,
        MY_CONFIG_SYMLINK_FLAG          => 0x00000004,
        MY_CONFIG_CURRENT_VER           => -1,
        
        MY_CONFIG_OVERLOAD_MODULE_NAME  => '\'@OVERLOAD\'',
        MY_CONFIG_SELFTEST_MODULE_NAME  => '\'@SELFTEST\'',
        MY_CONFIG_SELFTEST_TIME_KEY     => '\'update-time\'',
        MY_CONFIG_SELFTEST_DELAY_KEY    => '\'delay\'',
        MY_CONFIG_SELFTEST_ENABLED_KEY  => '\'enabled\'',
        
        MY_CONFIG_GROUP_ROOT_NAME       => '\'root\'',
    );

    constantiate \%CONST , \@CONST , \@EXPORT_OK;

    %ERRORS = (
        E_PREV_VERSION_MISMATCH => -1,
        E_NO_CHANGES            => -2,
        E_UNDEFINED             => -3,
        E_KEY_ALREADY_EXISTS    => -4,
        E_BAD_DATA              => -5
    );

    constantiate \%ERRORS , \@ERRORS , \@EXPORT_OK;
}

1;


