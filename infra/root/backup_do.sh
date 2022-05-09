#!/bin/sh

set -x
set -e

sec=/home/sec/sec

date=$(date +%Y%m%d)

# Transfer backup files to ca2 server
status=failed
if rsync -av /mnt/sec/cache ubuntu@158.69.54.122:/mnt/backups; then
    status=success
fi
$sec event --event cron --job ca2_rsync --status $status --config /home/sec/.config/sec

# Transfer backup files to waw1 server
status=failed
if rsync -av /mnt/sec/cache ubuntu@79.137.68.227:/mnt/backups; then
    status=success
fi
$sec event --event cron --job waw1_rsync --status $status --config /home/sec/.config/sec

pg_basebackup -D /home/backups/db_backup/db_$date -z -X fetch -F tar
status=failed
FILE=/home/backups/db_backup/db_$date.tar.xz
if [ -f "$FILE" ]; then
    $sec event --event cron --job db_backup --status $status --config /home/sec/.config/sec

	if scp /home/backups/db_backup/db_$date.tar.xz ubuntu@158.69.54.122:/mnt/backups; then
        status=success
    fi
    $sec event --event cron --job ca2_db_scp --status $status --config /home/sec/.config/sec

    if scp /home/backups/db_backup/db_$date.tar.xz ubuntu@79.137.68.227:/mnt/backups; then
        status=success
    fi
    $sec event --event cron --job waw1_db_scp --status $status --config /home/sec/.config/sec
fi

# Delete uncompleted files that are older than 2 days
find /home/backups -name "*--doing.tar.xz" -type f -mtime +5 -delete
