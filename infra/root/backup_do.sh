#!/bin/sh

set -x
set -e

sec=/home/sec/sec

/root/backup_store.sh
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job backup_store --status $status

/root/backup_restore.sh
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job backup_restore --status $status

/root/backup_verify.sh
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job backup_verify --status $status

/root/db_backup.sh
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job db_backup --status $status
