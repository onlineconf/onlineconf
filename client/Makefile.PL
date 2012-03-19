use ExtUtils::MakeMaker;

WriteMakefile(
    NAME      => 'MR::OnlineConf',
    MAKEFILE  => 'Makefile',
    PREREQ_PM => {
        'Carp'              => 0,
        'Class::Singleton'  => 0,
        'Data::Dumper'      => 0,
        'Exporter'          => 0,
        'JSON::XS'          => 0,
        'List::Util'        => 0,
        'Digest::MD5'       => 0,
        'POSIX'             => 0,
        'YAML'              => 0
    },
);
