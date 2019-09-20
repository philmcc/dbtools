create extension ltree;

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
  account_id integer REFERENCES account(account_id),
  "AWS" boolean NOT NULL DEFAULT false
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
    node_id integer NOT NULL,
    hostname text,
    ip_address text,
    last_checked timestamp without time zone NOT NULL,
    parent_id integer,
    cluster_id integer,
    env_id integer,
    collect_stats boolean NOT NULL,
    stats_frequency text,
    last_collected timestamp without time zone,
    account_id integer,
    parent_path ltree
);
CREATE INDEX node_parent_id_idx ON public.node USING btree (parent_id);
CREATE INDEX node_parent_path_idx ON public.node USING gist (parent_path);

CREATE OR REPLACE FUNCTION update_node_parent_path() RETURNS TRIGGER AS $$
    DECLARE
        path ltree;
    BEGIN
        IF NEW.parent_id = 0 THEN
            NEW.parent_path = 'root'::ltree;
        ELSEIF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' OR OLD.parent_id IS NULL OR OLD.parent_id != NEW.parent_id THEN
            SELECT parent_path || node_id::text FROM node WHERE node_id = NEW.parent_id INTO path;
            IF path IS NULL THEN
                RAISE EXCEPTION 'Invalid parent_id %', NEW.parent_id;
            END IF;
            NEW.parent_path = path;
        END IF;
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER parent_path_tgr
    BEFORE INSERT OR UPDATE ON node
    FOR EACH ROW EXECUTE PROCEDURE update_node_parent_path();


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


  create or replace view current_cname_host as
  select c.cname, ch1.hostname, c.cname_order
from public.cname_history ch1,
    (select max(run_id) run_id, cname_id
        FROM public.cname_history ch2
        group by cname_id) ch2,
    cnames c
WHERE ch2.run_id= ch1.run_id
and ch2.cname_id = ch1.cname_id
and c.cname_id = ch1.cname_id
order by cname
;


  CREATE OR REPLACE VIEW mapped_cluster AS
   WITH RECURSIVE replicas AS (
           SELECT n1.hostname,
              cname.cname,
              n1.node_id,
              n1.parent_id,
              e1.env,
              c1.cluster,
              n1.last_checked,
              n1.collect_stats,
              cname.cname_order
             FROM node n1
               JOIN environments e1 ON e1.env_id = n1.env_id
               JOIN clusters c1 ON c1.cluster_id = n1.cluster_id
               LEFT JOIN current_cname_host cname ON n1.hostname = cname.hostname
            WHERE n1.parent_id = 0
          UNION
           SELECT n.hostname,
              cname.cname,
              n.node_id,
              n.parent_id,
              e.env,
              c.cluster,
              n.last_checked,
              n.collect_stats,
              cname.cname_order
             FROM node n
               JOIN replicas r ON r.node_id = n.parent_id
               JOIN environments e ON e.env_id = n.env_id
               JOIN clusters c ON c.cluster_id = n.cluster_id
               LEFT JOIN current_cname_host cname ON n.hostname = cname.hostname
          )
   SELECT replicas.hostname,
      replicas.cname,
      replicas.node_id,
      replicas.parent_id,
      replicas.env,
      replicas.cluster,
      replicas.last_checked,
      replicas.collect_stats,
      replicas.cname_order,
      node.parent_path
     FROM replicas, node
     where replicas.node_id = node.node_id
    ORDER BY node.parent_path, replicas.cname_order, replicas.hostname;



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

INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'sprint'),'uk-master.sprint.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'sprint'),'uk-slave1.sprint.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'sprint'),'uk-slave2.sprint.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'sprint'),'uk-backup.sprint.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'sprint'),'uk-cron.sprint.iparadigms.com',31);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'sprint'),'uk-relay.sprint.iparadigms.com',21);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'sprint'),'uk-tip.sprint.iparadigms.com',22);

INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'live'),'uk-master.live.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'live'),'uk-slave1.live.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'live'),'uk-slave2.live.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'live'),'uk-slave3.live.iparadigms.com',13);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'live'),'uk-slave4.live.iparadigms.com',14);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'live'),'uk-backup.live.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'live'),'uk-cron.live.iparadigms.com',31);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'live'),'uk-relay.live.iparadigms.com',21);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'live'),'uk-tip.live.iparadigms.com',22);


INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'dev'),'uk-master.dev.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'dev'),'uk-slave1.dev.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'dev'),'uk-slave2.dev.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'dev'),'uk-backup.dev.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'dev'),'uk-cron.dev.iparadigms.com',31);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'dev'),'uk-relay.dev.iparadigms.com',21);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'uk'), (select env_id from environments where env = 'dev'),'uk-tip.dev.iparadigms.com',22);

INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'sprint'),'global-master.sprint.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'sprint'),'global-slave1.sprint.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'sprint'),'global-slave2.sprint.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'sprint'),'global-backup.sprint.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'sprint'),'global-cron.sprint.iparadigms.com',31);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'sprint'),'global-relay.sprint.iparadigms.com',21);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'sprint'),'global-tip.sprint.iparadigms.com',22);

INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'live'),'global-master.live.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'live'),'global-slave1.live.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'live'),'global-slave2.live.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'live'),'global-slave3.live.iparadigms.com',13);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'live'),'global-slave4.live.iparadigms.com',14);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'live'),'global-backup.live.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'live'),'global-cron.live.iparadigms.com',31);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'live'),'global-relay.live.iparadigms.com',21);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'live'),'global-tip.live.iparadigms.com',22);


INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'dev'),'global-master.dev.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'dev'),'global-slave1.dev.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'dev'),'global-slave2.dev.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'dev'),'global-backup.dev.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'dev'),'global-cron.dev.iparadigms.com',31);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'dev'),'global-relay.dev.iparadigms.com',21);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'global'), (select env_id from environments where env = 'dev'),'global-tip.dev.iparadigms.com',22);

INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'sprint'),'ares-master.sprint.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'sprint'),'ares-slave1.sprint.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'sprint'),'ares-slave2.sprint.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'sprint'),'ares-backup.sprint.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'sprint'),'ares-cron.sprint.iparadigms.com',31);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'sprint'),'ares-relay.sprint.iparadigms.com',21);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'sprint'),'ares-tip.sprint.iparadigms.com',22);

INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'live'),'ares-master.live.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'live'),'ares-slave1.live.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'live'),'ares-slave2.live.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'live'),'ares-slave3.live.iparadigms.com',13);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'live'),'ares-slave4.live.iparadigms.com',14);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'live'),'ares-backup.live.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'live'),'ares-cron.live.iparadigms.com',31);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'live'),'ares-relay.live.iparadigms.com',21);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'live'),'ares-tip.live.iparadigms.com',22);


INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'dev'),'ares-master.dev.iparadigms.com',1);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'dev'),'ares-slave1.dev.iparadigms.com',11);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'dev'),'ares-slave2.dev.iparadigms.com',12);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'dev'),'ares-backup.dev.iparadigms.com',41);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'dev'),'ares-cron.dev.iparadigms.com',31);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'dev'),'ares-relay.dev.iparadigms.com',21);
INSERT INTO cnames (cluster_id, env_id, cname, cname_order) VALUES ((select cluster_id from clusters where cluster = 'ares'), (select env_id from environments where env = 'dev'),'ares-tip.dev.iparadigms.com',22);
