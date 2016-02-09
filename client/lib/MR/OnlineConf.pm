package MR::OnlineConf;

use strict;

use base qw/Class::Singleton Exporter MR::OnlineConf::Preload/;

use YAML;
use CDB_File;
use CBOR::XS;
use Data::Dumper;
use Sys::Hostname ();
use POSIX qw/strftime/;
use Carp qw/carp confess/;

use MR::OnlineConf::Const qw/:all/;
use MR::OnlineConf::Util qw/from_json to_json/;

sub ERRORLOG()  {'/var/tmp/onlineconf_error.txt'}

my $DEFAULT_CONFIG = {
    database=> {
        host      => 'unknown', 
        user      => 'unknown',
        password  => 'unknown',
        base      => 'unknown',
        timeout   => 2,                                                                                                           
    },
    data_dir => '/usr/local/etc/onlineconf/',
    logfile  => '/var/log/onlineconf_updater.log', 
    pidfile  => '/var/run/onlineconf_updater.pid',
    update_interval => 60,
    test_interval   => 60,
    debug => 0,
    enable_cdb_client => 0,
};

{
    my $config;
    my $cache = {};
    my $checks = {};
    my $load = {};

    sub _read_config {
        my $file = $ENV{PERL_ONLINECONF_CONFIG} || '/usr/local/etc/onlineconf.yaml';
        if (-r $file) {
            $config = YAML::LoadFile($file) or
                confess "cant load config file at $file";
        } else {
            warn "WARNING: onlineconf can't load config file from `$file`. default config will be used.\n";
            $config = $DEFAULT_CONFIG;
        }
        return;
    }

    sub _new_instance {
        my ($class, %opts) = @_;
        %opts = (
            debug=>0,
            check_interval=>5,
            reload => 1,
            %opts);
        _read_config() unless $config;
        my $self = {
            cache_cdb => {},
            cache     => $cache,
            checks    => $checks,
            load      => $load,
            cfg       => \%opts,
            logstatus => undef,
            hostname  => Sys::Hostname::hostname(),
            config    => $config,
            test_expires => 0,
        };
        return bless $self , $class;
    }
}

sub _say {
    my ($self,$level,@msg) = @_;
    return 1 if $level > $self->{cfg}{debug};
    warn "[".strftime('%Y/%d/%m %H:%M:%S' , localtime)."] ".join( ":" , (caller())[0,2]).' '.(join " " , map {ref $_ ? Dumper $_ : $_} @msg);
    return 1;
}

sub _logerr {
    my ($self,$level,@msg) = @_;
    $self->_say($level,@msg);

    open(F,'>'.ERRORLOG());
    print F "OnlineConf SelfTest: ".(join(" ", map {ref $_ ? Dumper $_ : $_} @msg))." [source: $0]\n";
    close(F);
    $self->{logstatus} = 1;
}

sub _reseterr {
    my ($self) = @_;
    $self->_say(3,"no reason to clear errorlog") && return if defined $self->{logstatus} && !$self->{logstatus};
    truncate(ERRORLOG(),0) or $self->_say(-1,"can't truncate file: $!");
    $self->{logstatus} = 0;
    return 1;
}

sub get {
    my ($self,$module,$key,$default) = @_;

    $self->_test();

    if ($module && $module =~ /^\//) {
        $default = $key;
        $key = $module;
        $module = 'TREE';
    }

    $self->_say(-1,"incorrect call. module and  key must be defined\n") 
        and return $default unless $module && $key;

    $self->reload($module) if $self->{cfg}{reload};

    if ($self->{config}{enable_cdb_client}) {
        if (exists $self->{cache_cdb}{$module}{$key}) {
            return $self->{cache_cdb}{$module}{$key};
        }
    }

    $self->{cfg}{debug} < 2 || $self->_say(2,"cant find key $key in module $module: use default value\n") 
        and return $default unless exists $self->{cache}{$module} && exists $self->{cache}{$module}{$key};

    unless ($self->{config}{enable_cdb_client}) {
        return $self->{cache}{$module}{$key};
    }

    return $self->{cache_cdb}{$module}{$key} = $self->_get_cdb_value($module, $key);
}

sub getModule {
    my ($self, $module) = @_;
    $self->_say(-1,"incorrect call. module must be defined\n") and return unless $module;
    $self->reload($module) if $self->{cfg}{reload};

    if ($self->{config}{enable_cdb_client}) {
        return {
            map {
                $_ => $self->_get_cdb_value($module, $_)
            } keys %{$self->{cache}{$module}}
        }
    }

    return { %{$self->{cache}{$module}} };
}

sub reload {
    my ($self,$module,%opts) = @_;
    return unless $opts{force} || $self->_check($module);
    return $self->_reload($module);
}

sub _check {
    my ($self,$module) = @_;
    $self->_say(2,"module $module never checked and need to load\n") 
        and return 1 unless exists $self->{checks}{$module};

    $self->{cfg}{debug} < 2 || $self->_say(2,"skip check for module $module due timelimit\n") 
        and return 0 if $self->{checks}{$module} + $self->{cfg}{check_interval} > time;

    $self->{checks}{$module} = time;

    $self->_say(2,"module $module never loaded and need to load\n") 
        and return 1 unless exists $self->{load}{$module};

    my @stat = $self->_updater_statFile($module);
    unless (@stat){
        $self->_say(-1,"cant stat module $module\n");
        return undef;
    }
    my $r = ! !($stat[9] > $self->{load}{$module});
    $self->_say(2,"module $module ".($r ? 'changed and need to reload' : 'not changed')."\n");
    return $r;
}

