filesCount=`psql -t -U sec -c 'SELECT COUNT(*) FROM sec.secItemFile'`
companiesCount=`psql -t -U sec -c 'SELECT COUNT(DISTINCT companyname) FROM sec.secitemfile;'`
downloadsTodayCount=`psql -t -U sec -c "SELECT COUNT(*) FROM sec.events WHERE DATE(created_at) = current_date AND ev->>'event' = 'download';"`
sizeCache=`du -sh /mnt/sec/cache`
sizeUnzippedCache=`du -sh /mnt/sec/unzipped_cache/`

echo "
    Files Count in DB: $filesCount
    Companies Count in DB: $companiesCount
    Downloads Completed on Todays Date ($(date +%F)): $downloadsTodayCount
    Size of Cache (Raw, Index, & ZIP Files): $sizeCache
    Size of Unzipped Cache: $sizeUnzippedCache
" | mail -s "[equres.com] $(date +%F) daily report" sec@localhost