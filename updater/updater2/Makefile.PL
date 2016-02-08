use ExtUtils::MakeMaker;

WriteMakefile(
    NAME      => 'MR::OnlineConf::Updater',
    MAKEFILE  => 'Makefile',
    PREREQ_PM => {
        'AnyEvent'        => 0,
        'EV'              => 0,
        'JSON'            => 0,
        'Log::Dispatch'   => 0,
        'Mouse'           => 0,
        'YAML'            => 0,
        'CBOR::XS'        => 0,
        'File::Spec'      => 0,
        'Getopt::Long'    => 0,
        'POSIX'           => 0,
        'Scalar::Util'    => 0,
        'CDB_File'        => 0,
    },
    EXE_FILES => [
        'bin/onlineconf-updater',
    ],
);
