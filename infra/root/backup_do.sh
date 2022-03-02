#!/bin/sh

set -x
set -e

sec=/home/sec/sec

date=$(date +%Y%m%d)

# Create Cache Backup
tar -cJf /home/backups/cache_$date--doing.tar.xz /mnt/sec/cache
mv /home/backups/cache_$date--doing.tar.xz /home/backups/cache_$date.tar.xz
status=failed
FILE=/home/backups/db_backup/db_$date.tar.xz
if [ -f "$FILE" ]; then
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

# Delete uncompleted files that are older than 2 days
find /home/backups -name "*--doing.tar.xz" -type f -mtime +2 -delete
