. .creds

lftp -c "set ftp:list-options -a;
set ssl:check-hostname no;
open ftp://$login:$password@$FTP_server; 
lcd /mnt/sec/cache;
mirror --delete --use-cache --verbose --allow-chown  --allow-suid --no-umask --parallel=2;"