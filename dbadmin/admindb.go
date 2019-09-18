package dbadmin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/rapidloop/pgmetrics"

	network_tools "github.com/pmcclarence/dbtools/network"
	_ "github.com/rapidloop/pq"
)

const (
	host = "ip-dbhub01.s2prod"
	port = 5432
	user = "postgres"
	//password = ""
	dbname = "dbtools"
)

type PsqlHost struct {
	HostID    int
	ParentID  int
	HostName  string
	IPAddr    string
	CName     string
	EnvID     int
	ClusterID int
	Stats     bool
	AccountID int
}

func CreateManagementTable(admindb_conn *sql.DB) {
	sqlCreateClusterTable := `CREATE TABLE components  (
		component_id serial primary key,
		component TEXT UNIQUE NOT NULL );`

	fmt.Println("Creating Management TABLE")
	_, err := admindb_conn.Exec(sqlCreateClusterTable)

	if err != nil {
		log.Fatal(err)
	}
}

func Connect_to_admin_db() (admindb_conn *sql.DB, admindbname string) {

	/*
		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	*/
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s sslmode=disable",
		host, port, user, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)

	}
	//defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Successfully Connected to: ", dbname)
	return db, dbname
}

func IsComponetInstalled(admindb_conn *sql.DB, component string) (component_installed bool) {

	sqlStatement := `SELECT component FROM components WHERE component like $1;`

	row := admindb_conn.QueryRow(sqlStatement, component)
	switch err := row.Scan(&component); err {
	case nil:
		component_installed = true
	case sql.ErrNoRows:
		component_installed = false
	default:
		panic(err)
	}
	fmt.Println(component_installed)
	return component_installed
}

func InitCnamesTables(admindb_conn *sql.DB) {
	//set up sql for tables
	//run sql
	component := "CNAMES"

	switch IsComponetInstalled(admindb_conn, "CNAMES") {
	case true:
		fmt.Println("Component - ", component, " Already installed. - Skipping")
	case false:

		sqlCreateClusterTable := `CREATE TABLE clusters  (
		ClusterID serial primary key,
		cluster TEXT UNIQUE NOT NULL );`

		sqlCreateEnvironmentTable := `CREATE TABLE environments (
		EnvID SERIAL PRIMARY KEY,
		env TEXT
		);`

		sqlCreateCnamesTable := `CREATE TABLE cnames (
		cname_id serial primary key,
		ClusterID INTEGER REFERENCES clusters(ClusterID),
		CName TEXT UNIQUE NOT NULL,
		cname_order int,
		EnvID INTEGER REFERENCES environments(EnvID),
		active BOOL NOT NULL DEFAULT true);`

		sqlCreateRunsTable := `CREATE TABLE runs (
		run_id SERIAL PRIMARY KEY,
		run_date TIMESTAMP NOT NULL DEFAULT NOW(),
		run_source TEXT);`

		sqlCreateCnameHistTable := `CREATE TABLE cname_history (
		history_id SERIAL PRIMARY KEY,
		run_id INTEGER REFERENCES runs(run_id),
		cname_id INTEGER REFERENCES cnames(cname_id),
		HostName TEXT,
		IPAddr TEXT);`

		fmt.Println("Creating TABLE clusters")
		_, err := admindb_conn.Exec(sqlCreateClusterTable)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Creating TABLE environments")
		_, err = admindb_conn.Exec(sqlCreateEnvironmentTable)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Creating TABLE cnames")
		_, err = admindb_conn.Exec(sqlCreateCnamesTable)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Creating TABLE runs")
		_, err = admindb_conn.Exec(sqlCreateRunsTable)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Creating TABLE cname_history")
		_, err = admindb_conn.Exec(sqlCreateCnameHistTable)

		if err != nil {
			log.Fatal(err)
		}

		sqlStatement := "INSERT into components (component) VALUES ($1) RETURNING component"

		err = admindb_conn.QueryRow(sqlStatement, component).Scan(&component)
		if err != nil {
			panic(err)
		}
		fmt.Println("Component - ", component, " Successfully set up")

	}
}

