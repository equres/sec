#!/bin/sh

set -x
set -e

sec=/home/sec/sec

date=$(date +%Y%m%d)

# Transfer backup files to backup server
status=failed
if rsync -av --dry-run /mnt/sec/cache /home/backups; then
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
find /home/backups -name "*--doing.tar.xz" -type f -mtime +5 -delete
