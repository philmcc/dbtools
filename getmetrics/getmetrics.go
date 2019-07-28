package getmetrics

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pmcclarence/dbtools/dbadmin"
	"github.com/rapidloop/pgmetrics"
)

//TODO - Table - physicalreplicationslots
//TODO - Table - logicalreplicationslots
//TODO - Table - wal_files
//TODO - Table - roles
//TODO - Table - backends
//TODO - Table - vacuum
//TODO - Table - tablespaces
//TODO - Table - databases
//TODO - Table - extensions
//TODO - Table - functions
//TODO - Table - slowqueries
//TODO - Table - sequences
//TODO - Table - disabledtriggers
//TODO - Table - logicalreplicationpublications
//TODO - Table - logicalreplicationsubscriptions
//TODO - Table -  indexes
//TODO - Table - tables

func ClusterStats(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {

	version := getVersion(stats)
	bytesSincePrior, _ := lsnDiff(stats.RedoLSN, stats.PriorLSN)
	bytesSinceRedo, _ := lsnDiff(stats.CheckpointLSN, stats.RedoLSN)
	priorLSN := stats.PriorLSN
	//human_since_Prior := humanize.IBytes(uint64(bytesSincePrior))
	checkpointLSN := stats.CheckpointLSN
	redoLSN := stats.RedoLSN
	//human_since_redo := humanize.IBytes(uint64(bytesSinceRedo))
	//report_start_time := fmtTime(stats.Metadata.At)
	cluster_name := getSetting(stats, "cluster_name")
	server_version := getSetting(stats, "server_version")
	server_start_time := fmtTime(stats.StartTime)
	since_server_start_time := fmtTimeAndSince(stats.StartTime)

	// Transaction IDs
	//stats.OldestXid, stats.NextXid-1,
	transID_diff := stats.NextXid - 1 - stats.OldestXid

	lastTransaction := stats.LastXactTimestamp //TODO - Transaction is not being returned

	notificationQueue := stats.NotificationQueueUsage
	activeBackends := len(stats.Backends)
	recoveryMode := stats.IsInRecovery

	/* TODO  - Put system info in

	 */

	/*fmt.Println("Server Version Num: ", version)
	fmt.Println("Prior LSN: ", priorLSN)
	fmt.Println("since Prior: ", bytesSincePrior)
	fmt.Println("HUMAN since Prior: ", human_since_Prior)
	fmt.Println("redo LNS: ", redoLSN)
	fmt.Println("Since Redo: ", bytesSinceRedo)
	fmt.Println("HUMAN Since Redo: ", human_since_redo)
	fmt.Println("Checkpoint LSN: ", checkpointLSN)
	fmt.Println("Report Start Time: ", report_start_time)
	fmt.Println("Cluster Name: ", cluster_name) //TODO - Clustername is not being returned (is it even set?)
	fmt.Println("Server Version: ", server_version)
	fmt.Println("Server Start Time: ", server_start_time)
	fmt.Println("Server Start Time: ", since_server_start_time)
	fmt.Println("Oldest Transaction ID: ", stats.OldestXid)
	fmt.Println("Newest Transaction ID: ", stats.NextXid-1)
	fmt.Println("transaction ID Diff: ", transID_diff)
	fmt.Println("Last Transaction: ", lastTransaction)
	fmt.Println("Notification Queue % Used: ", notificationQueue)
	fmt.Println("Active Backends: ", activeBackends)
	fmt.Println("Recovery Mode: ", recoveryMode)
	fmt.Println("")
	*/
	admindb_conn, admindbname := dbadmin.Connect_to_admin_db()

	err := admindb_conn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully Connected to: ", admindbname)

	sqlStatement := ` Insert into dbcluster (  run_id, node_id,
                               version,     
                               bytesSincePrior,
                               bytesSinceRedo, 
                               priorLSN, checkpointLSN,  
                               redoLSN, 
                        	   cluster_name,   
                               server_version,
                               server_start_time,
                               since_server_start_time,
                               oldestXid,   
                               newestXid,   
                               transID_diff,   
                               lastTransaction,
                               notificationQueueUsedPercent,
                               activeBackends,
                               recoveryMode) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`

	//source = "getCnamesForClusterEnv"
	_, err = admindb_conn.Exec(sqlStatement, run_id, host.HostID, version, bytesSincePrior, bytesSinceRedo, priorLSN, checkpointLSN, redoLSN, cluster_name, server_version, server_start_time,
		since_server_start_time, stats.OldestXid, stats.NextXid-1, transID_diff, lastTransaction, notificationQueue, activeBackends, recoveryMode)
	if err != nil {
		panic(err)
	}

	admindb_conn.Close()

}

func GetOutGoingReplicationDetails(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {
	fmt.Println("Outgoing Replication")
	admindb_conn, _ := dbadmin.Connect_to_admin_db()

	routs := stats.ReplicationOutgoing
	sqlInsertRepDetails := `insert into outgoing_replication (run_id,
                                  node_id,
                                  account_id,
                                  rep_user,
                                  Application,
                                  clientaddress,
                                  State,
                                  startedAt,
                                  sentLSN,
                                  writtenuntil,
                                  flusheduntil,
                                  replayeduntil,
                                  syncpriority,
                                  syncstate
                                  ) Values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	for _, r := range routs {
		var sp string
		if r.SyncPriority != -1 {
			sp = strconv.Itoa(r.SyncPriority)
		}

		_, err := admindb_conn.Exec(sqlInsertRepDetails, run_id, host.HostID, host.AccountID,
			r.RoleName,
			r.ApplicationName,
			r.ClientAddr,
			r.State,
			fmtTime(r.BackendStart),
			r.SentLSN,
			r.WriteLSN,
			r.FlushLSN,
			r.ReplayLSN,
			sp,
			r.SyncState)

		if err != nil {
			panic(err)
		}

	}
	admindb_conn.Close()
	fmt.Println("Finished Outgoing Replication")
}

func GetIncomingReplication(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {
	fmt.Println("Starting GetIncomingReplication")
	admindb_conn, _ := dbadmin.Connect_to_admin_db()

	ri := stats.ReplicationIncoming

	sqlInsertRecovery := `INSERT INTO incoming_replication (run_id, node_id, account_id, status, receivedLSN, 
                                  timeline, 
                                  latency, 
                                  replicationslot) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := admindb_conn.Exec(sqlInsertRecovery, run_id, host.HostID, host.AccountID, ri.Status, ri.ReceivedLSN, ri.ReceivedTLI, ri.Latency, ri.SlotName)

	if err != nil {
		panic(err)
	}

	admindb_conn.Close()
	fmt.Println("Ending GetIncomingReplication")
}

func GetRecovery(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {
	fmt.Println("Starting GetRecovery")
	admindb_conn, _ := dbadmin.Connect_to_admin_db()

	sqlRecoveryInsert := `INSERT INTO recovery (run_id,
                       node_id,
                       account_id,
    IsWalReplayPaused,
    LastWALReceiveLSN,
  LastWALReplayLSN,
  LastXActReplayTimestamp) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := admindb_conn.Exec(sqlRecoveryInsert, run_id, host.HostID, host.AccountID, stats.IsWalReplayPaused, stats.LastWALReceiveLSN, stats.LastWALReplayLSN, fmtTime(stats.LastXActReplayTimestamp))

	if err != nil {
		panic(err)
	}

	admindb_conn.Close()
	fmt.Println("Ending GetRecovery")

}

func GetReplicationSlots(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {

	fmt.Println("Starting ReplicationSlots")
	sqlInsertReplicatioslots := `INSERT INTO replicationslots (
			run_id ,
			node_id ,
			account_id ,
			SlotName ,
			Plugin ,
			SlotType ,
			DBName ,
			Active ,
			OldestTXN ,
			CatalogXmin ,
			RestartLSN ,
			ConfirmedFlushLSN ,
			Temporary) Values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	admindb_conn, _ := dbadmin.Connect_to_admin_db()

	rslots := stats.ReplicationSlots
	for _, s := range rslots {
		_, err := admindb_conn.Exec(sqlInsertReplicatioslots, run_id, host.HostID, host.AccountID, s.SlotName, s.Plugin, s.SlotType, s.DBName, s.Active, s.Xmin, s.CatalogXmin, s.RestartLSN, s.ConfirmedFlushLSN, s.Temporary)
		if err != nil {
			panic(err)
		}
	}
	admindb_conn.Close()
	fmt.Println("Ending ReplicationSlots")
}

func GetSystemMetrics(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {
	fmt.Println("Starting SystemMetrics")
	sys := stats.System

	sqlInsertSystem := `INSERT INTO systemmetrics (run_id,
	node_id ,
	account_id , 
	CPUModel,
	NumCores,
	LoadAvg,
	MemUsed,
	MemFree,
	MemBuffers,
	MemCached,
	SwapUsed,
	SwapFree) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	admindb_conn, _ := dbadmin.Connect_to_admin_db()

	_, err := admindb_conn.Exec(sqlInsertSystem, run_id, host.HostID, host.AccountID, sys.Hostname,
		sys.NumCores,
		sys.CPUModel,
		sys.LoadAvg,
		sys.MemUsed,
		sys.MemFree,
		sys.MemBuffers,
		sys.MemCached,
		sys.SwapUsed,
		sys.SwapFree)

	if err != nil {
		panic(err)
	}
	fmt.Println("Ending SystemMetrics")

}

//TODO
func GetLocks(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetBackends(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetWALDetails(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {

	fmt.Println("Starting WALDetails")

	admindb_conn, _ := dbadmin.Connect_to_admin_db()

	sqlInsertWAL := `INSERT INTO  waldetails(
		run_id,
		node_id,
		account_id,
		archivemode,
		walfiles,
		rate,
		archivedcount,
		WALReadyCount,
		LastArchivedTime,
		LastFailedTime,
		failedcount,
		StatsReset,
		max_wal_size,
	wal_level,
	archive_timeout,
	wal_compression,
	min_wal_size,
	checkpoint_timeout,
	full_page_writes,
	wal_keep_segments) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`

	archiveMode := getSetting(stats, "archive_mode") == "on"
	var rate float64
	if archiveMode {
		secs := stats.Metadata.At - stats.WALArchiving.StatsReset
		if secs > 0 {
			rate = float64(stats.WALArchiving.ArchivedCount) / (float64(secs) / 60)
		}
	}

	_, err := admindb_conn.Exec(sqlInsertWAL, run_id, host.HostID, host.AccountID,
		archiveMode, stats.WALCount,
		rate,
		stats.WALArchiving.ArchivedCount,
		stats.WALReadyCount,

		NewNullString(fmtTime(stats.WALArchiving.LastArchivedTime)),
		NewNullString(fmtTime(stats.WALArchiving.LastFailedTime)),
		stats.WALArchiving.FailedCount,
		NewNullString(fmtTime(stats.WALArchiving.StatsReset)),
		getSetting(stats, "max_wal_size"),
		getSetting(stats, "wal_level"),
		getSetting(stats, "archive_timeout"),
		getSetting(stats, "wal_compression"),
		getSetting(stats, "min_wal_size"),
		getSetting(stats, "checkpoint_timeout"),
		getSetting(stats, "full_page_writes"),
		getSetting(stats, "wal_keep_segments"))

	if err != nil {
		panic(err)
	}
	fmt.Println("Ending WALDetails")

}

//TODO
func GetBGWriter(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetVacuums(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetRoles(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetTableSpaces(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetDatabases(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetFunctions(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetExtensions(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetSlowQueries(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetTables(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetIndexes(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetMaxIntegerValues() {}

//TODO
func GetLogicalReplicationPublications(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//TODO
func GetLogicalReplicationSubscriptions(run_id int, host dbadmin.PsqlHost, stats *pgmetrics.Model) {}

//------------------------------------------------------------------------------

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func chkzerotime(valin string) (valout string) {
	if valin == "" {
		valout = ""
	} else {
		valout = valin
	}
	return valout
}

func fmtTime(at int64) string {
	if at == 0 {
		return ""
	}
	return time.Unix(at, 0).Format("2 Jan 2006 3:04:05 PM")
}

func fmtTimeDef(at int64, def string) string {
	if at == 0 {
		return def
	}
	return time.Unix(at, 0).Format("2 Jan 2006 3:04:05 PM")
}

func fmtTimeAndSince(at int64) string {
	if at == 0 {
		return ""
	}
	t := time.Unix(at, 0)
	return fmt.Sprintf("%s (%s)", t.Format("2 Jan 2006 3:04:05 PM"),
		humanize.Time(t))
}

func fmtTimeAndSinceDef(at int64, def string) string {
	if at == 0 {
		return def
	}
	t := time.Unix(at, 0)
	return fmt.Sprintf("%s (%s)", t.Format("2 Jan 2006 3:04:05 PM"),
		humanize.Time(t))
}

func fmtSince(at int64) string {
	if at == 0 {
		return "never"
	}
	return humanize.Time(time.Unix(at, 0))
}

func fmtYesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func fmtYesBlank(v bool) string {
	if v {
		return "yes"
	}
	return ""
}

func fmtSeconds(s string) string {
	v, err := strconv.Atoi(s)
	if err != nil {
		return s
	}
	return (time.Duration(v) * time.Second).String()
}

func fmtLag(a, b, qual string) string {
	if len(qual) > 0 && !strings.HasSuffix(qual, " ") {
		qual += " "
	}
	if d, ok := lsnDiff(a, b); ok {
		if d == 0 {
			return " (no " + qual + "lag)"
		}
		return fmt.Sprintf(" (%slag = %s)", qual, humanize.IBytes(uint64(d)))
	}
	return ""
}

func fmtIntZero(i int) string {
	if i == 0 {
		return ""
	}
	return strconv.Itoa(i)
}

func fmtPropagate(ins, upd, del bool) string {
	parts := make([]string, 0, 3)
	if ins {
		parts = append(parts, "inserts")
	}
	if upd {
		parts = append(parts, "updates")
	}
	if del {
		parts = append(parts, "deletes")
	}
	return strings.Join(parts, ", ")
}

func fmtMicros(v int64) string {
	s := (time.Duration(v) * time.Microsecond).String()
	return strings.Replace(s, "µ", "u", -1)
}

func getSetting(result *pgmetrics.Model, key string) string {
	if s, ok := result.Settings[key]; ok {
		return s.Setting
	}
	return ""
}

func getSettingInt(result *pgmetrics.Model, key string) int {
	s := getSetting(result, key)
	if len(s) == 0 {
		return 0
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

func getSettingBytes(result *pgmetrics.Model, key string, factor uint64) string {
	s := getSetting(result, key)
	if len(s) == 0 {
		return s
	}
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil || val <= 0 {
		return s
	}
	return s + " (" + humanize.IBytes(val*factor) + ")"
}

func safeDiv(a, b int64) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}

func lsn2int(s string) int64 {
	if len(s) == 0 {
		return -1
	}
	if pos := strings.IndexByte(s, '/'); pos >= 0 {
		val1, err1 := strconv.ParseUint(s[:pos], 16, 64)
		val2, err2 := strconv.ParseUint(s[pos+1:], 16, 64)
		if err1 != nil || err2 != nil {
			return -1
		}
		return int64(val1<<32 | val2)
	}
	return -1
}

func lsnDiff(a, b string) (int64, bool) {
	va := lsn2int(a)
	vb := lsn2int(b)
	if va == -1 || vb == -1 {
		return -1, false
	}
	return va - vb, true
}

func getBlockSize(result *pgmetrics.Model) int {
	s := getSetting(result, "block_size")
	if len(s) == 0 {
		return 8192
	}
	v, err := strconv.Atoi(s)
	if err != nil || v == 0 {
		return 8192
	}
	return v
}

func getVersion(result *pgmetrics.Model) int {
	s := getSetting(result, "server_version_num")
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

func getMaxWalSize(result *pgmetrics.Model) (string, string) {
	var key string
	if version := getVersion(result); version >= 90500 {
		key = "max_wal_size"
	} else {
		key = "checkpoint_segments"
	}
	return key, getSettingBytes(result, key, 16*1024*1024)
}

//------------------------------------------------------------------------------

type tableWriter struct {
	data      [][]string
	hasFooter bool
}

func (t *tableWriter) add(cols ...interface{}) {
	row := make([]string, len(cols))
	for i, c := range cols {
		row[i] = fmt.Sprintf("%v", c)
	}
	t.data = append(t.data, row)
}

func (t *tableWriter) clear() {
	t.data = nil
}

func (t *tableWriter) cols() int {
	n := 0
	for _, row := range t.data {
		if n < len(row) {
			n = len(row)
		}
	}
	return n
}

func (t *tableWriter) write(fd io.Writer, pfx string) (tw int) {
	if len(t.data) == 0 {
		return
	}
	ncols := t.cols()
	if ncols == 0 {
		return
	}
	// calculate widths
	widths := make([]int, ncols)
	for _, row := range t.data {
		for c, col := range row {
			w := len(col)
			if widths[c] < w {
				widths[c] = w
			}
		}
	}
	// calculate total width
	tw = len(pfx) + 1 // "prefix", "|"
	for _, w := range widths {
		tw += 1 + w + 1 + 1 // blank, "value", blank, "|"
	}
	// print line
	line := func() {
		fmt.Fprintf(fd, "%s+", pfx)
		for _, w := range widths {
			fmt.Fprint(fd, strings.Repeat("-", w+2))
			fmt.Fprintf(fd, "+")
		}
		fmt.Fprintln(fd)
	}
	line()
	for i, row := range t.data {
		if i == 1 || (t.hasFooter && i == len(t.data)-1) {
			line()
		}
		fmt.Fprintf(fd, "%s|", pfx)
		for c, col := range row {
			fmt.Fprintf(fd, " %*s |", widths[c], col)
		}
		fmt.Fprintln(fd)
	}
	line()
	return
}