func DropCnameTables(admindb_conn *sql.DB) {

	sqlDropCnameHistrTable := "DROP TABLE cname_history;"
	sqlDropCnamesTable := "DROP TABLE cnames;"
	sqlDropClusterTable := "DROP TABLE clusters;"
	sqlDropEnvironmentsTable := "DROP TABLE environments;"
	sqlDropRunsTable := "DROP TABLE runs;"

	fmt.Println("DROPPING TABLES")

	fmt.Println("DROPPING TABLE cname_history")
	_, err := admindb_conn.Exec(sqlDropCnameHistrTable)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("DROPPING TABLE cnames")
	_, err = admindb_conn.Exec(sqlDropCnamesTable)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("DROPPING TABLE clusters")
	_, err = admindb_conn.Exec(sqlDropClusterTable)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("DROPPING TABLE runs")
	_, err = admindb_conn.Exec(sqlDropRunsTable)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("DROPPING TABLE environments")
	_, err = admindb_conn.Exec(sqlDropEnvironmentsTable)

	if err != nil {
		log.Fatal(err)
	}

	sqlStatement := "DELETE FROM components WHERE component = 'CNAMES'"

	_, err = admindb_conn.Exec(sqlStatement)
	if err != nil {
		panic(err)
	}
	fmt.Println("Component - CNMES removed")
}

func AddCluster(admindb_conn *sql.DB, clusterName string) {

	fmt.Println("Inserting Cluster ", clusterName)
	sqlStatement := "INSERT into clusters (cluster) VALUES ($1) RETURNING cluster"
	cluster := ""
	err := admindb_conn.QueryRow(sqlStatement, clusterName).Scan(&cluster)
	if err != nil {
		panic(err)
	}
	fmt.Println("Cluster - ", cluster, " Successfully inserted")

}

func InsertCname(admindb_conn *sql.DB, cname string, clusterName string) {

	// Get cluster id
	cluster_id := 0
	sqlStatement := `SELECT ClusterID FROM clusters WHERE cluster=$1;`
	row := admindb_conn.QueryRow(sqlStatement, clusterName)
	switch err := row.Scan(&cluster_id); err {
	case sql.ErrNoRows:
		fmt.Println("Cluster does not exist")
	case nil:
		fmt.Println(cluster_id)
	default:
		panic(err)
	}

	fmt.Println("Inserting CNAME ", cname, " for cluster ", clusterName)
	sqlStatement = "INSERT into cnames (ClusterID, CName) VALUES ($1, $2) RETURNING CName"
	err := admindb_conn.QueryRow(sqlStatement, cluster_id, cname).Scan(&cname)
	if err != nil {
		panic(err)
	}
	fmt.Println("CNAME - ", cname, " Successfully inserted")

}

