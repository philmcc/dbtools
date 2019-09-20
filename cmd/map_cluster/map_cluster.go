package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/pmcclarence/dbtools/dbadmin"
	_ "github.com/rapidloop/pq"
)

type node struct {
	cname      sql.NullString
	hostname   sql.NullString
	nodeID     int
	parentID   int
	parentPath string
}

func (n node) String() string {
	return fmt.Sprintf("%s - %s", n.hostname.String, n.cname.String)
}

func main() {

	var cluster = flag.String("cluster", "%", "Cluster to use - i.e. tii, marks ... Only enter one at a time.")
	var env = flag.String("env", "%", "Environments to use - i.e. live, sprint, dev ... Only enter one at a time.")
	//var mapCluster= flag.Bool("map", false, "map a cluster - requires cluster and env parameters and will begin with the master cname")

	admindbConn, _ := dbadmin.Connect_to_admin_db()
	err := admindbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Successfully Connected to: ", admindbname)

	flag.Parse()

	fmt.Println("Mapping  Cluster: ", *cluster, " in Env: ", *env)
	dbadmin.ClusterMap(admindbConn, *env, *cluster)

	sqlStatement := `SELECT t.cname ,t.hostname, t.node_id, t.parent_id, t.parent_path
											FROM public.mapped_cluster t
											WHERE env ilike TRIM($1)
    										AND cluster ilike TRIM($2)
    										and last_checked = (select max(last_checked) FROM public.mapped_cluster t
																						WHERE env  ilike TRIM($1)
    																				AND cluster  ilike TRIM($2)
																					 )
    									;`

	var anode node
	var padding string
	var level int
	rows, err := admindbConn.Query(sqlStatement, env, cluster)
	if err != nil {
		panic(err)
		// handle this error better than this
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&anode.cname, &anode.hostname, &anode.nodeID, &anode.parentID, &anode.parentPath)
		if err != nil {
			// handle this error
			panic(err)
		}
		padding = ""
		level = strings.Count(anode.parentPath, ".")

		for i := 1; i <= level; i++ {
			padding = padding + " --- "
		}

		fmt.Println(padding, anode)
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	admindbConn.Close()
}
