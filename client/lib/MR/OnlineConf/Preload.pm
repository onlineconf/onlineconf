package MR::OnlineConf::Preload;

use strict;
use warnings;

sub PRELOAD () {'_ALL_'}

sub preload {
    my ($self) = @_;
    my $preload = $self->PRELOAD();

    if ($self->{config}{enable_cdb_client}) {
        foreach my $i (keys %{$self->{cache}}) {
            $self->reload($i);
        }
    } elsif (ref $preload eq 'ARRAY'){
        foreach my $i (@$preload){
            $self->reload($i);
        }
    } else {
        $self->_say(-1,"unknown constant $preload") and return undef unless $preload eq '_ALL_';
        my $m = $self->_updater_localList() || return undef;
        $self->_reload($_) for grep {$self->_check($_)} keys %$m;
    }

    return 1;
}

sub _updater_localList {
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
        next if $1 eq 'ALL';
        $files{$1} = $self->LOCAL_CFG_PATH().$f;
    }
    closedir DIR;
    return \%files;
}

1;
