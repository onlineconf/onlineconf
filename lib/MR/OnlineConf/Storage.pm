package MR::OnlineConf::Storage;

use strict;
use MR::OnlineConf::Const qw/:all/;
use MR::OnlineConf::Util qw/from_json to_json/;
use DBI; 
use Data::Dumper;
use Carp qw/cluck/; 

sub CURRENT_TABLE()     {'my_config'}
sub LOG_TABLE()         {'my_config_log'}
sub TRANSACTION_TABLE() {'my_config_transaction'}
sub MODULE_TABLE()      {'my_config_module'}

sub DEBUG()             { ref $_[0] ? $_[0]->{config}{debug} : 0 }
    
sub new {
    my ($class,$opts) = @_;
    my $self = bless {} , $class;
    $self->{config} = $opts;
    return $self;
}

sub _connect_opts {
    my ($self) = @_;
    return (sprintf("dbi:mysql:%s:%s", $self->{config}{database}{base}, $self->{config}{database}{host}) , 
                                       $self->{config}{database}{user} ,
                                       $self->{config}{database}{password} );
}

sub _dbinfo {
    my ($self) = @_;
    my ($dsn , $user, $password) = $self->_connect_opts;
    return "[ DB: $dsn , user $user , password ".($password ? 'YES' : 'NO')."]";
}

sub _dbh {
    my ($self)  = @_;
    return $self->{dbh} if $self->{dbh};

    $self->{dbh} = $self->_db_wrap(
            sub {
                $self->_debug(0,"reconnect to database ".$self->_dbinfo()."\n");
                return DBI->connect($self->_connect_opts, { RaiseError => 1, AutoCommit => 1 })
            });

    $self->_debug(-1,"can't connect to database: ".$self->_dbinfo().": ".$DBI::errstr)
            unless $self->{dbh};
    return $self->{dbh};
}

# minimun interface for replace mDBI
sub _db_warn {
    my ($self) = @_;
    cluck "DB ERROR: ".($self->{dbh} ? "[ ".$self->{dbh}->err()." ] ".$self->{dbh}->errstr : $DBI::errstr )." database ".$self->_dbinfo;
    return 1; 
} 

sub _db_rollback {
    my ($self) = @_;
    $self->_debug(-1,"cant rollback without connection") and return 1 
        unless $self->{dbh};
    return $self->_db_wrap('rollback');
}

sub _db_commit {
    my ($self) = @_;
    $self->_debug(-1,"cant rollback without connection") and return 1 
        unless $self->{dbh};
    return $self->_db_wrap('commit');
}

sub _db_disconnect {
    my ($self) = @_;
    return unless $self->{dbh};
    #$self->_db_rollback();
    $self->{dbh}->disconnect();
    $self->{dbh} = undef;
    $self->_debug(-1,"db ".$self->_dbinfo()." disconnected\n");
}

sub _db_wrap {
    my ($self , $sub) = (shift,shift);
    my $timeout = $self->{config}{database}{timeout} || 2;
    local $SIG{ALRM} = sub { die "Database timeout ($timeout sec.)\n" };
    my $result;
    if (ref $sub eq 'CODE'){
        alarm($timeout);
        $result = eval { $sub->(@_) };
    }else{
        my $dbh = $self->_dbh() or return undef;
        alarm($timeout);
        $result = eval { $dbh->$sub(@_) };
    }
    alarm(0);
    $self->_debug(-1,$@) if $@;
    return $result;
}

sub _db_ping {
    my ($self) = @_;
    return 1 if $self->_db_wrap('ping');
    $self->_debug(-1,"cant ping database ".$self->_dbinfo()."\n");
    $self->_db_disconnect;
    return undef;
}

sub _db_begin {
    my ($self) = @_;
    my $dbh = $self->_dbh() or return;
    unless ($dbh->ping()){
        $self->_db_disconnect();
        return undef;
    }    
    return $self->_db_wrap('begin_work');
}

