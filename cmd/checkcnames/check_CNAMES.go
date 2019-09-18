package main

import (
	"flag"
	"fmt"

	admindb "github.com/pmcclarence/dbtools/dbadmin"
	//_ "github.com/rapidloop/p"

	"log"
)

func main() {
	// Set up and parse command line flags
	//var InitCnamesTables = flag.Bool("InitCnamesTables", false, "Inits the CNAMES db tables - Default FALSE")
	//var dropcnamestables = flag.Bool("dropCnamesTables", false, "Drops the CNAMES db tables - Default FALSE")
	//var CreateManagementTable = flag.Bool("create-management-table", false, "Creates the management table needed for component management - Default FALSE")
	var cluster = flag.String("cluster", "%", "Cluster to use - i.e. tii, marks ... Only enter one at a time.")
	var env = flag.String("env", "%", "Environments to use - i.e. live, sprint, dev ... Only enter one at a time.")
	//var mapCluster= flag.Bool("map", false, "map a cluster - requires cluster and env parameters and will begin with the master cname")
	//var insertCNAME = flag.Bool("insertCname", false, "inserts a CNAME - requires a cluster flag - Default FALSE")
	//var cname = flag.String("cname", "", "Cluster to be used (used with other flags")
	//var cluster = flag.String("cluster", "", "Cluster to be used (used with other flags")

	flag.Parse()
	//fmt.Println(*initdb)

	// get a connection to the admin db

	var myCluster string
	myCluster = *cluster
	var myEnv string
	myEnv = *env

	admindbConn, _ := admindb.Connect_to_admin_db()
	err := admindbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	//	fmt.Println("Successfully Connected to: ", admindbname)

	fmt.Println("Finding CNAMES for Cluster: ", *cluster, " in Env: ", *env)
	admindb.GetCnamesForClusterEnv(admindbConn, myEnv, myCluster, true)

	/*	if *CreateManagementTable == true {
			admindb.CreateManagementTable(admindb_conn)
		} else {
			if *InitCnamesTables == true {
				admindb.InitCnamesTables(admindb_conn)
			} else {
				if *dropcnamestables == true {
					admindb.DropCnameTables(admindb_conn)
				} else {
					fmt.Println("Finding CNAMES for Cluster: ", *cluster, " in Env: ", *env)
					admindb.GetCnamesForClusterEnv(admindb_conn, my_env, my_cluster, true)
				}
			}
		}

	*/
}

//
/*
	if *insertCNAME == true {
		if (*cluster == "") {
		fmt.Println("insertCNAME requires both the cluster and CNAME flags.")
		} else if (*cname == "") {
		fmt.Println("insertCNAME requires both the cluster and CNAME flags.")
		} else {
			// Need to check if cluster name exists here before moving forward

			insertCname(admindb_conn, *cname, *cluster)
		}
*/

//	cname := "tii-master.sprint.iparadigms.com"
//	ip_addr,hostname  := get_details(cname)
//	fmt.Println(ip_addr, " ", hostname)
