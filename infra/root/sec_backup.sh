. .creds

lftp -c "set ftp:list-options -a;
open ftp://$login:$password@$FTP_server; 
cd /mnt/sec/cache;
mirror --reverse --delete --use-cache --verbose --allow-chown  --allow-suid --no-umask --parallel=2;"