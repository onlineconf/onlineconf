use ExtUtils::MakeMaker;

WriteMakefile(
    NAME      => 'MR::OnlineConf::Updater',
    MAKEFILE  => 'Makefile',
    PREREQ_PM => {
        'AnyEvent'      => 0,
        'EV'            => 0,
        'File::Spec'    => 0,
        'Getopt::Long'  => 0,
        'JSON'          => 0,
        'JSON::XS'      => 0,
        'Log::Dispatch' => 0,
        'Mouse'         => 0,
        'MR::DBI'       => 0,
        'POSIX'         => 0,
        'Scalar::Util'  => 0,
        'Sys::Hostname' => 0,
        'Text::Glob'    => 0,
        'YAML'          => 0,
    },
    EXE_FILES => [
        'bin/onlineconf-updater',
    ],
);
