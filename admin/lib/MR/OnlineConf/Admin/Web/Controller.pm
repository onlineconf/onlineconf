package MR::OnlineConf::Admin::Web::Controller;
use Mojo::Base 'Mojolicious::Controller';

sub rendered {
    my ($self, $status) = @_;
    $self->SUPER::rendered($status);
    if ($self->res->is_status_class(500)) {
        MR::OnlineConf::Admin::Storage->rollback();
    } else {
        MR::OnlineConf::Admin::Storage->commit();
    }
    return $self;
}

sub render_exception {
    my ($self, $e) = @_;
    $e = Mojo::Exception->new($e);
    $self->app->log->error($e);
    $self->render(json => { result => 'Error', message => $e->message }, status => 500);
    return;
}

sub render_access_denied {
    my ($self) = @_;
    $self->render(json => { result => 'error', message => 'Access denied' }, status => 403);
    return;
}

1;
