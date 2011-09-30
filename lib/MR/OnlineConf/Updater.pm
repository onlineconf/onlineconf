package MR::OnlineConf::Updater;

use strict;
use MR::OnlineConf::Storage;
use MR::OnlineConf::Util qw/from_json to_json/;
use MR::OnlineConf::Const qw/:const/;
use Data::Dumper;
use List::Util qw/max/;
use Digest::MD5 qw/md5_hex/;
use EV;
use AnyEvent;
use POSIX qw/strftime/;
use Scalar::Util qw/weaken/;
use IO::Handle;

sub LOCAL_CFG_PATH()        {$_[0]->{config}{data_dir} }

sub new {
    my ($class,$opts) = @_;
    my $self = bless {
        local=>{},
        remote=>{},
        config=>$opts
    } , $class;
    $self->{storage} = new MR::OnlineConf::Storage($opts);
    return $self; 
}

sub _say {
    my ($self,$level,@msg) = @_;
    return 1 if $level > $self->{config}{debug};
    warn "[".strftime("%Y/%d/%m %H:%M:%S" , localtime)."] ".join( ":" , (caller())[0,2])." ".(join " " , map {ref $_ ? Dumper $_ : $_} @msg);
    return 1;
}

sub reloadLocal {
    my ($self) = @_;
    my $files = $self->localList() || return undef;
    while (my ($n,$f) = each %$files){
        my $data = $self->readFile($n,md5_check=>0);
        $self->{local}{$n} = {Version=>$data->{Version},Name=>$data->{Name}};
    }
    return 1;
}