func GetCnamesForClusterEnv(admindb_conn *sql.DB, env string, cluster string, allcnames bool, hostToCheck string) {

	// insert into runs and return id
	sqlStatement := "INSERT into runs (run_source) VALUES ($1)	RETURNING run_id"
	run_id := 0
	source := "getCnamesForClusterEnv"
	err := admindb_conn.QueryRow(sqlStatement, source).Scan(&run_id)
	if err != nil {
		panic(err)
	}

	// get all cnames in that cluster for the env
	type Cnames struct {
		cname_id   int
		cluster_id int
		cname      string
		env_id     int
	}
	var returned_cnames []Cnames
	//var env1 string
	//var cluster1 string

	//sqlStatement = "SELECT $1, $2"
	//err = admindb_conn.QueryRow(sqlStatement, env, cluster).Scan(&env1, &cluster1)

	//if err != nil {
	//	panic(err)
	//}

	sqlStatement = `SELECT cname_id, cluster_id, cname, env_id
		FROM  cnames
		WHERE env_id in (select env_id from environments where env ilike TRIM($1))
		AND cluster_id in (select cluster_id from clusters where cluster ilike TRIM($2))
		AND active is true
		ORDER BY cluster_id, env_id, cname_order asc;`

	rows, err := admindb_conn.Query(sqlStatement, env, cluster)

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {

		var cname_id int
		var cluster_id int
		var cname string
		var env_id int

		err = rows.Scan(&cname_id, &cluster_id, &cname, &env_id)

		if err != nil {
			log.Fatal(err)
		}

		// get HostName and ip address
		ip_addr, hostname := network_tools.Get_CNAME_details(cname)
		if hostToCheck == "1" {
			fmt.Println(cname, " - ", ip_addr, " - ", hostname)
		} else {
			if hostToCheck == hostname {
				fmt.Println(cname, " - ", ip_addr, " - ", hostname)
			}
		}

		// insert each one into the history table
		sqlInsertHistory := "INSERT INTO cname_history (run_id, cname_id, hostName, ip_address) values ($1, $2, $3, $4);"

		_, err = admindb_conn.Exec(sqlInsertHistory, run_id, cname_id, hostname, ip_addr)
		if err != nil {
			panic(err)
		}

		returned_cnames = append(returned_cnames, Cnames{cname_id: cname_id, cluster_id: cluster_id, cname: cname, env_id: env_id})
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

}

func Populate_node_details(admindb_conn *sql.DB, node_to_check PsqlHost) (ret_node PsqlHost) {
	fmt.Println("FUNCTION - Populate_node_details")
	fmt.Println("Node To Check: ", node_to_check)

	// if CName there use that to get ip address and host name
	if node_to_check.CName != "" {
		addr, err := net.LookupIP(node_to_check.CName)
		if err != nil {
			fmt.Println("Unknown addr")
			node_to_check.IPAddr = addr[0].String()
		} else {
			node_to_check.IPAddr = addr[0].String()
			fmt.Println("IPAddr: ", node_to_check.IPAddr)
		}
	}

	// if ip address there get host name and put it into HostName
	host, err := net.LookupAddr(node_to_check.IPAddr)
	if err != nil {
		fmt.Println("Unknown host")
		node_to_check.HostName = "Unknown host"
	} else {
		node_to_check.HostName = strings.TrimRight(host[0], ".")
		fmt.Println("Hostname: ", node_to_check.HostName)
	}

	// select on HostName from db and get HostID

	// Select a row and return specific columns from that row
	// Capture no rows returned and print a message
	// Otherwise print values from the row
	//
	sqlStatement := `SELECT node_id, hostName, ip_address FROM node where hostName ilike TRIM($1);`

	var node_id int
	var hostname string
	var ip_address string
	if node_to_check.HostName != "Unknown host" {
		row := admindb_conn.QueryRow(sqlStatement, node_to_check.HostName)
		switch err := row.Scan(&node_id, &hostname, &ip_address); err {
		case sql.ErrNoRows:
			fmt.Println("No rows were returned!")
			// Row does not exist - Insert the row
			sqlInsert := `Insert into node (hostName, ip_address, parent_id, env_id, cluster_id, last_checked) values ($1, $2, $3, $4, $5, now()) RETURNING node_id;`
			err = admindb_conn.QueryRow(sqlInsert, node_to_check.HostName, node_to_check.IPAddr, node_to_check.ParentID, node_to_check.EnvID, node_to_check.ClusterID).Scan(&node_id)
			if err != nil {
				panic(err)
			}
			node_to_check.HostID = node_id
		case nil:
			fmt.Println("no errors and row returned - update the record")
			// no errors and row returned - update the record
			sqlUpdate := `Update node set ip_address = $2 , parent_id = $3, env_id = $4, cluster_id = $5, last_checked = now() where node_id = $1 RETURNING node_id;`
			err := admindb_conn.QueryRow(sqlUpdate, node_id, node_to_check.IPAddr, node_to_check.ParentID, node_to_check.EnvID, node_to_check.ClusterID).Scan(&node_id)
			if err != nil {
				fmt.Println("Error from the Update.")
				panic(err)
			}
			fmt.Println("returned node : ", node_id)
			node_to_check.HostID = node_id
		default:
			panic(err)
		}
	}
	fmt.Println(node_to_check)
	//fmt.Println("LEAVING Populate_node_details")
	return node_to_check
}

func Get_replicated_hosts(admindb_conn *sql.DB, host_details PsqlHost) (returned_hosts []PsqlHost) {
	fmt.Println("Function get_replicated_hosts for : ", host_details)
	// loop through []hosts to check
	// for each
	// connect to db on host passed in
	// retrieve replicated hosts
	// write them to []returned_hosts
	//

	// Open a db connection
	//psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	//	"password=%s sslmode=disable",
	//	host_details.HostName, port, user, password)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"sslmode=disable",
		host_details.HostName, port, user)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Println("Errored in the PING")
		//log.Fatal(err)
	} else {
		//fmt.Println("Successfully Connected to: ", dbname)

		/////////////

		sqlStatement := `select client_addr from pg_stat_replication;`
		var current_host PsqlHost
		fmt.Println("Current Host: ", current_host)
		rows, err := db.Query(sqlStatement)
		if err != nil {
			// handle this error better than this
		}
		defer rows.Close()
		for rows.Next() {
			//fmt.Println("***********")
			current_host.HostName = ""
			//fmt.Println("Current Host: ", current_host)
			//f/mt.Println("current_host.IPAddr: ", current_host.IPAddr)
			err = rows.Scan(&current_host.IPAddr)
			//fmt.Println("current_host.IPAddr: ", current_host.IPAddr)
			if err != nil {
				// handle this error
				panic(err)
			}
			//fmt.Println(current_host.IPAddr)
			current_host.ParentID = host_details.HostID
			current_host.EnvID = host_details.EnvID
			current_host.ClusterID = host_details.ClusterID

			//fmt.Println("Current Host : ", current_host)
			current_host = Populate_node_details(admindb_conn, current_host)

			// call get HostName and put in current_host

			hostname, err := net.LookupAddr(current_host.IPAddr)

			if err != nil {
				fmt.Println("Unknown host")
				current_host.HostName = "NotFound"
			} else {

				current_host.HostName = strings.TrimRight(hostname[0], ".")
				//fmt.Println(HostName)

				// Insert ParentID
				fmt.Println("Parent ID: ", host_details.HostID)
				current_host.ParentID = host_details.HostID

				//append to returned hosts only if that host can be reached
				returned_hosts = append(returned_hosts, current_host)
			}
			//fmt.Println(current_host.HostName)

		}
		// get any error encountered during iteration
		err = rows.Err()
		if err != nil {
			panic(err)
		}
		//fmt.Println(returned_hosts)

	}

	return returned_hosts
}

