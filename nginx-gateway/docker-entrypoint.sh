#!/bin/sh
set -e

CONSUL_ADDR="${CONSUL_HTTP_ADDR:-consul:8500}"

# Generar el nginx.conf para que no lea algo vacío
consul-template \
  -consul-addr "${CONSUL_ADDR}" \
  -template "/etc/consul-templates/nginx.conf.ctmpl:/etc/nginx/nginx.conf" \
  -once

# Arrancar nginx en segundo plano
nginx -g 'daemon off;' &

# Dejar consul-template en primer plano
exec consul-template \
  -consul-addr "${CONSUL_ADDR}" \
  -template "/etc/consul-templates/nginx.conf.ctmpl:/etc/nginx/nginx.conf:nginx -s reload"
