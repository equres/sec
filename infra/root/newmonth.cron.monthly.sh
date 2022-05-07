#!/bin/sh

set -x
set -e

todayDate=`date '+%d'`
sec=/home/sec/sec

if [ $todayDate -ne 01 ]; then
    return
fi

prevMonth=`date -d '-1 month' '+%m'`
prevYear=`date -d '-1 month' '+%Y'`

$sec de $prevYear/$prevMonth --verbose --syslog
