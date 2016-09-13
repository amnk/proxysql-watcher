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

//Format mysql command for later use
func formatMySQL(username string, password string, hostname string, port int, sql string) []string {
	user_template := fmt.Sprintf("-u%s", username)
	pass_template := fmt.Sprintf("-p%s", password)
	host_template := fmt.Sprintf("-h%s", hostname)
	port_template := fmt.Sprintf("-P%d", port)

	mysql := []string{user_template,
		pass_template,
		host_template,
		port_template,
		"-e",
		sql}
	return mysql
}

//Run command and return error if any
func runCommand(command string, opts []string) (string, error) {
	run := exec.Command(command, opts...)
	output, err := run.CombinedOutput()
	return string(output), err
}

//Helper function to run mysql
func runMySQL(username string, password string, hostname string, port int, sql string) (string, error) {
	command := "mysql"
	opts := formatMySQL(username, password, hostname, port, sql)
	output, err := runCommand(command, opts)
	return output, err
}

func addUser(username string, password string) string {
	add_user_tmpl := "REPLACE INTO mysql_users (username, password, active, default_hostgroup, max_connections) VALUES ('%s', '%s', 1, 0, 200);"
	return fmt.Sprintf(add_user_tmpl, username, password)
}

func addServer(address string) string {
	add_server_tmpl := "REPLACE INTO mysql_servers (hostgroup_id, hostname, port, max_replication_lag) VALUES (0, '%s', 3306, 20);"
	return fmt.Sprintf(add_server_tmpl, address)
}

func grantRights(username string, hostname string, password string) string {
	grant_rights_tmpl := "GRANT ALL ON *.* TO '%s'@'%s' IDENTIFIED BY '%s'"
	return fmt.Sprintf(grant_rights_tmpl, username, hostname, password)
}

func reloadConfig() string {
	reload_tmpl := "LOAD MYSQL SERVERS TO RUNTIME; SAVE MYSQL SERVERS TO DISK; LOAD MYSQL USERS TO RUNTIME; SAVE MYSQL USERS TO DISK;"
	return reload_tmpl
}

func main() {
	var prefix = flag.String("prefix", "/pxc-cluster/k8scluster2", "Etcd prefix to listen")
	var mysql_user = flag.String("mysql_user", "user", "Mysql proxy user")
	var mysql_pass = flag.String("mysql_pass", "pass", "Mysql proxy password")
	var proxy_address = flag.String("proxy_address", "127.0.0.18", "Host IP address of ProxySQL instance")
	var root_pass = flag.String("root_pass", "rootpass", "Mysql root password")
	var etcd_service = flag.String("etcd_service", "etcd-client:2379", "Etcd hostname:port")
	var proxy_user = flag.String("proxy_user", "admin", "ProxySQL backend admin username")
	var proxy_pass = flag.String("proxy_pass", "admin", "ProxySQL backend admin password")
	flag.Parse()

	conn := fmt.Sprintf("http://%s", *etcd_service)
	c, _ := NewEtcdClient([]string{conn})
	//nodes := make(map[string]string)

	//watcher := c.client.Watcher(*prefix, &client.WatcherOptions{AfterIndex: uint64(0), Recursive: true})
	for {
		//We populate ProxySQL watcher database in a loop to prevent
		//loosing state
		output1, err1 := runMySQL(*proxy_user, *proxy_pass, *proxy_address, 6032, addUser(*mysql_user, *mysql_pass))
		output2, err2 := runMySQL(*proxy_user, *proxy_pass, *proxy_address, 6032, addUser("root", *root_pass))
		output3, err3 := runMySQL(*proxy_user, *proxy_pass, *proxy_address, 6032, reloadConfig())

		if err1 != nil {
			log.Printf(string(output1))
		}
		if err2 != nil {
			log.Printf(string(output2))
		}
		if err3 != nil {
			log.Printf(string(output3))
		}
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
			//_, ok := nodes[key]
			ok := false
			if !ok {
				new_node := make(map[string]string)
				// Apparently recursive watch does not populate
				// Nodes structure, so we have to make another
				// call to get ipaddr
				r, _ := c.client.Get(context.Background(), key, &client.GetOptions{Recursive: true})
				nodeParse(r.Node, new_node, "/ipaddr")

				//First we grant rights to proxy user
				output1, err1 := runMySQL("root", *root_pass, new_node[key], 3306,
					grantRights(*mysql_user, "%", *mysql_pass))
				//Then we add server to proxysql backend
				output2, err2 := runMySQL(*proxy_user, *proxy_pass, *proxy_address, 6032,
					addServer(new_node[key]))
				if err1 == nil && err2 == nil {
					log.Printf("New node %s, adding to the cluster", key)
				} else {
					if err1 != nil {
						log.Printf(output1)
					}
					if err2 != nil {
						log.Printf(output2)
					}
				}
			}

		}
	}
}
