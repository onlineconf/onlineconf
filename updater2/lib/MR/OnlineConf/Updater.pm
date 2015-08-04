package MR::OnlineConf::Updater;

use Mouse;
use Scalar::Util 'weaken';
use Sys::Hostname ();
use MR::OnlineConf::Updater::Admin;
use MR::OnlineConf::Updater::PerlMemory;
use MR::OnlineConf::Updater::Parameter;
use MR::OnlineConf::Updater::ConfFiles;

our $VERSION = '2.0';

has log => (
    is  => 'ro',
    isa => 'Log::Dispatch',
    required => 1,
);

has config => (
    is  => 'ro',
    isa => 'HashRef',
    required => 1,
);

has _loop => (
    is  => 'ro',
    lazy    => 1,
    default => sub { AnyEvent->condvar() },
);

has _signals => (
    is  => 'ro',
    isa => 'ArrayRef',
    lazy    => 1,
    default => sub {
        my ($self) = @_;
        my $el = $self->_loop;
        my $log = $self->log;
        return [
            map {
                my $signal = $_;
                AnyEvent->signal(
                    signal => $signal,
                    cb     => sub {
                        $log->info("SIG $signal received, terminating...\n");
                        $el->send();
                        return;
                    }
                )
            } 'INT', 'TERM'
        ];
    }
);

has _update_timer => (
    is  => 'ro',
    lazy    => 1,
    default => sub {
        my ($self) = @_;
        weaken($self);
        AnyEvent->timer(
            interval => $self->config->{update_interval} || 5,
            cb       => sub { $self->_update_config() },
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
            cb       => sub { $self->_update_online() },
        );
    },
);

has _admin => (
    is => 'ro',
    isa => 'MR::OnlineConf::Updater::Admin',
    lazy => 1,
    default => sub {
        my ($self) = @_;

        return MR::OnlineConf::Updater::Admin->new(
            %{$self->config->{admin}}, version => $VERSION
        );
    }
);

has _reselect => (
    is => 'ro',
    isa => 'Int',
    lazy => 1,
    default => sub {
        my ($self) = @_;
        return $self->config->{reselect_interval} || $self->config->{update_interval} * 2;
    }
);

has conf_files => (
    is  => 'ro',
    isa => 'MR::OnlineConf::Updater::ConfFiles',
    lazy    => 1,
    default => sub {
        return MR::OnlineConf::Updater::ConfFiles->new(
            log => $_[0]->log,
            dir => $_[0]->config->{data_dir}
        )
    },
);

has _update_time => (
    is  => 'rw',
    isa => 'Int',
    default => 0,
);

sub run {
    my ($self) = @_;
    $self->_signals;
    $self->_update_timer;
    $self->_online_timer;
    $self->_loop->recv();
    return;
}

sub _update_config {
    my ($self) = @_;

    $self->_update_time(time);

    if ($self->_is_need_update) {
        my $list = $self->_admin->get_config();
        my $tree = MR::OnlineConf::Updater::PerlMemory->new(log => $_[0]->log);

        if ($list && @$list) {
            my $count = 0;
            my @slist = sort {$a->{Path} cmp $b->{Path}} @$list;

            foreach my $row (@slist) {
                my $param = MR::OnlineConf::Updater::Parameter->new(
                    id           => $row->{ID},
                    name         => $row->{Name},
                    path         => $row->{Path},
                    version      => $row->{Version},
                    data         => $row->{Value},
                    content_type => $row->{ContentType},
                );

                if (defined $param->value || $param->is_null || $param->is_case) {
                    $count++ if $tree->put($param);
                }
            }

            if ($count) {
                if (eval { $self->conf_files->update($tree); 1 }) {
                    $self->log->info("Updated $count versions, last modification was at " . $self->_admin->mtime);
                } else {
                    $self->log->error("Failed to update config: $@");
                }
            } else {
                $self->log->debug("Nothing to update");
            }
        } else {
            $self->log->debug("Nothing to update");
        }
    }

    return;
}

sub _update_online {
    my ($self) = @_;
    return $self->_admin->post_activity();
}

sub _is_need_update {
    my ($self) = @_;

    return 0 unless $self->_admin->mtime;
    return 0 unless $self->_update_time > time() - $self->_reselect;

    return 1;
}

no Mouse;
__PACKAGE__->meta->make_immutable();

1;
