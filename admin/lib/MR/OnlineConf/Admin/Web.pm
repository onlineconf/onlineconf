package MR::OnlineConf::Admin::Web;
use Mojo::Base 'Mojolicious';
use YAML;
use MR::OnlineConf::Admin::Storage;
use MR::ChangeBot::Database;

sub startup {
    my ($self) = @_;
    my $config = YAML::LoadFile('/usr/local/etc/onlineconf.yaml');
    MR::OnlineConf::Admin::Storage->new(%{$config->{database}}, log => $self->log);
    $config->{notification_database}->{database} ||= $config->{notification_database}->{base};
    MR::ChangeBot::Database->new(%{$config->{notification_database}}, log => $self->log);
    my $web_config = YAML::LoadFile('/usr/local/etc/onlineconf-admin.yaml');
    $self->plugin('mysql_basic_auth', %{$web_config->{auth}});
    my $r = $self->authenticate('/');
    $r->route('/config/(*path)')->via('GET')->to('config#get', path => '');
    $r->route('/config/(*path)')->via('POST')->to('config#set', path => '');
    $r->route('/config/(*path)')->via('DELETE')->to('config#delete', path => '');
    $r->route('/search')->via('GET')->to('config#search');
    $r->route('/global-log')->via('GET')->to('log#global');
    $r->route('/log/(*path)')->via('GET')->to('log#list', path => '');
    my $group = $r->bridge('/group')->to('group#can_edit');
    $group->route('/')->via('GET')->to('group#list');
    $group->route('/(.group)')->via('GET')->to('group#get');
    $group->route('/(.group)')->via('POST')->to('group#create');
    $group->route('/(.group)')->via('DELETE')->to('group#delete');
    $group->route('/(.group)/(.user)')->via('POST')->to('group#add_user');
    $group->route('/(.group)/(.user)')->via('DELETE')->to('group#delete_user');
    $r->route('/user')->via('GET')->to('group#user_list');
    $r->route('/whoami')->via('GET')->to('group#whoami');
    $r->route('/access/(*path)')->via('GET')->to('access#list', path => '');
    $r->route('/access/(*path)')->via('POST')->to('access#set', path => '');
    $r->route('/access/(*path)')->via('DELETE')->to('access#delete', path => '');
    $r->route('/monitoring')->via('GET')->to('monitoring#list');
    my $validate = $r->bridge('/client')->to('updater#validate');
    $validate->route('/config')->via('GET')->to('updater#config');
    $validate->route('/activity')->via('POST')->to('updater#activity');
    return;
}

1;