func Get_env_and_cluster_ids(admindb_conn *sql.DB, env string, cluster string) (env_id int, cluster_id int) {
	fmt.Println("FUNCTION: Get_env_and_cluster_ids")
	fmt.Println("Cluster: ", cluster, " in Env: ", env)

	sqlStatement_env := `select env_id from environments where env ilike TRIM($1)`
	sqlstatement_cluster := `select cluster_id from clusters where cluster ilike TRIM($1)`

	row := admindb_conn.QueryRow(sqlStatement_env, env)
	switch err := row.Scan(&env_id); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
	case nil:
		//fmt.Println(EnvID)
	default:
		panic(err)
	}

	row = admindb_conn.QueryRow(sqlstatement_cluster, cluster)
	switch err := row.Scan(&cluster_id); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
	case nil:
		//fmt.Println(ClusterID)
	default:
		panic(err)
	}
	//fmt.Println("Returning Cluster_id: ", cluster_id, "  Env_id: ", env_id)
	return env_id, cluster_id
}

func ClusterMap(admindb_conn *sql.DB, env string, cluster string) {
	fmt.Println("MAPPING CLUSTER .....")

	/*
	   type PsqlHost struct {
	   	HostID int
	   	ParentID int
	   	HostName string
	   	IPAddr string
	   	CName string
	   	EnvID int
	   	ClusterID int
	   }*/

	master := PsqlHost{}
	//	fmt.Println("master: ",master)
	hosts_to_check := make([]PsqlHost, 0)
	checked_hosts := make([]PsqlHost, 0)

	env_id, cluster_id := Get_env_and_cluster_ids(admindb_conn, env, cluster)

	// Get master host name - return
	// TODO - can update this to use the ids instead of the env and cluster strings now.
	sqlStatement := `SELECT CName
		FROM  cnames
		WHERE env_id in (select env_id from environments where env ilike TRIM($1))
		AND cluster_id in (select cluster_id from clusters where cluster ilike TRIM($2))
		AND cname_order = 1
		LIMIT 1;`

	row := admindb_conn.QueryRow(sqlStatement, env, cluster)
	switch err := row.Scan(&master.CName); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
	case nil:
		fmt.Println(master.CName)
	default:
		panic(err)
	}
	master.EnvID = env_id
	master.ClusterID = cluster_id

	//fmt.Println("Master Before: ", master)
	master = Populate_node_details(admindb_conn, master)
	//fmt.Println("Master After: ", master)
	// Checking

	fmt.Println("MASTER: ", master)
	//for i, s := range hosts_to_check {
	//	i = i
	//	fmt.Println("hosts_to_check: ", s)
	//}
	//fmt.Println("checked_hosts: ",checked_hosts)
	//

	//store in struct master host

	master.IPAddr, master.HostName = network_tools.Get_CNAME_details(master.CName)
	//fmt.Println("master Hostname: ",master.HostName, " - Master ip Address: ", master.IPAddr)

	hosts_to_check = append(hosts_to_check, master)

	// Start with master - get replicated hots - add them to hosts to check - go through them adding additional hosts until there are none levt
	for hosts_remaining := true; hosts_remaining == true; {

		returned_hosts := Get_replicated_hosts(admindb_conn, hosts_to_check[0])
		checked_hosts = append(checked_hosts, hosts_to_check[0])
		hosts_to_check = append(hosts_to_check, returned_hosts...)

		hosts_to_check = hosts_to_check[1:]
		fmt.Println("Hosts To Check: ", hosts_to_check)

		if len(hosts_to_check) < 1 {
			hosts_remaining = false
		}

	}
	//fmt.Println("AFTER FOR")

	//returned_hosts := Get_replicated_hosts(master)

	// insert into database host or update host record with parent

}

