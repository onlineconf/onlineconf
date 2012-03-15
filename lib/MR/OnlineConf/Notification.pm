package MR::OnlineConf::Notification;

use strict;
use warnings;
use utf8;
use Encode;
use MR::ChangeBot::Notification;

my @message;

sub _d ($) { Encode::decode('CP1251', $_[0]) }

sub init {
    my ($class, %options) = @_;
    $options{database} = $options{base};
    MR::ChangeBot::Database->new(\%options);
    return;
}

sub begin {
    @message = ();
    MR::ChangeBot::Database->begin();
    return;
}

sub commit {
    MR::ChangeBot::Notification->new(origin => 'onlineconf', message => join("\n", @message))->create();
    @message = ();
    MR::ChangeBot::Database->commit();
    return;
}

sub rollback {
    @message = ();
    MR::ChangeBot::Database->rollback();
    return;
}

sub on_add {
    my ($class, $module, $key, $value, $comment, $commiter) = @_;
    my $message = sprintf '%s добавил параметр %s:%s.', _d $commiter, _d $module, _d $key;
    $message .= ' ' . _d $comment if $comment;
    push @message, $message;
    return;
}

sub on_update {
    my ($class, $module, $key, $value, $comment, $commiter) = @_;
    my $message = sprintf '%s изменил параметр %s:%s.', _d $commiter, _d $module, _d $key;
    $message .= ' ' . _d $comment if $comment;
    push @message, $message;
    return;
}

sub on_delete {
    my ($class, $module, $key, $commiter) = @_;
    my $message = sprintf '%s удалил параметр %s:%s.', _d $commiter, _d $module, _d $key;
    push @message, $message;
    return;
}

1;
