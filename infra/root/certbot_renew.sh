#!/bin/sh

set -x
set -e

sec=/home/sec/sec

/usr/bin/certbot renew --pre-hook "service nginx stop" --post-hook "service nginx start"
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job certbot_renew --status $status