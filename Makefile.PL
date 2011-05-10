use ExtUtils::MakeMaker;

WriteMakefile(
    NAME      => 'MR::OnlineConfig',
    MAKEFILE  => 'Makefile',
    PREREQ_PM => {
        Exporter            => 0,
        AnyEvent            => '5.26',
        EV                  => '3.90',
        'JSON::XS'          => 0,
        'Data::Dumper'      => 0,
        'List::Util'        => 0,
        'Digest::MD5'       => 0,
        'POSIX'             => 0,
        'Class::Singleton'  => 0,
        'Getopt::Long'      => 0,
        'IO::Handle'        => 0
    },
    EXE_FILES => [
        'bin/onlineconf_updater',
    ],
);
