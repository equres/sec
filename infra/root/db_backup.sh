. /root/.creds

date=$(date +%Y%m%d)

set -x
set -e

echo "Doing the backup in /home/sec/db_backup/db_$date" && pg_basebackup -D /home/sec/db_backup/db_$date -U sec && 
echo "Change dir to /home/sec/db_backup/db_$date" && cd /home/sec/db_backup && 
echo "Compressing to db_$date.tar.xz--doing" && tar cfJ db_$date.tar.xz--doing ./db_$date &&
echo "Moving/Renaming compressed file to ../db_$date.tar.xz" && mv db_$date.tar.xz--doing db_$date.tar.xz &&
echo "Deleting uncompressed files at /home/sec/db_backup/db_$date" && cd .. && rm -r /home/sec/db_backup/db_$date &&
lftp -c "set ftp:list-options -a;
set ssl:check-hostname no;
open ftp://$login:$password@$FTP_server; 
lcd /home/sec/db_backup;
mkdir db_backup
cd db_backup/;
mirror --reverse --delete --use-cache --verbose --allow-chown  --allow-suid --no-umask --parallel=2;"
