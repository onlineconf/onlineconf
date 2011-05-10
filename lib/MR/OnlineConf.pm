package MR::OnlineConf;

use strict;
use MR::OnlineConf::Const qw/:all/;
use MR::OnlineConf::Updater;
use Class::Singleton;
use Exporter;
use Data::Dumper;
use Carp qw/carp/;
use base qw/Class::Singleton Exporter/;

sub PRELOAD()   {'_ALL_'}

sub _new_instance {
    my ($class,%opts) = @_;
    %opts = (
        debug=>1,
        check_interval=>5,
        %opts);
    my $self = {
        cache=>{},
        check_all => 0,
        checks    => {},
        load      => {},
        updater   => MR::OnlineConf::Updater->new({debug=>$opts{debug}}),
        cfg       => \%opts
    };
    return bless $self , $class;
}

sub _say {
    my ($self,$level,@msg) = @_;
    return 1 if $level > $self->{cfg}{debug};
    warn join( ":" , (caller())[0,2]).' '.(join " " , map {ref $_ ? Dumper $_ : $_} @msg);
    return 1;
}

sub get {
    my ($self,$module,$key,$default) = @_;
    $self->_say(-1,"incorrect call. module and  key must be defined\n") 
        and return $default unless $module && $key;
    $self->reload($module);
    $self->_say(-1,"cant find key $key in module $module: use default value\n") 
        and return $default unless exists $self->{cache}{$module} && exists $self->{cache}{$module}{$key};
    return $self->{cache}{$module}{$key};
}

sub preload {
    my ($self) = @_;
    my $preload = $self->PRELOAD();
    unless (ref $preload eq 'ARRAY'){
        $self->_say(-1,"unknown constant $preload") and return undef unless $preload eq '_ALL_';
        return $self->reloadAll();
    }
    foreach my $i (@$preload){
        $self->reload($i);
    }
    return 1;
}

sub _check {
    my ($self,$module) = @_;
    $self->_say(2,"module $module never checked and need to load\n") 
        and return 1 unless exists $self->{checks}{$module};

    $self->_say(2,"skip check for module $module due timelimit\n") 
        and return 0 if $self->{checks}{$module} + $self->{cfg}{check_interval} > time;

    $self->{checks}{$module} = time;

    $self->_say(2,"module $module never loaded and need to load\n") 
        and return 1 unless exists $self->{load}{$module};

    my @stat = $self->{updater}->statFile($module);
    unless (@stat){
        $self->_say(-1,"cant stat module $module\n");
        return undef;
    }
    my $r = ! !($stat[9] > $self->{load}{$module});
    $self->_say(2,"module $module ".($r ? 'changed and need to reload' : 'not changed')."\n");
    return $r;
}

sub _check_all {
    my ($self) = @_;
    my $m = $self->{updater}->localList() || return [];
    return [ grep {$self->_check($_)} keys %$m ];
}

sub _reload {
    my ($self,$module) = @_;
    my $data = $self->{updater}->readFile($module,md5_check=>1);
    unless ($data){
        $self->_say(-1,"cant reload config $module\n");
        $self->{checks}{$module} = time;
        return undef;
    }
    $self->_say(3,"read: " , $data);
    $self->{cache}{$module} = $data->{Data};
    $self->{checks}{$module} = time;
    $self->{load}{$module} = time;
    $self->_say(1,"reload config $module ok\n");
    return 1;
}

sub _reload_all {
    my ($self) = @_;
    my $m = $self->{updater}->localList() || return undef;
    $self->_reload($_) for keys %$m;
}

sub reload {
    my ($self,$module,%opts) = @_;
    %opts = (force=>0,%opts);
    unless ($opts{force}){
        return unless $self->_check($module);
    }
    return $self->_reload($module);
}

sub reloadAll {
    my ($self,%opts) = @_;
    %opts = (force=>0,%opts);
    unless ($opts{force}){
        my $m = $self->_check_all() || return undef;
        $self->_reload($_) for @$m;
    }else{
        $self->_reload_all();
    }
}

1;
