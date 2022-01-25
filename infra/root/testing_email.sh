#!/bin/sh

set -x
set -e

sec=/home/sec/sec
job=testing_email

# Must Fail
$sec dow --verbose --syslog
status=success
if [ $? > 0 ]; then
    status=failed
    echo $job - $status at $(date) | mail -s "$job - $status at $(date)" hazem.xbox@gmail.com wkoszek@gmail.com
fi
$sec event --event cron --job testing_email --status $status

