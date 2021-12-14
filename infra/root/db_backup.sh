date=$(date +%Y%m%d)

echo "Doing the backup in /home/sec/db_backup/db_$date" && pg_basebackup -D /home/sec/db_backup/db_$date -U sec && 
echo "Change dir to /home/sec/db_backup/db_$date" && cd /home/sec/db_backup && 
echo "Compressing to db_$date.tar.gz--doing" && tar -zcf db_$date.tar.gz--doing ./db_$date && 
echo "Moving/Renaming compressed file to ../db_$date.tar.gz" && mv db_$date.tar.gz--doing db_$date.tar.gz &&
echo "Deleting uncompressed files at /home/sec/db_backup/db_$date" && cd .. && rm -r /home/sec/db_backup/db_$date