sub _db_update {
    my ($self,$q,@opts) = @_;
    my $dbh = $self->_dbh() or return undef;
    return $self->_db_wrap(
        sub {
            my $sth = $dbh->prepare($q) or return undef;
            my $c = 0;
            foreach my $p (@opts){
                $c++;
                $sth->bind_param($c,$p);
            }
            unless ($sth->execute){
                $self->_debug(-1,"cant execute db query $q [ @opts ]");
                return undef;
            }
            return 1;
        }
    );    
}

sub _db_query_array {
    my ($self,$q,@opts) = @_;
    my $dbh = $self->_dbh() or return undef;
    return $self->_db_wrap(
        sub {
            my $sth = $dbh->prepare($q) or return undef;
            my $c = 0;
            foreach my $p (@opts){
                $c++;
                $sth->bind_param($c,$p);
            }
            unless ($sth->execute){
                $self->_debug(-1,"cant execute db query $q [ @opts ]");
                return undef;
            }
            my @r;
            while (my $r = $sth->fetchrow_hashref){
                push @r , $r;
            }
            return \@r;
        }
    );
}

sub _debug {
    my ($self,$level,@msg) = @_;
    return if $self->DEBUG < $level;
    warn join "\n" , map {ref $_ ? Dumper($_) : $_} @msg;
}

# do not 
sub open {
    my ($self) = @_;
    return $self->_db_ping();
}

sub close {
    my ($self) = @_;
    return $self->_db_disconnect;
}

sub get {
    my ($self,$module,$key,%opts) = @_;
    %opts = (version => MY_CONFIG_CURRENT_VER , default=> undef , json2perl=>1 , %opts);
    my $r = $self->getMulti($module,[$key],%opts);
    return $opts{default} unless $r && ref $r eq 'HASH' && %$r;
    return +(values %$r)[0];
}

sub getMulti {
    my ($self,$module,$keys,%opts) = @_;
    %opts = (version => MY_CONFIG_CURRENT_VER,json2perl=>1,flags=>0,access=>MY_CONFIG_ACCESS_ANY,with_deleted=>0,%opts);
    my $mid = $self->module($module);
    warn "$self: undefined module $module\n" and return undef unless $mid;
    return undef unless $keys && ref $keys eq 'ARRAY' && @$keys;
    my $r;
    my $addon = '';
    my @addon_f;
    if ($opts{flags}){
        $addon .= ' AND (log.`Flags` & ?) ';
        push @addon_f , $opts{flags};
    }
    if ($opts{access}){
        unless ($opts{access} == MY_CONFIG_ACCESS_ANY){
            $addon .= ' AND (log.`Access` & ?) ';
            push @addon_f , $opts{access};
        }
    }
    unless ($opts{with_deleted}){
        $addon .= ' AND NOT (log.`Flags` & ?) ';
        push @addon_f , MY_CONFIG_DELETED_FLAG;
    }
    $opts{version} ||= MY_CONFIG_CURRENT_VER;
    if ($opts{version} == MY_CONFIG_CURRENT_VER){
        $r = $self->_db_query_array(
                "SELECT log.`Version`,log.`Key`,log.`Value`,log.`Flags`,log.`Access`,log.`Comment` FROM ".$self->CURRENT_TABLE.
                " as log WHERE log.`Module` = ? AND log.`Key` IN (".(join ',' , map {'?'} @$keys).") $addon",
                $mid->{ID}, @$keys, @addon_f);
    }else{
        $r = $self->_db_query_array(
                "SELECT log.`Version`,log.`Key`,log.`Value`,log.`Flags`,log.`Access`,log.`Comment`  FROM ".$self->LOG_TABLE." as log,
                    (SELECT max(log1.`Version`) as ver , log1.`Key`, log1.`Module` from ".$self->LOG_TABLE." as log1 
                        WHERE log1.`Module` = ? AND log1.`Version` <= ? AND log1.`Key` IN (".(join ',' , map {'?'} @$keys).") GROUP BY log1.`Key`,log1.`Module`)
                     as max_ver
                     WHERE log.`Module` = max_ver.`Module` AND log.`Key` = max_ver.`Key` AND log.`Version` = max_ver.ver $addon",
                 $mid->{ID},$opts{version},@$keys,@addon_f);
    }
    unless ($r && ref $r eq 'ARRAY'){    
        $self->_db_warn;
        return undef;
    }
    $r = {map {$_->{Key} => $_} @$r};
    $self->_unpack($r) if $opts{json2perl};
    return $r;
}

