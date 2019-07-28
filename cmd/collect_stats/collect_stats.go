package main

import (
	"flag"
	"fmt"

	"github.com/pmcclarence/dbtools/getmetrics"

	"github.com/pmcclarence/dbtools/dbadmin"
	"github.com/rapidloop/pgmetrics/collector"

)



func main() {

	var cluster = flag.String("cluster", "%", "Cluster to use - i.e. tii, marks ... Only enter one at a time.")
	var env = flag.String("env", "%", "Environments to use - i.e. live, sprint, dev ... Only enter one at a time.")
	var host = flag.String("host", "none", "Specific host to collect stats on.")
 	var account_id = flag.Int("account_id", 1, "Account ID to work with")

	flag.Parse()

	cc := collector.DefaultCollectConfig()
	cc.Host = "127.0.0.1"
	cc.User = "postgres"

	fmt.Println("Env: ", *env, "   Cluster: ", *cluster)

	// Get a run Id
	run_id := dbadmin.Get_run_id("collect_stats")
	fmt.Println("Run_id: ", run_id)

	// get the list of hosts to check
	//TODO - Loop through them, get the stats and store the results

	host_list := dbadmin.GetHostList(*env, *cluster, *host, *account_id)
	for _, h := range host_list {
		cc.Host = h.HostName
		if h.Stats == true {
			fmt.Println(cc.Host, " ", h.HostID)
			m := collector.Collect(cc, []string{""})

			//dbadmin.InsertStatsResult(run_id, h, *m)

			getmetrics.ClusterStats(run_id, h, m)

			if m.System != nil {
				getmetrics.GetSystemMetrics(run_id, h, m)
			}

			if m.IsInRecovery {
				getmetrics.GetRecovery(run_id, h, m)
			}

			if m.ReplicationIncoming != nil {
				getmetrics.GetIncomingReplication(run_id, h, m)
			}

			if len(m.ReplicationOutgoing) > 0 {
				getmetrics.GetOutGoingReplicationDetails(run_id, h, m)
			}

			if len(m.ReplicationSlots) > 0 {
				getmetrics.GetReplicationSlots(run_id, h, m)
			}

			getmetrics.GetWALDetails(run_id, h, m)

		}

	}
}
