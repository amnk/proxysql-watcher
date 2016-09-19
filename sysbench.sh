ACTION=$1

mysql_user="proxyuser"
mysql_pass="XTg1o9g8aP"
mysql_host="mariadb"

mysql -u $mysql_user -p$mysql_pass -h $mysql_host -P 3306 -e "CREATE DATABASE IF NOT EXISTS sbtest"

sysbench --num-threads=4 --max-time=20 --max-requests=0 --test=oltp \
         --oltp-reconnect-mode=transaction --oltp-auto-inc=off \
         --oltp-table-size=50000 \
         --db-driver=mysql  --db-ps-mode=disable \
         --mysql-user="$mysql_user" --mysql-password="$mysql_pass" \
         --mysql-host=$mysql_host --mysql-port=3306 \
         $ACTION