sub _unpack {
    my ($self,$r) = @_;
    foreach my $v (values %$r){
        next unless $v->{Flags} & MY_CONFIG_JSON_FLAG;
        my $c = eval {from_json($v->{Value})};
        if ($@){
            warn "$self: cant convert string `$v->{Value}` to perl (key `$v->{Key}` module `$v->{Module}`)\n";
            $v->{Value} = undef;
        }else{
            $v->{Value} = $c;
        }
    }
    return $r;
}

sub _strip_deleted {
    my ($self,$r)  = @_;
    foreach my $v (values %$r){
        delete $r->{$v->{Key}} if $v->{Flags} & MY_CONFIG_DELETED_FLAG;
    }
    return $r;
}

sub getAll {
    my ($self,$module,%opts) = @_;
    %opts = (version => MY_CONFIG_CURRENT_VER , json2perl=>1 , flags => 0 , access=>MY_CONFIG_ACCESS_ANY , with_deleted=>0,%opts);
    my $mid = $self->module($module);
    warn "$self: undefined module $module\n" and return undef unless $mid;
    my $r;

    my $addon = '';
    my @addon_f;
    if ($opts{flags}){
        $addon .= ' AND (log.`Flags` & ?) ';
        push @addon_f , $opts{flags};
    }
    if ($opts{access}){
        unless ($opts{access} == MY_CONFIG_ACCESS_ANY){
            $addon .= ' AND (log.`Access` & ?) ';
            push @addon_f , $opts{access};
        }
    }
    unless ($opts{with_deleted}){
        $addon .= ' AND NOT (log.`Flags` & ?) ';
        push @addon_f , MY_CONFIG_DELETED_FLAG;
    }
    $opts{version} ||= MY_CONFIG_CURRENT_VER;
    if ($opts{version} == MY_CONFIG_CURRENT_VER){
        $r = $self->_db_query_array(
                "SELECT log.`Version`,log.`Key`,log.`Value`,log.`Flags`,log.`Access`,log.`Comment`  FROM ".$self->CURRENT_TABLE.
                " as log WHERE `Module` = ? $addon ",$mid->{ID},@addon_f);
    }else{
       $r = $self->_db_query_array(
                "SELECT log.`Version`,log.`Key`,log.`Value`,log.`Flags`,log.`Access`,log.`Comment`  FROM ".$self->LOG_TABLE." as log,
                    (SELECT max(log1.`Version`) as ver , log1.`Key`, log1.`Module` from ".$self->LOG_TABLE." as log1 
                        WHERE log1.`Module` = ? AND log1.`Version` <= ? GROUP BY log1.`Key`,log1.`Module`)
                     as max_ver
                     WHERE log.`Module` = max_ver.`Module` AND log.`Key` = max_ver.`Key` AND log.`Version` = max_ver.ver $addon",
                 $mid->{ID},$opts{version},@addon_f);
    }
    unless ($r && ref $r eq 'ARRAY'){    
        $self->_db_warn;
        return undef;
    }
    $r = {map {$_->{Key} => $_} @$r};
    $self->_unpack($r) if $opts{json2perl};
    return $r;
}

sub getAllFromModules {
    my ($self,$mod) = @_;
    return {} unless $mod && ref $mod eq 'ARRAY' && @$mod;
    my $r = $self->_db_query_array(
                "SELECT log.`Key`,log.`Value`,log.`Flags`,log.`Access`,log.`Module`,log.`Comment` md.`Version` FROM ".$self->CURRENT_TABLE.
                " as log LEFT JOIN ".$self->MODULE_TABLE." as md ON md.`ID` = log.`Module`
                WHERE log.`Module` IN (".(join ',' , map {'?'} @$mod).")",
                @$mod) || $self->_db_warn;
    return undef unless $r && ref $r eq 'ARRAY';
    my %r;
    foreach my $i (@$r){
        $r{$i->{Module}}->{$i->{Key}} = $i;
    }    
    $self->_unpack($_) for values %r;
    return \%r;
}