sub _reload {
    my ($self,$module) = @_;

    if ($self->{config}{enable_cdb_client}) {
        if (-e (my $file = $self->LOCAL_CFG_PATH().$module.'.cdb')) {
            $self->{cache_cdb}{$module} = {};

            if ($self->{cache}{$module}) {
                untie $self->{cache}{$module};
            }

            delete $self->{cache}{$module};

            tie %{
                $self->{cache}{$module}
            }, 'CDB_File', $file or die "tie failed: $!\n";;
        }
    } else {
        my $data = $self->_updater_readFile($module,md5_check=>1);
        unless ($data){
            $self->_say(-1,"cant reload config $module\n");
            $self->{checks}{$module} = time;
            return undef;
        }
        $self->_say(3,"read: " , $data);
        $self->{cache}{$module} = $data->{Data};
    }

    $self->{load}{$module} = time;
    $self->{checks}{$module} = time;
    
    $self->_say(1,"reload config $module ok\n");

    return 1;
}

sub _test {
    my ($self) = @_;

    return if time() < $self->{test_expires};

    $self->{test_expires} = time() + ($self->{config}->{test_interval} || 60);

    $self->reload(MY_CONFIG_SELFTEST_MODULE_NAME);
    $self->_logerr(1,"cant read selftest module\n") 
        and return unless exists $self->{cache}{MY_CONFIG_SELFTEST_MODULE_NAME()} && exists $self->{cache}{MY_CONFIG_SELFTEST_MODULE_NAME()}{MY_CONFIG_SELFTEST_TIME_KEY()};

    $self->_reseterr() and return unless $self->{cache}{MY_CONFIG_SELFTEST_MODULE_NAME()}{MY_CONFIG_SELFTEST_ENABLED_KEY()};

    my $last_update = $self->{cache}{MY_CONFIG_SELFTEST_MODULE_NAME()}{MY_CONFIG_SELFTEST_TIME_KEY()};
    my $delay       = $self->{cache}{MY_CONFIG_SELFTEST_MODULE_NAME()}{MY_CONFIG_SELFTEST_DELAY_KEY()} || 0;
    my $diff        = time - $last_update;

    if ($diff > $delay) {
        $self->_logerr(1,"$self->{hostname}: selftest module hasnt updates for $diff secs (limit: $delay, last update: $last_update)");
    } else {
        $self->_reseterr();
    }
}

sub _get_cdb_value {
    my ($self, $module, $key) = @_;
    my $val = $self->{cache}{$module}{$key};
    my $typ = substr $val, 0, 1, '';

    if ($typ eq 's') {
        utf8::decode($val);
    }

    if ($typ eq 'j') {
        $val = eval {
            from_json($val);
        };

        if ($@) {
            $self->_say(-1,"cant parse json variable $key => $val\n: $@");
            return undef;
        }
    }

    if ($typ eq 'c') {
        $val = eval {
            CBOR::XS::decode_cbor($val);
        };

        if ($@) {
            $self->_say(-1,"cant parse cbor variable $key => $val\n: $@");
            return undef;
        }
    }

    return $val;
}

sub LOCAL_CFG_PATH () { $_[0]->{config}{data_dir} }

sub _updater_statFile {
    my ($self,$mod) = @_;
    my $name = $self->LOCAL_CFG_PATH().$mod.'.conf';
    return stat $name;
}

sub _updater_readFile {
    my ($self,$name,%opts) = @_;
    %opts = (md5_check=>1,%opts);
    my $file = $self->LOCAL_CFG_PATH().$name.'.conf';
    unless (open F , $file){
        $self->_say(-1,"cant open file $file\n");
        return undef;
    }
    my @s = <F>;
    close F;
    my $data = $self->_updater_restore(\@s);
    unless ($data){
        $self->_say(-1,"cant restore config from file $file\n");
        return undef;
    }
    unless ($name eq $data->{Name}){
        $self->_say(-1,"module mismatch: file $file contain config from module $data->{Name}\n");
        return undef;
    }
    return $data;
}

sub _updater_restore {
    my ($self,$s) = @_;
    my $data = {Data=>{}};
    foreach (@$s){
        $self->_say(4,"parse line ".$_."\n");
        if (/^\s*#\!\s*(\S+)\s+(.+)$/){
            $self->_say(4,"found special var $1 => $2\n");
            $data->{$1} = $2;
        }elsif(/^\s*#@\s*(\S+)\s+(\S+?):(\S+)$/){
            $self->_say(4,"found symlink $1 -> $2:$3\n");
            push @{$data->{SymLink}}, { LinkName => $1, TargetModule => $2, TargetName => $3 };
        }elsif(/^\s*#/){
            $self->_say(4,"found comment\n");
            next;
        }elsif(/\s*(\S+)\s+(.+)$/){
            $self->_say(4,"found var $1 => $2\n");
            my ($k,$v) = ($1,$2);
            $v=~s/\\n/\n/g;
            $v=~s/\\r/\r/g;
            if ($k=~/^(.+?)\:JSON$/){
                $self->_say(4,"var $k is JSON");
                $k = $1;
                my $p = eval {from_json($v)};
                $self->_say(-1,"cant parse json variable $k => $v\n: $@")
                    and return undef if $@;
                $v = $p;
            } else {
                utf8::decode($v);
            }
            $data->{Data}{$k} = $v;
        }
    }
    unless ($s->[-1] eq '#EOF'){
        $self->_say(-1,'cant find EOF marker') and return undef;
    }
    $self->_say(-1,"cant find Version or/and Name variable") and return undef 
        unless $data->{Version} && $data->{Name};
    $self->_say(3,"read config " , $data);
    return $data;    
}

1;
