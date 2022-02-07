#!/bin/sh

set -x
set -e

sec=/home/sec/sec

date=$(date +%Y%m%d)

# Create Cache Backup
tar -cJf cache_$date.tar.xz /mnt/sec/cache
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job cache_compressing --status $status

echo "Doing the backup in /home/backups/db_backup/db_$date" && pg_basebackup -D /home/backups/db_backup/db_$date && 
echo "Change dir to /home/backups/db_backup/db_$date" && cd /home/backups/db_backup && 
echo "Compressing to db_$date.tar.xz--doing" && tar cfJ db_$date--doing.tar.xz ./db_$date &&
echo "Deleting uncompressed files at /home/backups/db_backup/db_$date" && cd .. && rm -r /home/backups/db_backup/db_$date &&
echo "Moving/Renaming compressed file to ../db_$date.tar.xz" && mv db_$date--doing.tar.xz db_$date.tar.xz
status=failed
FILE=/home/backups/db_backup/db_$date.tar.xz
if [ -f "$FILE" ]; then
	status=success
fi
$sec event --event cron --job db_backup --status $status
