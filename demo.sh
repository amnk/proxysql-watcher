#!/bin/bash

k8s=`which kubectl`
SLEEP="sleep 5"
k8s_describe="$k8s get pods" 
sysbench='./sysbench.sh'
files=(
       "ccp/etcd.yaml"
       "ccp/percona.yaml"
       "ccp/proxysql-watcher.yaml"
      )

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
     "$k8s create -f ${files[0]} && $k8s_describe"
     "$k8s create -f ${files[1]} && $k8s_describe"
     "$k8s create -f ${files[2]} && $k8s_describe && $SLEEP"
     "$sysbench prepare && $sysbench run"
     "$k8s scale --replicas 9 rc/pxc-rc && $k8s_describe"
     "$sysbench run"
     "$k8s scale --replicas 5 rc/pxc-rc && $k8s_describe"
     "$sysbench run"
     "cat README.demo"
    )

msg=(
     "Small demo of Percona cluster with proxysql on the clean k8s env:"
     "First we will create etcd cluster for Percona service discovery:"
     "Then we create percona replication controller, proxysql replication controller and their services:"
     "And finally we add proxysql-watcher:"
     "Now we can run sysbench testsuit against pxc-service, which is an entrypoint for all MySQL connections:"
     "We can scale replication controller to bigger number of nodes:"
     "And run the test again:"
     "Down scale the replication controller:"
     "And run the test again:"
     "Thank you!"
    )

for i in `seq 0 ${#arr[@]}`
do
  echo ${msg[$i]}
  $SLEEP
  eval "${arr[$i]}"
  $SLEEP
  echo "======================="
done
