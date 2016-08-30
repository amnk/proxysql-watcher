ACTION=$1

mysql -u proxyuser -ps3cret -h pxc-service -P 3306 -e "CREATE DATABASE IF NOT EXISTS sbtest"

sysbench --num-threads=4 --max-time=10 --test=oltp --db-ps-mode=disable \
         --mysql-user='proxyuser' --mysql-password='s3cret' \
         --oltp-table-size=10000  --oltp-reconnect-mode=transaction --oltp-auto-inc=off \
         --mysql-host=pxc-service --mysql-port=3306 \
         $ACTION