sub _add {
    my ($self,$mid,$version,$values) = @_;
    foreach my $k (keys %$values){
        $values->{$k}{Flags} ||=0;
        $values->{$k}{Access} |= MY_CONFIG_ACCESS_ADMIN; 
        unless ($self->_db_update("INSERT INTO ".$self->CURRENT_TABLE."(`Version`,`Module`,`Key`,`Value`,`Flags`,`Access`,`Comment`) 
                                       VALUES(?,?,?,?,?,?,?)",
                                       $version,$mid,$k,$values->{$k}{Value},$values->{$k}{Flags},$values->{$k}{Access},$values->{$k}{Comment})){
            $self->_db_warn;
            return E_UNDEFINED;            
        }
    } 
    unless ($self->_db_update(
           "INSERT INTO ".$self->LOG_TABLE."(`Version`,`Module`,`Key`,`Value`,`Flags`,`Access`,`Comment`)
            SELECT `Version`,`Module`,`Key`,`Value`,`Flags`,`Access`,`Comment` FROM ".$self->CURRENT_TABLE."
                WHERE `Module` = ? AND `Key` IN (".(join "," , map {"?"} keys %$values).")",    
            $mid,keys %$values)){
        $self->_db_warn;
        return E_UNDEFINED;      
    }
    return undef;
}

# update or insert
sub _update {
    my ($self,$module,$version,$values)  = @_;
    foreach my $v (values %$values){
        next unless $v->{Flags} && $v->{Flags} & MY_CONFIG_JSON_FLAG;
        eval {from_json($v->{Value})};
        if ($@){
            warn "invalid json $v->{Value} for key $v->{Key}: $@\n";
            return E_BAD_DATA;
        }
    }
    foreach my $k (keys %$values){
        $values->{$k}{Flags} ||=0;
        $values->{$k}{Access} |= MY_CONFIG_ACCESS_ADMIN; 
        unless ($self->_db_update(
               "INSERT INTO ".$self->CURRENT_TABLE."(`Version`,`Module`,`Key`,`Value`,`Flags`,`Access`,`Comment`) 
                VALUES(?,?,?,?,?,?,?) ON
                DUPLICATE KEY UPDATE `Version` = ? ,`Value` = ?,`Flags` = ?,`Access`=?,`Comment`=?", 
                $version,$module,$k,$values->{$k}{Value},$values->{$k}{Flags},$values->{$k}{Access},$values->{$k}{Comment},
                $version,$values->{$k}{Value},$values->{$k}{Flags},$values->{$k}{Access},$values->{$k}{Comment})){
            $self->_db_warn;
            return E_UNDEFINED      
        }
    }
    my $p = join "," , map {'?'} keys %$values;
    unless ($self->_db_update(
           "INSERT INTO ".$self->LOG_TABLE."(`Version`,`Module`,`Key`,`Value`,`Flags`,`Access`,`Comment`)
            SELECT `Version`,`Module`,`Key`,`Value`,`Flags`,`Access`,`Comment` FROM ".$self->CURRENT_TABLE."
                WHERE `Module` = ? AND `Key` IN ($p)",    
            $module,keys %$values)){
        $self->_db_warn;
        return E_UNDEFINED;      
    }
    return undef;
}

sub _delete {
   my ($self,$module,$version,$val)  = @_;
   my $values = ref $val eq 'HASH' ? [keys %$val] : ref $val eq 'ARRAY' ? $val : [];
   my $p = join "," , map {'?'} @$values;
   unless ($self->_db_update(
           "INSERT INTO ".$self->LOG_TABLE."(`Version`,`Module`,`Key`,`Value`,`Flags`,`Access`,`Comment`)
            SELECT $version,`Module`,`Key`,`Value`,`Flags`|".MY_CONFIG_DELETED_FLAG.",`Access`,`Comment` FROM ".$self->CURRENT_TABLE."
                WHERE `Module` = ? AND `Key` IN ($p)",    
            $module,@$values)){
        $self->_db_warn;
        return E_UNDEFINED;      
   }
   unless ($self->_db_update("DELETE FROM ".$self->CURRENT_TABLE." WHERE `Module` = ? AND `Key` IN ($p)",$module,@$values)){
       $self->_db_warn;
       return E_UNDEFINED;
   }
   return undef;
}