sub readFile {
    my ($self,$name,%opts) = @_;
    %opts = (md5_check=>1,%opts);
    my $file = $self->LOCAL_CFG_PATH().$name.'.conf';
    unless (open F , $file){
        $self->_say(-1,"cant open file $file\n");
        return undef;
    }
    my @s = <F>;
    close F;
    my $data = $self->restore(\@s);
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

sub statFile {
    my ($self,$mod) = @_;
    my $name = $self->LOCAL_CFG_PATH().$mod.'.conf';
    return stat $name;
}

sub localList {
    my ($self) = @_;
    unless (-d $self->LOCAL_CFG_PATH()){
        $self->_say(-1,"try to create catalog ".$self->LOCAL_CFG_PATH()."\n");
        unless (mkdir $self->LOCAL_CFG_PATH()){
            $self->_say(-1,"cant create config catalog ".$self->LOCAL_CFG_PATH().": $!\n");
            return undef;
        }
    }
    unless (opendir (DIR, $self->LOCAL_CFG_PATH())){
        $self->_say(-1,"cant open local catalog `".$self->LOCAL_CFG_PATH()." for reading.\n");
        return undef;
    } 
    my %files;
    while (my $f = readdir DIR){
        next unless $f=~/^(.+?)\.conf$/;
        $files{$1} = $self->LOCAL_CFG_PATH().$f;
    }
    closedir DIR;
    return \%files;
}

sub needUpdate {
    my ($self) = @_;
    my $modules = $self->{storage}->modules();
    unless ($modules){
        $self->_say(-1,"cant load modules list from storage.\n");
        return ();
    }
    $self->{remote} = { map {$_->{Name} => $_} @$modules};

    my %to_update;
    my $overload_updated;
    foreach my $m (@$modules){
        unless ($self->{local}{$m->{Name}} && $self->{local}{$m->{Name}}{Version} == $m->{Version}){
            if($m->{Version} > 0) {
                $to_update{$m->{Name}} = 1;
                $overload_updated = 1 if $m->{Name} eq MY_CONFIG_OVERLOAD_MODULE_NAME;
            }
        }
    }
    if($overload_updated) {
        %to_update = map {$_->{Name} => 1} @$modules;
    }

    my %mod = map {$_->{Name} => 1} @$modules;
    foreach my $m (keys %{$self->{local}}){
        unless (exists $mod{$m}){
            $self->_say(-1,"module $m exists on local machine but no presents in storage\n");
            delete $self->{local}{$m};
        }
    }
    return keys %to_update;
}

sub update {
    my ($self,@mod) = @_;
    return 1 unless @mod;

    my $overload;
    if(grep {$_ eq MY_CONFIG_OVERLOAD_MODULE_NAME} @mod) {
        $overload = $self->_get_overload({map {$_ => 1} @mod});
    }

    my $data = $self->{storage}->getAllFromModules([ map {$self->{remote}{$_}{ID}} @mod ]);
    my %names = map {$_->{ID}=>$_} values %{$self->{remote}};
    my %ids   = map {$_->{Name}=>$_} values  %{$self->{remote}};
    unless ($data && ref $data eq 'HASH'){
        $self->_say(-1,"cant load config from modules [ @mod ]\n");
        return undef;
    }
    foreach my $k (keys %$data){
        my $name = $names{$k}->{Name};
        unless ($name){
            $self->_say(-1,"unknown module id $k\n");
            next;
        }

        my $ver = max map {$_->{Version}} values %{$data->{$k}};
        
        if($overload) {
            foreach my $overload_module (keys %$overload) {
                foreach my $overload_key (keys %{$overload->{$overload_module}}) {
                    my $mid = $ids{$overload_module}->{ID};
                    $data->{$mid}->{$overload_key}->{Key} = $overload_key;
                    $data->{$mid}->{$overload_key}->{Value} = $overload->{$overload_module}->{$overload_key};
                }
            }
        }

        my $str = $self->dump(
            {
                Name=>$name,
                Version=>$ver,
                Data=>{map {$_->{Key} => $_->{Value}} values %{$data->{$k}}}
            });
        #my $md5 = md5_hex $str;
        unless (open F , '>'.$self->LOCAL_CFG_PATH().$name.'.conf_tmp'){
            $self->_say(-1,"cant open ".$self->LOCAL_CFG_PATH().$name.".conf_tmp for writing: $!\n");
            next;
        }
=h        
        unless (open MD5 , '>'.$self->LOCAL_CFG_PATH().$name.'.md5'){
           $self->_say(-1, "cant open ".$self->LOCAL_CFG_PATH().$name.".md5 for writing: $!\n");
            close F;
            next;
        }
=cut        
        print F $str;
#        print MD5 $md5;
        close F;
        unless (rename $self->LOCAL_CFG_PATH().$name.'.conf_tmp' , $self->LOCAL_CFG_PATH().$name.'.conf'){
            $self->_say(-1,"cant rename ".$self->LOCAL_CFG_PATH().$name.".conf_tmp to to .conf: $!");
            next;
        } 
#       close MD5;
        $self->{local}{$name}{Version} = $ver;
    }
}

sub _get_overload {
    my ($self,$known_modules) = @_;

    my $hostname = `hostname`;
    chomp $hostname;

    unless($hostname) {
        $self->_say(-1,"cant get hostname: $!");
        return;
    }
    $self->_say(4,"hostname = $hostname");

    my $data = $self->{storage}->getAll(MY_CONFIG_OVERLOAD_MODULE_ID,json2perl=>0);
    unless($data && keys %$data) {
        $self->_say(-1,"nothing for overload");
        return;
    }

    my $overload;
    for my $k (keys %$data) {
        my ($host,$module,$key) = split /:/, $k, 3;
        unless($host && $module && $key) {
            $self->_say(-1,"illegal overload key format $k");
            next;
        }
        $self->_say(4,"fetch for overload host = $host, module = $module, key = $key");
        next unless $host eq $hostname;

        unless($known_modules->{$module}) {
            $self->_say(-1,"unknown module for overload $module");
            next;
        }
        $overload->{$module}->{$key} = $data->{$k}->{Value};
        $self->_say(-1,"overload $module.$key with value = ".$overload->{$module}->{$key});
    }
    return $overload;
}

sub _start_logrotate {
    my ($self) = @_;
    if ($self->{config}{logfile}) {
    	unless (open(STDOUT, ">>$self->{config}{logfile}")){
            $self->_say(-1,"cant open logfile $self->{config}{logfile}: $!\n");
            return;
        }
        STDOUT->autoflush(1);
    	unless (open(STDERR, '>>&STDOUT')){
            $self->_say(-1,"cant dup stdout to stderr: $!\n");
            return ;
        }
	    STDERR->autoflush(1);
    }

    weaken($self);
    my $guard = AnyEvent->signal(signal=>'HUP' , cb=>sub {
        $self->_say(-1,"accept HUP signal\n");
        if (defined $self->{config}{logfile}) {
            close STDOUT;
            close STDERR;
            unless (open(STDOUT, ">>$self->{config}{logfile}")) {
                $self->_say(-1,"can't write to logfile: $self->{config}{logfile}: $!\n");
                return ;
            }
            STDOUT->autoflush(1);
            unless (open(STDERR, '>>&STDOUT')){
                $self->_say(-1,"cant dup stdout to stderr: $!\n");
                return ;
            }
            STDERR->autoflush(1);
        }
    });

    $self->{sighup_watcher} = $guard;
}

sub _stop_logrotate {
    my ($self) = @_;
    delete $self->{sighup_watcher};
}

sub _update_periodically {
    my ($self) = @_;
    delete $self->{update_timer};
    $self->{iteration} ||=0;
    $self->{iteration}++;
    $self->{first} = 1 unless defined $self->{first};
    if( ($self->{config}{reload_local} && !($self->{iteration} % $self->{config}{reload_local})) 
            || $self->{first}){
            $self->{iteration} = 0;
            $self->_say(1,"reload local files\n");
            $self->reloadLocal();
            $self->{first} = 0;
    }
    $self->_say(1,"check updates\n");
    $self->{storage}->open();

    my $hostname = `hostname`;
    chomp $hostname;
    $self->{storage}->updateActivity($hostname);
    
    my @d = $self->needUpdate();
    if (@d){
        $self->_say(1,"update modules [ @d ]\n");
        $self->update(@d);
    }else{
        $self->_say(1,"no updates available\n");
    }
    $self->{update_timer} = AnyEvent->timer(
        after=>$self->{config}{update_interval},
        cb => sub {$self->_update_periodically()}
    );
    $self->{storage}->close(); # dont leave connect to db
}

sub updatePeriodically {
    my ($self,$logrotate) = @_;
    my $i = 0;
    my $first = 1;
    $self->_start_logrotate() if $logrotate;
    foreach my $sig ( qw( TERM INT ) ) {
        $self->{$sig.'_watch'} = AnyEvent->signal(
            signal => $sig,
            cb     => sub {
                $self->_say(-1,"caught a SIG$sig - shutting down\n");
                delete $self->{update_timer};
                $self->_stop_logrotate() if $logrotate;
                EV::unloop;
                return;
            },
        );
    }
    $self->_update_periodically();
    EV::loop;
}

sub _dump_dd {
    my ($self,$data) = @_;
    {
        local $Data::Dumper::Indent = 1;
        local $Data::Dumper::Terse = 1;
        my $d = Dumper($data);
        return $d;
    }
}

sub _restore_dd {
    my ($self,$s) = @_;
    my $str = join "" , @$s;
    my $v = eval $str;
    if ($@){
        $self->_say(-1,"cant restore data: $@\n");
        return undef;
    }
    return $v;
}

sub _dump_plain {
    my ($self,$data) = @_;
    my $s = "# This file is autogenerated by $0 at ".strftime("%Y/%d/%m %H:%M:%S" , localtime)."\n";
    $s.="#! Name ".$data->{Name}."\n";
    $s.="#! Version ".$data->{Version}."\n\n";
    foreach my $i (sort keys %{$data->{Data}}){
        my $v = $data->{Data}{$i};
        if (ref $v){
            $i.=':JSON';
            $v = eval { to_json($v) };
        }
        $v=~s/\n/\\n/g;
        $v=~s/\r/\\r/g;
        $s.=$i.' '.$v."\n";
    }
    $s.="#EOF";
    return $s;
}

sub _restore_plain {
    my ($self,$s) = @_;
    my $data = {Data=>{}};
    foreach (@$s){
        $self->_say(4,"parse line ".$_."\n");
        if (/^\s*#\!\s*(\S+)\s+(.+)$/){
            $self->_say(4,"found special var $1 => $2\n");
            $data->{$1} = $2;
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

*dump    = \&_dump_plain;
*restore = \&_restore_plain;

1;

