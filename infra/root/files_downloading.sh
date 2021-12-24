#!/bin/sh

set -x
set -e

sec=/home/sec/sec

$sec dow index --verbose --syslog
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job dow_index --status $status

# let's rewrite the bottom to look like the top ^^^

#$sec dowz --verbose --syslog
#$sec index --verbose --syslog
#$sec indexz --verbose --syslog 
#
#$sec event --event cron --job files_downloading --status failed
