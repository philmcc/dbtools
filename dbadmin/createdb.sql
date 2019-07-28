CREATE TABLE account (
  account_id SERIAL PRIMARY KEY,
  account_name TEXT
);

CREATE TABLE clusters  (
	cluster_id serial primary key,
  cluster TEXT UNIQUE NOT NULL,
  account_id integer REFERENCES account(account_id));

  CREATE TABLE environments (
  env_id SERIAL PRIMARY KEY,
  env TEXT,
  account_id integer REFERENCES account(account_id)
  );

CREATE TABLE cnames (
cname_id serial primary key,
cluster_id INTEGER REFERENCES clusters(cluster_id),
cname TEXT UNIQUE NOT NULL,
cname_order int,
env_id INTEGER REFERENCES environments(env_id),
active bool NOT NULL DEFAULT true,
account_id integer REFERENCES account(account_id)
);

CREATE TABLE runs (

run_id SERIAL PRIMARY KEY,
run_date TIMESTAMP NOT NULL DEFAULT NOW(),
run_source TEXT,
account_id integer REFERENCES account(account_id)
);

CREATE TABLE cname_history (
history_id SERIAL PRIMARY KEY,
run_id INTEGER REFERENCES runs(run_id),
cname_id INTEGER REFERENCES cnames(cname_id),
hostname TEXT,
ip_address TEXT,
account_id integer REFERENCES account(account_id)
);

CREATE TABLE node (
	node_id SERIAL PRIMARY KEY,
	hostname TEXT UNIQUE,
	ip_address TEXT,
	last_checked  TIMESTAMP NOT NULL DEFAULT NOW(),
  parent_id int default 0,
  cluster_id INT,
  env_id INT,
  collect_stats bool not null default false,
  stats_frequency TEXT,
  last_collected   TIMESTAMP,
  account_id integer REFERENCES account(account_id)
	);

CREATE TABLE STATS (
  stats_id SERIAL PRIMARY KEY,
  run_id INTEGER NOT NULL REFERENCES runs(run_id),
  node_id INTEGER REFERENCES node(node_id),
  stats jsonb,
  account_id integer REFERENCES account(account_id)
);



Create TABLE dbcluster (
  dbcluster_id SERIAL PRIMARY KEY,
  run_id INTEGER NOT NULL REFERENCES runs(run_id),
  node_id INTEGER REFERENCES node(node_id),
  version INTEGER,
  bytesSincePrior TEXT,
  bytesSinceRedo TEXT,
  priorLSN TEXT,
  checkpointLSN TEXT,
  redoLSN TEXT,
  report_start_time TIMESTAMP,
  cluster_name TEXT,
  server_version FLOAT,
  server_start_time TIMESTAMP,
  since_server_start_time TEXT,
  oldestXid INTEGER,
  newestXid INTEGER,
  transID_diff TEXT,
  lastTransaction TEXT,
  notificationQueueUsedPercent FLOAT,
  activeBackends INTEGER,
  recoveryMode BOOLEAN,
  account_id integer REFERENCES account(account_id)
);

CREATE TABLE outgoing_replication (
  id serial primary key,
  run_id INTEGER REFERENCES runs(run_id),
  node_id INTEGER REFERENCES node(node_id),
  rep_user TEXT,
    account_id integer,
      Application TEXT,
      clientaddress TEXT,
      State TEXT,
      startedAt TIMESTAMP,
      sentLSN TEXT,
      writtenuntil TEXT,
      flusheduntil TEXT,
      replayeduntil TEXT,
      syncpriority INTEGER,
      syncstate TEXT);




CREATE TABLE incoming_replication (id SERIAL PRIMARY KEY,
run_id INTEGER REFERENCES runs(run_id),
node_id INTEGER REFERENCES node(node_id),
account_id INTEGER REFERENCES account(account_id),
status TEXT,
receivedLSN TEXT,
timeline INTEGER,
latency float,
replicationslot TEXT

);

CREATE TABLE recovery(
                       run_id INTEGER REFERENCES runs(run_id),
                       node_id INTEGER REFERENCES node(node_id),
                       account_id INTEGER REFERENCES account(account_id),
    IsWalReplayPaused BOOLEAN,
    LastWALReceiveLSN TEXT,
  LastWALReplayLSN TEXT,
  LastXActReplayTimestamp TIMESTAMP
);


CREATE TABLE replicationslots(
  slot_id SERIAL PRIMARY KEY,
run_id INTEGER REFERENCES runs(run_id),
node_id INTEGER REFERENCES node(node_id),
account_id INTEGER REFERENCES account(account_id),
SlotName TEXT,
Plugin TEXT,
SlotType TEXT,
DBName TEXT,
Active BOOLEAN,
OldestTXN INTEGER,
CatalogXmin INTEGER,
RestartLSN TEXT,
ConfirmedFlushLSN TEXT,
Temporary BOOLEAN);

CREATE TABLE systemmetrics (
run_id INTEGER REFERENCES runs(run_id),
node_id INTEGER REFERENCES node(node_id),
account_id INTEGER REFERENCES account(account_id),
CPUModel   TEXT,
NumCores   INTEGER,
LoadAvg    FLOAT,
MemUsed    BIGINT,
MemFree    BIGINT,
MemBuffers BIGINT,
MemCached  BIGINT,
SwapUsed   BIGINT,
SwapFree   BIGINT);


