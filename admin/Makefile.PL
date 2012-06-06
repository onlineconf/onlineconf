use ExtUtils::MakeMaker;

WriteMakefile(
    NAME      => 'MR::OnlineConf::Admin',
    MAKEFILE  => 'Makefile',
    PREREQ_PM => {
        'File::Spec'                          => 0,
        'Getopt::Long'                        => 0,
        'JSON'                                => 0,
        'JSON::XS'                            => 0,
        'List::MoreUtils'                     => 0,
        'Log::Dispatch'                       => 0,
        'Mojolicious'                         => 0,
        'Mojolicious::Plugin::MysqlBasicAuth' => 0,
        'Mouse'                               => 0,
        'MR::ChangeBot::Database'             => 0,
        'MR::DBI'                             => 0,
        'POSIX'                               => 0,
        'Pod::Usage'                          => 0,
        'Starman'                             => 0,
        'Text::Glob'                          => 0,
        'YAML'                                => 0,
    },
    EXE_FILES => [
        'bin/onlineconf-admin',
        'bin/onlineconf-import',
        'bin/onlineconf-migration',
        'bin/onlineconf-selftest',
    ],
);
