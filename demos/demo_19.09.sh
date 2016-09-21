#!/bin/bash

#Virtual env for ccp
. fuel-ccp/.venv/bin/activate
mysql_user="proxyuser"
mysql_pass="XTg1o9g8aP"
mysql_host="mariadb"
k8s=`which kubectl`
SLEEP="sleep 5"
k8s_describe="$k8s get pods"
sysbench='./sysbench.sh'
mysql_status="show global status like 'wsrep_cluster_%'"
get_mysql_status="mysql -u$mysql_user -p$mysql_pass -h$mysql_host -P3306 -e \"$mysql_status\""
get_pods="kubectl get pods -l app=mariadb --namespace=ccp -o json | jq '.items[] | .metadata.name'"
function cleanup {
    echo "Cleanup all."
    for item in "${files[@]}"
    do
      $k8s delete -f $item
    done
    $k8s_describe
}

if [ "$1" == "cleanup" ]; then
    cleanup
    exit 0
fi

arr=(
     "$k8s_describe"
     "docker images --format '{{.ID}}: {{.Repository}}'"
     "./deploy.sh"
     "sleep 35"
     "$get_mysql_status"
     "$sysbench prepare && $sysbench run"
     "$k8s scale --replicas=3 deployments/mariadb-percona --namespace=ccp"
     "$sysbench run"
     "echo '->>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>'"
     "$sysbench run"
     "$k8s scale --replicas=7 deployments/mariadb-percona --namespace=ccp"
     "$sysbench run"
     ""
     "$get_mysql_status"
     "cat README.demo"
    )

msg=(
     "Small demo of Percona cluster with proxysql on the clean k8s env:"
     "Images are pre-built to save time:"
     "First, we deploy Percona with ProxySQL using CCP framework:"
     "We need to allow some time for Percona to create cluster"
     "By default, Percona cluster is started with 1 node:"
     "Now we can run sysbench testsuit against mariadb, which is an entrypoint for all MySQL connections:"
     "Let's then scale our deployment to 3 nodes:"
     "And run the test again:"
     "Destructive test is monkey, which kills one of the Percona pods each 25 seconds (it will be running in background)"
     "Let's run the test again:"
     "We can scale deployment to 7 nodes:"
     "And repeat the test:"
     "'ignored_errors' means that test had to reconnect to the cluster when pod is deleted"
     "But Percona cluster remains operational:"
     "Thank you!"
    )

for i in `seq 0 ${#arr[@]}`
#for i in `seq 0 6`
do
  echo ${msg[$i]}
  $SLEEP
  eval "${arr[$i]}"
  $SLEEP
  echo "======================="
done
