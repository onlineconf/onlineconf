package MR::OnlineConf::Util;

use strict;
use warnings;
use Exporter;
use JSON::XS qw//;
use base qw/Exporter/;
use vars qw/@EXPORT_OK/;

@EXPORT_OK = qw/from_json to_json/;

eval { JSON::XS::to_json({}) };

if ($@){
    *from_json = \&JSON::XS::decode_json;
    *to_json = \&JSON::XS::encode_json;
}else{
    *from_json = \&JSON::XS::from_json;
    *to_json = \&JSON::XS::to_json;
}

1;
