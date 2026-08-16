#!/bin/sh
set -e

cp /etc/harbor-ca/ca.crt /usr/local/share/ca-certificates/harbor-ca.crt
update-ca-certificates

exec su-exec app /app/image-signer
