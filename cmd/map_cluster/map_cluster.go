package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/pmcclarence/dbtools/dbadmin"
	_ "github.com/rapidloop/pq"
)

func main() {

	var cluster = flag.String("cluster", "%", "Cluster to use - i.e. tii, marks ... Only enter one at a time.")
	var env = flag.String("env", "%", "Environments to use - i.e. live, sprint, dev ... Only enter one at a time.")
	//var mapCluster= flag.Bool("map", false, "map a cluster - requires cluster and env parameters and will begin with the master cname")

	admindb_conn, _ := dbadmin.Connect_to_admin_db()
	err := admindb_conn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Successfully Connected to: ", admindbname)

	flag.Parse()

	fmt.Println("Mapping  Cluster: ", *cluster, " in Env: ", *env)
	dbadmin.ClusterMap(admindb_conn, *env, *cluster)
	/*
		sqlStatement := `SELECT cname
			FROM  cnames
			WHERE env_id in (select env_id from environments where env ilike TRIM($1))
			AND cluster_id in (select cluster_id from clusters where cluster ilike TRIM($2))
			AND cname_order = 1
			LIMIT 1;`

		currHost := &dbadmin.PsqlHost{}
	*/
}
