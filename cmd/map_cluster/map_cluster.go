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

	admindbConn, _ := dbadmin.Connect_to_admin_db()
	err := admindbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Successfully Connected to: ", admindbname)

	flag.Parse()

	fmt.Println("Mapping  Cluster: ", *cluster, " in Env: ", *env)
	dbadmin.ClusterMap(admindbConn, *env, *cluster)
	dbadmin.PrintMappedCluster(admindbConn, *env, *cluster)
	admindbConn.Close()

}
