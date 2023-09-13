#!/bin/bash

ENVIRONMENT="$1"
if [ -z "${ENVIRONMENT:-}" ]; then
    echo "Usage: renew.sh <ENVIRONMENT>"
    exit 1
fi

path="chart/infra-server/configuration/$ENVIRONMENT/tls"
mkdir -p "$path"
openssl genrsa -out "$path/key.pem" 4096
openssl req -nodes -new -x509 -sha256 -days 3650 -config scripts/cert/tls.cnf -extensions 'req_ext' -key "$path/key.pem" -out "$path/cert.pem"
