package main

import (
	"flag"
	"fmt"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"log"
	"os/exec"
	"strings"
	"time"
)

type Client struct {
	client client.KeysAPI
}

// NewEtcdClient returns an *etcd.Client with a connection to named machines.
func NewEtcdClient(machines []string) (*Client, error) {
	var c client.Client
	var kapi client.KeysAPI
	var err error

	cfg := client.Config{
		Endpoints:               machines,
		HeaderTimeoutPerRequest: time.Duration(3) * time.Second,
	}

	c, err = client.New(cfg)
	if err != nil {
		return &Client{kapi}, err
	}

	kapi = client.NewKeysAPI(c)
	return &Client{kapi}, nil
}

// Parse etcd Node and populate vars if key has needed suffix
func nodeParse(node *client.Node, vars map[string]string, suffix string) error {
	if node != nil {
		key := node.Key
		if !node.Dir {
			if strings.HasSuffix(key, suffix) {
				k := strings.TrimSuffix(key, suffix)
				vars[k] = node.Value
			}
		} else {
			for _, node := range node.Nodes {
				nodeParse(node, vars, suffix)
			}
		}
	}
	return nil
}
func main() {
	var prefix = flag.String("prefix", "/pxc-cluster/k8scluster2", "Etcd prefix to listen")
	var proxy_user = flag.String("proxy_user", "user", "Mysql proxy user")
	var proxy_pass = flag.String("proxy_pass", "pass", "Mysql proxy password")
	var proxy_address = flag.String("proxy_address", "127.0.0.18", "Host IP address of ProxySQL instance")
	var root_pass = flag.String("root_pass", "rootpass", "Mysql root password")
	var etcd_service = flag.String("etcd_service", "etcd-client:2379", "Etcd hostname:port")
	flag.Parse()

	conn := fmt.Sprintf("http://%s", *etcd_service)
	c, _ := NewEtcdClient([]string{conn})
	nodes := make(map[string]string)

	mysql3 := []string{"-uadmin",
		"-padmin",
		fmt.Sprintf("-h%s", *proxy_address),
		"-P6032",
		"-e",
		fmt.Sprintf("REPLACE INTO mysql_users (username, password, active, default_hostgroup, max_connections) VALUES ('%s', '%s', 1, 0, 200);", *proxy_user, *proxy_pass)}

	mysql4 := []string{"-uadmin",
		"-padmin",
		fmt.Sprintf("-h%s", *proxy_address),
		"-P6032",
		"-e",
		"LOAD MYSQL SERVERS TO RUNTIME; SAVE MYSQL SERVERS TO DISK; LOAD MYSQL USERS TO RUNTIME; SAVE MYSQL USERS TO DISK;"}

	log.Printf("Populating mysql_users table")
	mysql3_r := exec.Command("mysql", mysql3...)
	output, err3 := mysql3_r.CombinedOutput()
	if err3 != nil {
		log.Fatal(mysql3_r, string(output))
	}

        log.Printf("Populating values to runtime")
	mysql4_r := exec.Command("mysql", mysql4...)
	output, err4 := mysql4_r.CombinedOutput()
	if err4 != nil {
		log.Fatal(mysql4_r, string(output))
	}

	//watcher := c.client.Watcher(*prefix, &client.WatcherOptions{AfterIndex: uint64(0), Recursive: true})
	for {
		// endless loop that watches for new nodes in "prefix" and
		// populates nodes map
		// TODO: implement delete action
		watcher := c.client.Watcher(*prefix, &client.WatcherOptions{AfterIndex: uint64(0), Recursive: true})
		ctx, _ := context.WithCancel(context.Background())
		resp, err := watcher.Next(ctx)
		if err != nil {
			log.Printf("Error occured: %s", err)
		}

                // Delete event is tricky, because it is either TTL or actual
                // node poweroff. In the first case we should do nothing,
                // in the second - delete node from ProxySQL. But we rely on
                // ProxySQL monitoring to do this now.
                if resp.Action == "delete" {
                        continue
                }
		if resp.Node.Dir {
			key := resp.Node.Key
			_, ok := nodes[key]
			if !ok {
				new_node := make(map[string]string)
				// Apparently recursive watch does not populate
				// Nodes structure, so we have to make another
				// call to get ipaddr
				r, _ := c.client.Get(context.Background(), key, &client.GetOptions{Recursive: true})
				nodeParse(r.Node, new_node, "/ipaddr")
				mysql1 := []string{"-uroot",
					fmt.Sprintf("-p%s", *root_pass),
					fmt.Sprintf("-h%s", new_node[key]),
					"-e",
					fmt.Sprintf("GRANT ALL ON *.* TO '%s'@'%s' IDENTIFIED BY '%s'", *proxy_user, "%", *proxy_pass)}

				mysql2 := []string{"-uadmin",
					"-padmin",
					fmt.Sprintf("-h%s", *proxy_address),
					"-P6032",
					"-e",
					fmt.Sprintf("REPLACE INTO mysql_servers (hostgroup_id, hostname, port, max_replication_lag) VALUES (0, '%s', 3306, 20);", new_node[key])}

				mysql1_r := exec.Command("mysql", mysql1...)
				output, err1 := mysql1_r.CombinedOutput()
				if err1 != nil {
					//new node takes some time to boot, so
					//if connection failes we should not
					//exit
                                        log.Printf(string(output))
					continue

				}

				mysql2_r := exec.Command("mysql", mysql2...)
				output, err2 := mysql2_r.CombinedOutput()
				if err2 != nil {
					log.Printf(string(output))
				}

                                if err1 == nil && err2 == nil {
                                        log.Printf("New node %s, adding to the cluster", key)
                                        nodes[key] = new_node[key]
                                }
			}

		}
	}
}
