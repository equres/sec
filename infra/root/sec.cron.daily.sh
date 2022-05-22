#!/bin/sh

set -x
set -e

sec=/home/sec/sec

$sec de $(date +%Y) --verbose --syslog

$sec dow index --verbose --syslog
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job dow_index --status $status

$sec dowz --verbose --syslog
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job dowz --status $status

$sec index --verbose --syslog
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job index --status $status

$sec indexz --verbose --syslog
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job indexz --status $status
