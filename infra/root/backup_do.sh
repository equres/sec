#!/bin/sh

set -x
set -e

sec=/home/sec/sec

date=$(date +%Y%m%d)

# Create Cache Backup
tar -cJf cache_$date--doing.tar.xz /mnt/sec/cache
mv cache_$date--doing.tar.xz cache_$date.tar.xz
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job cache_compressed --status $status --config /home/sec/.config/sec

pg_basebackup -D /home/backups/db_backup/db_$date -z -X fetch -F tar
status=failed
FILE=/home/backups/db_backup/db_$date.tar.xz
if [ -f "$FILE" ]; then
	status=success
fi
$sec event --event cron --job db_backup --status $status --config /home/sec/.config/sec