sub _eq {
    return $_[0]->{Value} eq $_[1]->{Value} && 
           $_[0]->{Flags} == $_[1]->{Flags} &&
           $_[0]->{Access} == $_[1]->{Access} &&
           $_[0]->{Comment} eq $_[1]->{Comment};
           
}

sub _replaceAll {
   my ($self,$module,$version,$values)  = @_;
  
   foreach my $v (values %$values){
        $v->{Flags} = 0 unless defined $v->{Flags};
        $v->{Access} = MY_CONFIG_ACCESS_ADMIN unless defined $v->{Access}; 
        $v->{Comment} = '' unless defined $v->{Comment};
   }
 
   $self->_debug(3,"replace all " , $values);
   my $prev = $self->getAll($module,json2perl=>0);
   my $merged = {%$values,%$prev};
   $self->_debug(3,"prev values" , $prev);   

   my @to_rm;
   foreach my $k (keys %$prev){
       unless (exists $values->{$k}){
           $self->_debug(3,"key $k not exists in new config");
           push @to_rm , $k; 
           delete $merged->{$k};
       }else{
           delete $merged->{$k} if _eq($prev->{$k},$values->{$k});
       }
   }
   $self->_debug(3,"To remove " , \@to_rm);
   $self->_debug(3,"To update " , [keys %$merged]);
   return E_NO_CHANGES unless @to_rm || %$merged;
   my $r = $self->_delete($module,$version,\@to_rm) if @to_rm;
   return $r if $r;
   $r = $self->_update($module,$version,{map {$_=>$values->{$_}} keys %$merged}) if %$merged;
   return $r if $r;
   return undef;
}

sub _transaction { 
    my ($self,$module,$commiter,$log,$func,$values,%opts) = @_;
    return E_BAD_DATA unless $commiter && $log;
    my $mid = $self->module($module);
    warn "$self: undefined module $module\n" and return undef unless $mid;
    warn "$self: cant lock module $mid->{Name}" and  return $self->_unlock(E_UNDEFINED) unless $self->_lock($mid->{ID});
    my $version = $self->version($mid->{ID});
    $version ||=0;
    return $self->_unlock(E_PREV_VERSION_MISMATCH) if $opts{prev_version} && !($version == $opts{prev_version});
    my $r = $self->$func($mid->{ID},$version+1,$values);
    return $self->_unlock($r) if $r;
    unless ($self->_db_update(
                "INSERT INTO ".$self->TRANSACTION_TABLE.'(`Module`,`Version`,`Comment`,`ChangedBy`) VALUES(?,?,?,?)',
                $mid->{ID},$version+1,$log,$commiter)){
        $self->_db_warn;
        return $self->_unlock(E_UNDEFINED);        
    }
    unless ($self->_db_update(
                "UPDATE ".$self->MODULE_TABLE.' SET `Version` = ? WHERE `ID` = ?',
                $version+1,$mid->{ID})){
        $self->_db_warn;
        return $self->_unlock(E_UNDEFINED);        
    }
    return $self->_unlock();
}

sub add { 
    my ($self,$module,$values,%opts) = @_;
    return $self->_transaction($module,delete $opts{commiter},delete $opts{log},'_add',$values,%opts);
}

sub update { 
    my ($self,$module,$values,%opts) = @_;
    return $self->_transaction($module,delete $opts{commiter},delete $opts{log},'_update',$values,%opts);
}

sub delete { 
    my ($self,$module,$values,%opts) = @_;
    return $self->_transaction($module,delete $opts{commiter},delete $opts{log},'_delete',$values,%opts);
}

sub replaceAll {
    my ($self,$module,$values,%opts) = @_;
    return $self->_transaction($module,delete $opts{commiter},delete $opts{log},'_replaceAll',$values,%opts);
}