func Get_run_id(source string) (run_id int) {

	admindb_conn, _ := Connect_to_admin_db()

	err := admindb_conn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Successfully Connected to: ", admindbname)

	sqlStatement := "INSERT into runs (run_source) VALUES ($1)	RETURNING run_id"
	run_id = 0
	//source = "getCnamesForClusterEnv"
	err = admindb_conn.QueryRow(sqlStatement, source).Scan(&run_id)
	if err != nil {

		fmt.Println("Error in Get_run_id")
		panic(err)
	}

	admindb_conn.Close()

	return run_id

}

func GetHostList(env string, cluster string, hostname string, account_id int) (returned_hosts []PsqlHost) {

	//fmt.Println("hostname is: ", hostname)
	admindb_conn, admindbname := Connect_to_admin_db()

	err := admindb_conn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully Connected to: ", admindbname)
	var sqlStatement string

	if hostname != "none" {

		sqlStatement = `SELECT node_id, hostname, ip_address, collect_stats, account_id
			FROM  node
			WHERE hostname ilike TRIM($1)
			AND account_id = $2;`

		var current_host PsqlHost

		row := admindb_conn.QueryRow(sqlStatement, hostname, account_id)
		switch err := row.Scan(&current_host.HostID, &current_host.HostName, &current_host.IPAddr, &current_host.Stats, &current_host.AccountID); err {
		case sql.ErrNoRows:
			fmt.Println("No matching hosts were returned!")
		case nil:
			returned_hosts = append(returned_hosts, current_host)
		default:
			panic(err)
		}
	} else {
		sqlStatement = `SELECT node_id, hostname, ip_address, collect_stats, account_id
		 FROM  node
		 WHERE env_id in (select env_id from environments where env ilike TRIM($1))
		 AND cluster_id in (select cluster_id from clusters where cluster ilike TRIM($2));`

		var current_host PsqlHost

		//	rows, err := admindb_conn.Query(sqlStatement, env, cluster)
		//if err != nil {
		//panic(err)// handle this error better than this
		//}
		rows, err := admindb_conn.Query(sqlStatement, env, cluster)
		if err != nil {
			panic(err)
			// handle this error better than this
		}
		defer rows.Close()
		for rows.Next() {
			err = rows.Scan(&current_host.HostID, &current_host.HostName, &current_host.IPAddr, &current_host.Stats, &current_host.AccountID)
			if err != nil {
				// handle this error
				panic(err)
			}
			returned_hosts = append(returned_hosts, current_host)
		}
		// get any error encountered during iteration
		err = rows.Err()
		if err != nil {
			panic(err)
		}
	}

	admindb_conn.Close()

	//fmt.Println(returned_hosts)
	fmt.Println(returned_hosts)
	return returned_hosts

}

func InsertStatsResult(run_id int, host PsqlHost, stats pgmetrics.Model) {

	fmt.Println("Run_id: ", run_id)
	fmt.Println("host: ", host)
	fmt.Println("Model: ", stats.Databases)

	file, _ := json.MarshalIndent(stats, "", " ")

	admindb_conn, _ := Connect_to_admin_db()

	sqlInsertStats := "INSERT INTO stats (run_id, node_id, stats) values ($1, $2, $3);"

	_, err := admindb_conn.Exec(sqlInsertStats, run_id, host.HostID, file)
	if err != nil {
		panic(err)
	}

}
