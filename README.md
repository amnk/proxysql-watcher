Very basic and experimental watcher for ProxySQL: https://github.com/sysown/proxysql

Watcher monitors changes to specific Etcd prefix and populates those changes to ProxySQL.

For now a lot of values are hardcoded.

Use at your own risk!

## How to use
Create k8s configmap `percona-config`:
```
kind: ConfigMap
apiVersion: v1
metadata:
  name: percona-config
data:
  mysql-root-password: "JdUDnZyTsL"
  discovery-service: "etcd-client:2379"
  cluster-name: "k8scluster"
  xtrabackup-password: "p9jwtcv3WX"
  mysql-proxy-user: "proxyuser"
  mysql-proxy-password: "XTg1o9g8aP"
```

Create proxysql-watcher replication controller: `kubectl create -f proxysql-watcher.yaml`

Run tests (using [sysbench](https://launchpad.net/sysbench)):
```
sysbench --num-threads=4 --max-time=20 --test=oltp --db-ps-mode=disable
--mysql-user='proxyuser' --mysql-password='XTg1o9g8aP' --oltp-table-size=10000
--mysql-host=pxc-service --mysql-port=3306 prepare
```
```
sysbench --num-threads=4 --max-time=20 --test=oltp --db-ps-mode=disable
--mysql-user='proxyuser' --mysql-password='XTg1o9g8aP' --oltp-table-size=10000
--mysql-host=pxc-service --mysql-port=3306 run
```

## How it works?
Watcher monitors prefix in Etcd, where Percona nodes store their status. After
the node is created and started, watcher configures permissions for
`mysql-proxy-user` and adds new server to ProxySQL.

## How to test
I've included `monkey.py` which is able to randomly destroy pods. Use it on
your own risk!