sub _lock {
    my ($self,$mid) = @_;
    unless ($self->_db_begin){
        $self->_db_warn;
        return undef;
    }
    my $r = $self->_db_query_array("SELECT `Version` FROM ".$self->MODULE_TABLE.
                ' WHERE `ID` = ? FOR UPDATE',$mid);
    unless ($r && ref $r eq 'ARRAY' && @$r){    
        $self->_db_warn unless $r;
        return undef;        
    }
    return 1;
}

sub _unlock {
    my ($self,$reason) = @_;
    if ($reason){
        warn "rollback";
        $self->_db_rollback() || $self->_db_warn;
    }else{
        $self->_db_commit() || $self->_db_warn;
    }
    return ($reason);
}

sub version {
    my ($self,$module) = @_;
    my $mid = $self->module($module);
    warn "$self: undefined module $module\n" and return undef unless $mid;
    return $mid->{Version};
}

sub log {
    my ($self,$module,%opts) = @_;
    my $mid = $self->module($module);
    warn "$self: undefined module $module\n" and return undef unless $mid;
    %opts = (version=>undef,limit=>undef,%opts);
    if ($opts{version} && $opts{version} == MY_CONFIG_CURRENT_VER){
        $opts{version} = $self->version();
    }
    my @f = ($mid->{ID});
    my $q = "SELECT `Version`,`ChangedBy`,`ChangedTime`,`Comment` FROM ".$self->TRANSACTION_TABLE." WHERE `Module` = ?";
    if ($opts{version}){
        $q.=" AND `Version` = ?";
        push @f , $opts{version};
    }
    $q.=' ORDER BY `Version` DESC ';
    if (defined $opts{limit}){
        $q.=" LIMIT ?";
        push @f , $opts{limit};
    }
    my $r = $self->_db_query_array($q,@f) || $self->_db_warn;
    return $r;
}

sub logAll {
    my ($self,%opts) = @_;
    %opts = (limit=>undef,%opts);
    my $q = "SELECT `Module`,`Version`,`ChangedBy`,`ChangedTime`,`Comment` FROM ".$self->TRANSACTION_TABLE.
            " WHERE `Module` > 0 ORDER BY `ChangedTime` DESC , Version DESC";
    my @f;        
    if (defined $opts{limit}){
        $q.=" LIMIT ?";
        push @f , $opts{limit};
    }
    my $r = $self->_db_query_array($q,@f) || $self->_db_warn;
    return $r;
}

