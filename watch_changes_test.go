package main

import (
	"fmt"
	"strings"
	"testing"
)

func toString(s []string) string {
	return strings.Join(s, " ")
}

func TestFormatMySQL(t *testing.T) {
	opts := "-uadmin -padmin -h127.0.0.1 -P6032 -e %s"
	tmpl := "REPLACE INTO mysql_users (username, password, active, default_hostgroup, max_connections) VALUES ('%s', '%s', 1, 0, 200);"
	test := fmt.Sprintf(tmpl, "test", "test")
	mysql := fmt.Sprintf(opts, test)
	output := toString(formatMySQL("admin", "admin", "127.0.0.1", 6032, addUser("test", "test")))
	if output != mysql {
		t.Error("Update users failed, want ", output, "got ", mysql)
	}
}
