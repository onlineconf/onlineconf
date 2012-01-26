package MR::OnlineConf::Notification;

use strict;
use warnings;
use utf8;
use Encode;
use MR::ChangeBot::Notification;

sub _d ($) { Encode::decode('CP1251', $_[0]) }

sub init {
    my ($class, %options) = @_;
    $options{database} = $options{base};
    MR::ChangeBot::Database->new(\%options);
    return;
}

sub begin {
    MR::ChangeBot::Database->begin();
}

sub commit {
    MR::ChangeBot::Database->commit();
}

sub rollback {
    MR::ChangeBot::Database->rollback();
}

sub on_add {
    my ($class, $module, $key, $value, $comment) = @_;
    my $message = sprintf 'Добавлен параметр %s:%s со значением "%s".', _d $module, _d $key, _d $value;
    $message .= ' ' . _d $comment if $comment;
    MR::ChangeBot::Notification->new(origin => 'onlineconf', message => $message)->create();
    return;
}

sub on_update {
    my ($class, $module, $key, $value, $comment) = @_;
    my $message = sprintf 'Изменен параметр %s:%s, новое значение "%s".', _d $module, _d $key, _d $value;
    $message .= ' ' . _d $comment if $comment;
    MR::ChangeBot::Notification->new(origin => 'onlineconf', message => $message)->create();
    return;
}

sub on_delete {
    my ($class, $module, $key) = @_;
    my $message = sprintf 'Удален параметр %s:%s.', _d $module, _d $key;
    MR::ChangeBot::Notification->new(origin => 'onlineconf', message => $message)->create();
    return;
}

1;