CREATE TABLE waldetails(
                         run_id INTEGER REFERENCES runs(run_id),
                         node_id INTEGER REFERENCES node(node_id),
                         account_id INTEGER REFERENCES account(account_id),
                         archivemode BOOLEAN,
                         walfiles INTEGER,
                         rate FLOAT,
                         archivedcount INTEGER,
                         WALReadyCount INTEGER,
                         LastArchivedTime TIMESTAMP,
                         LastFailedTime TIMESTAMP,
                         failedcount INTEGER,
                         StatsReset TIMESTAMP,
                         max_wal_size INTEGER,
                         wal_level TEXT,
                         archive_timeout INTEGER,
                         wal_compression BOOLEAN,
                         min_wal_size INTEGER,
                         checkpoint_timeout INTEGER,
                         full_page_writes BOOLEAN,
                         wal_keep_segments INTEGER);



Create view





SELECT cname_id, cluster_id, cname, env_id
		FROM  cnames
		WHERE env_id in (select env_id from environments where env ilike TRIM('%'))
		AND cluster_id in (select cluster_id from clusters where cluster ilike TRIM('%'))
		AND active is true
		ORDER BY cname_order asc;

//Update node set hostname = 'staging-db4.s2prod', ip_address = '10.2.144.21' , parent_id = 1 where node_id = 2 RETURNING node_id;



SELECT node_id, hostname, ip_address
FROM  node
WHERE env_id in (select env_id from environments where env ilike TRIM('sprint'))
  AND cluster_id in (select cluster_id from clusters where cluster ilike TRIM(' tii'))
  AND collect_stats is true;


create view mapped_cluster as
WITH RECURSIVE replicas AS (
  SELECT n1.hostname, n1.node_id, n1.parent_id, e1.env, c1.cluster, n1.last_checked, n1.collect_stats
  FROM node n1
  INNER JOIN environments e1 ON e1.env_id = n1.env_id
  INNER JOIN clusters c1 ON c1.cluster_id = n1.cluster_id
  WHERE parent_id = 0
  UNION
  SELECT n.hostname, n.node_id, n.parent_id, e.env, c.cluster, n.last_checked, n.collect_stats
  FROM node n
         INNER JOIN replicas r ON r.node_id = n.parent_id
         INNER JOIN environments e ON e.env_id = n.env_id
         INNER JOIN clusters c ON c.cluster_id = n.cluster_id
) SELECT * FROM replicas;




//////////////////////////

INSERT INTO account (account_name) VALUES ('Turnitin');

INSERT INTO environments (env) VALUES ('live');
INSERT INTO environments (env) VALUES ('sprint');
INSERT INTO environments (env) VALUES ('dev');
INSERT INTO environments (env) VALUES ('upgrade');

INSERT INTO clusters (cluster) VALUES ('tii');
INSERT INTO clusters (cluster) VALUES ('marks');


INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'sprint'),'tii-master.sprint.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'sprint'),'tii-slave1.sprint.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'sprint'),'tii-slave2.sprint.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'sprint'),'tii-backup.sprint.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'sprint'),'tii-cron.sprint.iparadigms.com',31);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'sprint'),'tii-relay.sprint.iparadigms.com',21);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'sprint'),'tii-tip.sprint.iparadigms.com',22);

INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'live'),'tii-master.live.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'live'),'tii-slave1.live.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'live'),'tii-slave2.live.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'live'),'tii-slave3.live.iparadigms.com',13);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'live'),'tii-slave4.live.iparadigms.com',14);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'live'),'tii-backup.live.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'live'),'tii-cron.live.iparadigms.com',31);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'live'),'tii-relay.live.iparadigms.com',21);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'live'),'tii-tip.live.iparadigms.com',22);

INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'sprint'),'marks-master.sprint.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'sprint'),'marks-slave1.sprint.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'sprint'),'marks-slave2.sprint.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'sprint'),'marks-slave3.sprint.iparadigms.com',13);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'sprint'),'marks-backup.sprint.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'sprint'),'marks-cron.sprint.iparadigms.com',31);

INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'live'),'marks-master.live.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'live'),'marks-slave1.live.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'live'),'marks-slave2.live.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'live'),'marks-slave3.live.iparadigms.com',13);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'live'),'marks-slave4.live.iparadigms.com',14);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'live'),'marks-backup.live.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'live'),'marks-cron.live.iparadigms.com',31);


INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'dev'),'tii-master.dev.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'dev'),'tii-slave1.dev.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'dev'),'tii-slave2.dev.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'dev'),'tii-backup.dev.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'dev'),'tii-cron.dev.iparadigms.com',31);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'dev'),'tii-relay.dev.iparadigms.com',21);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'dev'),'tii-tip.dev.iparadigms.com',22);

INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'dev'),'marks-master.dev.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'dev'),'marks-slave1.dev.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'dev'),'marks-slave2.dev.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'dev'),'marks-slave3.dev.iparadigms.com',13);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'dev'),'marks-backup.dev.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'marks'), (select env_id from environments where env = 'dev'),'marks-cron.dev.iparadigms.com',31);

INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'upgrade'),'tii-master.upgrade.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'tii'), (select env_id from environments where env = 'upgrade'),'tii-slave1.upgrade.iparadigms.com',11);
