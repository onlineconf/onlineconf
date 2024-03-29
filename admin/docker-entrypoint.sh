#!/bin/sh

if [ -z "$ONLINECONF_AUTH_HEADER" ]; then 
  MYSQL_AUTH="mysql"
fi

echo "# Generated by docker-entrypoint.sh
http:
  listen: ${ONLINECONF_TLS_CERT:+':443'
  tls:
    cert: ${ONLINECONF_TLS_CERT}
    key: ${ONLINECONF_TLS_KEY}
    redirect_from_http: }':80'
  static_root: /var/www/onlineconf-admin
  ${ONLINECONF_AUTH_HEADER:+behind_reverse_proxy: true}
database:
  host: ${ONLINECONF_DATABASE_HOST}
  base: ${ONLINECONF_DATABASE_BASE:-onlineconf}
  user: ${ONLINECONF_DATABASE_USER:-onlineconf}
  password: ${ONLINECONF_DATABASE_PASSWORD}
  timeout: 2
auth:
  ${ONLINECONF_AUTH_HEADER:+method: header
  header: ${ONLINECONF_AUTH_HEADER}}
  ${MYSQL_AUTH:+method: mysql
  host: ${ONLINECONF_AUTH_HOST:-${ONLINECONF_DATABASE_HOST}}
  base: ${ONLINECONF_AUTH_BASE:-${ONLINECONF_DATABASE_BASE:-onlineconf}}
  user: ${ONLINECONF_AUTH_USER:-${ONLINECONF_DATABASE_USER:-onlineconf}}
  password: ${ONLINECONF_AUTH_PASSWORD:-${ONLINECONF_DATABASE_PASSWORD}}
  table: mod_auth
  name_field: Name
  password_field: Password
  timeout: 2}
  realm: onlineconf
notification_database:
  host: ${ONLINECONF_NOTIFICATION_HOST:-${ONLINECONF_DATABASE_HOST}}
  base: ${ONLINECONF_NOTIFICATION_BASE:-${ONLINECONF_DATABASE_BASE:-onlineconf}}
  user: ${ONLINECONF_NOTIFICATION_USER:-${ONLINECONF_DATABASE_USER:-onlineconf}}
  password: ${ONLINECONF_NOTIFICATION_PASSWORD:-${ONLINECONF_DATABASE_PASSWORD}}
  timeout: 2
" > /usr/local/etc/onlineconf-admin.yaml

exec /usr/local/bin/onlineconf-admin