sub diff {
    my ($self,$module,$left,$right,%opts) = @_;
    %opts = (json2perl=>1,with_deleted=>0,%opts);
    my $mid = $self->module($module);
    warn "$self: undefined module $module\n" and return undef unless $mid;
    return {} if $left == $right;
    my $swap;
    if ($left > $right){
        $swap = 1;
        ($left,$right) = ($right,$left);
    }
    my $r = $self->_db_query_array("SELECT log.`Version`,log.`Key`,log.`Value`,log.`Flags`,log.`Comment` FROM ".$self->LOG_TABLE." as log, 
            (SELECT max(log1.`Version`) as max_ver,log1.`Key`,log1.`Module`
                FROM ".$self->LOG_TABLE." as log1 WHERE log1.`Module` = ? AND log1.`Version` > ? AND log1.`Version` <=?
                GROUP BY log1.`Key`,log1.`Module`)
              as vers  
            WHERE log.`Module` = vers.`Module` AND log.`Version` = vers.max_ver AND log.`Key` = vers.`Key`
            ORDER BY log.`Version`",$mid->{ID},$left,$right) || $self->_db_warn;
    return {} unless $r && ref $r eq 'ARRAY';
    my $l = $self->getMulti($module,[map {$_->{Key}} @$r],version=>$left,json2perl=>0,with_deleted=>1);
    my %r;
    my($ll,$lr) = ($swap ? ('right','left') : ('left','right'));
    foreach my $i (@$r){
        $r{$i->{Key}}{$lr} = $i if $opts{with_deleted} || !($i->{Flags} & MY_CONFIG_DELETED_FLAG);
        $r{$i->{Key}}{$ll} = $l->{$i->{Key}} if exists $l->{$i->{Key}} && 
            ($opts{with_deleted} || !($l->{$i->{Key}}{Flags} & MY_CONFIG_DELETED_FLAG));
    }
    if ($opts{json2perl}){
        foreach my $i (keys %r){
            $self->_unpack($r{$i});
        }
    }
    return (\%r);
}

sub modules {
    my ($self) = @_;
    my $r = $self->_db_query_array("SELECT `ID`,`Name`,`Version`,`Comment` FROM ".$self->MODULE_TABLE." ORDER BY `Name`");
    unless ($r && ref $r eq 'ARRAY'){
        $self->_db_warn;
        return undef;
    }
    return $r;
}

sub module {
    my ($self,$id) = @_;
    if ($id=~/^\d+$/){
        my $r = $self->_module_by(id=>$id);
        return $r if $r;
    }
    return $self->_module_by(name=>$id);
}

sub _module_by {
    my ($self,%opt) = @_;
    my ($f,$v) = ($opt{id} ? ('ID',$opt{id}) : ('Name',$opt{name}));
    return undef unless $f && $v;
    my $r = $self->_db_query_array("SELECT `ID`,`Name`,`Version`,`Comment` FROM ".$self->MODULE_TABLE." WHERE `$f` = ?",$v) 
            || $self->_db_warn;
    return ($r && @$r ? $r->[0] : undef);
}

sub saveModule {
    my ($self,$name,$comment) = @_;
    $name =~s/^\s*(.+?)\s*$/$1/;
    return undef unless $name;
    unless ($self->_db_update("INSERT INTO ".$self->MODULE_TABLE."(`Name`,`Comment`) VALUES(?,?)",$name,$comment)){
        $self->_db_warn;
        return undef;
    }
    return 1;
}

sub revert {
    my ($self,$module,$version,%opts) = @_;
    %opts = (force=>0,%opts);
    my $mid = $self->module($module);
    warn "$self: undefined module $module\n" and return undef unless $mid;
    return undef unless $version && $version > 0;
    my $diff = $self->diff($module,$version-1,$version,json2perl=>0,with_deleted=>1);
    unless (%$diff){
        warn "no changes between ".($version-1)." and $version\n";
        return 0;
    }
    unless ($opts{force}){
        my $now = $self->getMulti($module,[keys %$diff],json2perl=>0,with_deleted=>1);
        my %changed;
        foreach my $k (keys %$diff){
            if($now->{$k}{Version} > $version){
                $changed{$k}->{right} = $now->{$k} unless  $now->{$k}{Flags} & MY_CONFIG_DELETED_FLAG;
                $changed{$k}->{left} = $diff->{$k}{left} if exists $diff->{$k}{left} && !($diff->{$k}{left}{Flags} & MY_CONFIG_DELETED_FLAG);
            }
        }
        return \%changed if %changed;
    }
    $opts{log}.='Revert v.'.$version.($opts{log} ? ': '.$opts{log} : '');
    foreach my $k (keys %$diff){
        $diff->{$_}{left}{Flags} = $diff->{$_}{left}{Flags} &~ MY_CONFIG_DELETED_FLAG if $diff->{$_}{left};
    }
    return $self->_transaction($module,delete $opts{commiter},delete $opts{log},'_revert',{map {$_=>$diff->{$_}{left}} keys %$diff},%opts);
}

sub _revert {
    my ($self,$mid,$version,$values) = @_;
    my @to_rm = grep { !($values->{$_}) } keys %$values;
    $self->_debug(3,"To remove " , \@to_rm);
    my %to_upd = map {$_=>$values->{$_}} grep {$values->{$_}} keys %$values;
    $self->_debug(3,"To update " , [keys %to_upd]);
    return E_NO_CHANGES unless @to_rm || %to_upd;
    my $r = $self->_delete($mid,$version,\@to_rm) if @to_rm;
    return $r if $r;
    $r = $self->_update($mid,$version,{map {$_=>$values->{$_}} keys %to_upd}) if %to_upd;
    return $r if $r;
    return 0;
}


1;
