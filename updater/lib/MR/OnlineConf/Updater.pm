package MR::OnlineConf::Updater;

use Mouse;
use Scalar::Util 'weaken';
use Sys::Hostname ();
use MR::OnlineConf::Updater::Admin;
use MR::OnlineConf::Updater::PerlMemory;
use MR::OnlineConf::Updater::Parameter;
use MR::OnlineConf::Updater::ConfFiles;

our $VERSION = '1.0';

has config => (
    is  => 'ro',
    isa => 'HashRef',
    required => 1,
);

has log => (
    is  => 'ro',
    isa => 'Log::Dispatch',
    required => 1,
);

has _signals => (
    is  => 'ro',
    isa => 'ArrayRef',
        lazy    => 1,
    default => sub {
        my ($self) = @_;
        my $el = $self->_eventloop;
        my $log = $self->log;
        return [
            map {
                my $signal = $_;
                AnyEvent->signal(
                    signal => $signal,
                    cb     => sub {
                        $log->info("SIG$signal received, terminating...\n");
                        $el->send();
                        return;
                    }
                )
            } 'INT', 'TERM'
        ];
    }
);

has _eventloop => (
    is  => 'ro',
    lazy    => 1,
    default => sub { AnyEvent->condvar() },
);

has _timer => (
    is  => 'ro',
    lazy    => 1,
    default => sub {
        my ($self) = @_;
        weaken($self);
        AnyEvent->timer(
            interval => $self->config->{update_interval} || 5,
            cb       => sub { $self->update() },
        );
    },
);

has _online_timer => (
    is  => 'ro',
    lazy    => 1,
    default => sub {
        my ($self) = @_;
        weaken($self);
        AnyEvent->timer(
            interval => $self->config->{online_interval} || 60,
            cb       => sub { $self->_update_activity() },
        );
    },
);

has conf_files => (
    is  => 'ro',
    isa => 'MR::OnlineConf::Updater::ConfFiles',
    lazy    => 1,
    default => sub { MR::OnlineConf::Updater::ConfFiles->new(dir => $_[0]->config->{data_dir}, log => $_[0]->log) },
);

has _tree => (
    is  => 'ro',
    isa => 'MR::OnlineConf::Updater::PerlMemory',
    lazy    => 1,
    default => sub { MR::OnlineConf::Updater::PerlMemory->new(log => $_[0]->log) },
    clearer => '_clear_tree',
);

has _mtime => (
    is  => 'rw',
    isa => 'Str',
    trigger => sub { $_[0]->_update_activity() },
);

has _update_time => (
    is  => 'rw',
    isa => 'Int',
    default => 0,
);

has admin => (
    is => 'ro',
    isa => 'MR::OnlineConf::Updater::Admin',
    lazy => 1,
    default => sub {
        my ($self) = @_;

        return MR::OnlineConf::Updater::Admin->new(
            %{$self->config->{admin}}
        );
    }
);

sub run {
    my ($self) = @_;
    $self->_signals;
    $self->_timer;
    $self->_online_timer;
    $self->_eventloop->recv();
    return;
}

sub initialize {
    my ($self) = @_;
    $self->log->debug("Initializing config tree");
    my $mtime = $self->admin->get_mtime();
    return unless $mtime;
    my $list = $self->admin->get_config();
    return unless $list;
    my $count = @$list;
    $self->_clear_tree();
    my $tree = $self->_tree;
    foreach my $row (@$list) {
        my $param = MR::OnlineConf::Updater::Parameter->new(
            id           => $row->{ID},
            name         => $row->{Name},
            path         => $row->{Path},
            version      => $row->{Version},
            data         => $row->{Value},
            content_type => $row->{ContentType},
        );
        $tree->put($param);
    }
    $tree->finalize();
    if (eval { $self->conf_files->update($tree); 1 }) {
        $self->_mtime($mtime);
        $self->log->info("Config tree was initialized with $count parameters, last modification was at $mtime");
    } else {
        $self->log->error("Failed to initialize config: $@");
    }
    return;
}

sub update {
    my ($self) = @_;
    my $update_time = time();
    my $reselect = $self->config->{reselect_interval} || $self->config->{update_interval} * 2;
    if ((my $mtime = $self->_mtime) && $self->_update_time > time() - $reselect) {
        my $tree = $self->_tree;
        my $list = $self->admin->get_config($mtime, $reselect);
        if ($list && @$list) {
            my $count = 0;
            foreach my $row (@$list) {
                my $param = MR::OnlineConf::Updater::Parameter->new(
                    id           => $row->{ID},
                    name         => $row->{Name},
                    path         => $row->{Path},
                    version      => $row->{Version},
                    data         => $row->{Value},
                    content_type => $row->{ContentType},
                );
                if ($row->{Deleted}) {
                    $count++ if $tree->delete($param);
                } else {
                    $count++ if $tree->put($param);
                }
            }
            if ($count) {
                $tree->finalize();
                if (eval { $self->conf_files->update($tree); 1 }) {
                    my $mtime = $list->[-1]->{MTime};
                    $self->_mtime($mtime);
                    $self->log->info("Updated $count versions, last modification was at $mtime");
                } else {
                    $self->log->error("Failed to update config: $@");
                }
            } else {
                $self->log->debug("Nothing to update");
            }
        } else {
            $self->log->debug("Nothing to update");
        }
    } else {
        $self->initialize();
    }
    $self->_update_time($update_time);
    return;
}

sub _update_activity {
    my ($self) = @_;
    $self->admin->post_activity(
        Sys::Hostname::hostname(), $self->_mtime || 0, $VERSION
    );
    return;
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
