#!/bin/sh

set -x
set -e

sec=/home/sec/sec

$sec regen pages --verbose --syslog
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job pages_cached --status $status