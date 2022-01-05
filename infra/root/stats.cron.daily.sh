set -x
set -e

sec=/home/sec/sec

perl /usr/lib/cgi-bin/awstats.pl -config=equres.com -output > /home/sec/_stats/$(date +%F)-stats.html
status=failed
if [ $# -eq 0 ]; then
	status=success
fi
$sec event --event cron --job stats --status $status