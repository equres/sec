. .creds

lftp -c "set ftp:list-options -a;
set ssl:check-hostname no;
open ftp://$login:$password@$FTP_server; 
lcd /mnt/sec/cache;
mirror --dry-run --verbose --allow-chown  --allow-suid --no-umask;" | logger