// pg_shdescription

// pg_description

package pgcatalog

type PGTable interface {
	DDL() string
}

var PGTables = []PGTable{
	pg_am{},
	pg_attrdef{},
	pg_attribute{},
	pg_class{},
	pg_constraint{},
	pg_cursors{},
	pg_depend{},
	pg_database{},
	pg_enum{},
	pg_event_trigger{},
	pg_index{},
	pg_indexes{},
	pg_locks{},
	pg_namespace{},
	pg_proc{},
	pg_publication{},
	pg_publication_tables{},
	pg_range{},
	pg_roles{},
	pg_settings{},
	pg_stats{},
	pg_subscription{},
	pg_subscription_rel{},
	pg_tables{},
	pg_tablespace{},
	pg_type{},
	pg_views{},
}

type pg_am struct{}

func (pg_am) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_am (
		oid oid NOT NULL,
		amname name NOT NULL,
		amhandler regproc NOT NULL,
		amtype "char" NOT NULL
	)`
}

type pg_attrdef struct{}

func (pg_attrdef) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_attrdef (
		oid oid NOT NULL,
		adrelid oid NOT NULL,
		adnum smallint NOT NULL,
		adbin pg_node_tree NOT NULL COLLATE pg_catalog."C"
	)`
}

type pg_attribute struct{}

func (pg_attribute) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_attribute (
		attrelid oid NOT NULL,
		attname name NOT NULL,
		atttypid oid NOT NULL,
		attstattarget integer NOT NULL,
		attlen smallint NOT NULL,
		attnum smallint NOT NULL,
		attndims integer NOT NULL,
		attcacheoff integer NOT NULL,
		atttypmod integer NOT NULL,
		attbyval boolean NOT NULL,
		attalign "char" NOT NULL,
		attstorage "char" NOT NULL,
		attcompression "char" NOT NULL,
		attnotnull boolean NOT NULL,
		atthasdef boolean NOT NULL,
		atthasmissing boolean NOT NULL,
		attidentity "char" NOT NULL,
		attgenerated "char" NOT NULL,
		attisdropped boolean NOT NULL,
		attislocal boolean NOT NULL,
		attinhcount integer NOT NULL,
		attcollation oid NOT NULL,
		attacl aclitem[],
		attoptions text[] COLLATE pg_catalog."C",
		attfdwoptions text[] COLLATE pg_catalog."C",
		attmissingval anyarray
	)`
}

type pg_class struct{}

func (pg_class) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_class (
		oid oid NOT NULL,
		relname name NOT NULL,
		relnamespace oid NOT NULL,
		reltype oid NOT NULL,
		reloftype oid NOT NULL,
		relowner oid NOT NULL,
		relam oid NOT NULL,
		relfilenode oid NOT NULL,
		reltablespace oid NOT NULL,
		relpages integer NOT NULL,
		reltuples real NOT NULL,
		relallvisible integer NOT NULL,
		reltoastrelid oid NOT NULL,
		relhasindex boolean NOT NULL,
		relisshared boolean NOT NULL,
		relpersistence "char" NOT NULL,
		relkind "char" NOT NULL,
		relnatts smallint NOT NULL,
		relchecks smallint NOT NULL,
		relhasrules boolean NOT NULL,
		relhastriggers boolean NOT NULL,
		relhassubclass boolean NOT NULL,
		relrowsecurity boolean NOT NULL,
		relforcerowsecurity boolean NOT NULL,
		relispopulated boolean NOT NULL,
		relreplident "char" NOT NULL,
		relispartition boolean NOT NULL,
		relrewrite oid NOT NULL,
		relfrozenxid xid NOT NULL,
		relminmxid xid NOT NULL,
		relacl aclitem[],
		reloptions text[] COLLATE pg_catalog."C",
		relpartbound pg_node_tree COLLATE pg_catalog."C"
	)`
}

type pg_constraint struct{}

func (pg_constraint) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_constraint (
		oid oid NOT NULL,
		conname name NOT NULL,
		connamespace oid NOT NULL,
		contype "char" NOT NULL,
		condeferrable boolean NOT NULL,
		condeferred boolean NOT NULL,
		convalidated boolean NOT NULL,
		conrelid oid NOT NULL,
		contypid oid NOT NULL,
		conindid oid NOT NULL,
		conparentid oid NOT NULL,
		confrelid oid NOT NULL,
		confupdtype "char" NOT NULL,
		confdeltype "char" NOT NULL,
		confmatchtype "char" NOT NULL,
		conislocal boolean NOT NULL,
		coninhcount integer NOT NULL,
		connoinherit boolean NOT NULL,
		conkey smallint[],
		confkey smallint[],
		conpfeqop oid[],
		conppeqop oid[],
		conffeqop oid[],
		confdelsetcols smallint[],
		conexclop oid[],
		conbin pg_node_tree COLLATE pg_catalog."C"
	)`
}

type pg_cursors struct{}

func (pg_cursors) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_cursors (
		name STRING,
		statement STRING,
		is_holdable BOOL,
		is_binary BOOL,
		is_scrollable BOOL,
		creation_time TIMESTAMPTZ
	)`
}

type pg_database struct{}

func (pg_database) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_database (
		oid OID,
		datname Name,
		datdba OID,
		encoding INT4,
		datcollate STRING,
		datctype STRING,
		datistemplate BOOL,
		datallowconn BOOL,
		datconnlimit INT4,
		datlastsysoid OID,
		datfrozenxid INT,
		datminmxid INT,
		dattablespace OID,
		datacl STRING[]
	)`
}

type pg_depend struct{}

func (pg_depend) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_depend (
		classid OID,
		objid OID,
		objsubid INT4,
		refclassid OID,
		refobjid OID,
		refobjsubid INT4,
		deptype "char"
  	)`
}

type pg_enum struct{}

func (pg_enum) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_enum (
		oid OID,
		enumtypid OID,
		enumsortorder FLOAT4,
		enumlabel STRING
	)`
}

type pg_event_trigger struct{}

func (pg_event_trigger) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_event_trigger (
		evtname NAME,
		evtevent NAME,
		evtowner OID,
		evtfoid OID,
		evtenabled "char",
		evttags TEXT[],
		oid OID
	)`
}

type pg_index struct{}

func (pg_index) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_index (
		indexrelid OID,
		indrelid OID,
		indnatts INT2,
		indisunique BOOL,
		indnullsnotdistinct BOOL,
		indisprimary BOOL,
		indisexclusion BOOL,
		indimmediate BOOL,
		indisclustered BOOL,
		indisvalid BOOL,
		indcheckxmin BOOL,
		indisready BOOL,
		indislive BOOL,
		indisreplident BOOL,
		indkey INT2VECTOR,
		indcollation OIDVECTOR,
		indclass OIDVECTOR,
		indoption INT2VECTOR,
		indexprs STRING,
		indpred STRING,
		indnkeyatts INT2
	)`
}

type pg_indexes struct{}

func (pg_indexes) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_indexes (
		crdb_oid OID,
		schemaname NAME,
		tablename NAME,
		indexname NAME,
		tablespace NAME,
		indexdef STRING
	)`
}

type pg_locks struct{}

func (pg_locks) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_locks (
		locktype TEXT,
		database OID,
		relation OID,
		page INT4,
		tuple SMALLINT,
		virtualxid TEXT,
		transactionid INT,
		classid OID,
		objid OID,
		objsubid SMALLINT,
		virtualtransaction TEXT,
		pid INT4,
		mode TEXT,
		granted BOOLEAN,
		fastpath BOOLEAN
	)`
}

type pg_namespace struct{}

func (pg_namespace) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_namespace (
		oid OID,
		nspname NAME NOT NULL,
		nspowner OID,
		nspacl STRING[],
		INDEX (oid)
	)`
}

type pg_proc struct{}

func (pg_proc) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_proc (
		oid OID,
		proname NAME,
		pronamespace OID,
		proowner OID,
		prolang OID,
		procost FLOAT4,
		prorows FLOAT4,
		provariadic OID,
		prosupport REGPROC,
		prokind "char",
		prosecdef BOOL,
		proleakproof BOOL,
		proisstrict BOOL,
		proretset BOOL,
		provolatile "char",
		proparallel "char",
		pronargs INT2,
		pronargdefaults INT2,
		prorettype OID,
		proargtypes OIDVECTOR,
		proallargtypes OID[],
		proargmodes "char"[],
		proargnames STRING[],
		proargdefaults STRING,
		protrftypes OID[],
		prosrc STRING,
		probin STRING,
		prosqlbody STRING,
		proconfig STRING[],
		proacl STRING[],
		INDEX(oid)
	)`
}

type pg_publication struct{}

func (pg_publication) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_publication (
		oid OID,
		pubname NAME,
		pubowner OID,
		puballtables BOOL,
		pubinsert BOOL,
		pubupdate BOOL,
		pubdelete BOOL,
		pubtruncate BOOL,
		pubviaroot BOOL
	)`
}

type pg_publication_tables struct{}

func (pg_publication_tables) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_publication_tables (
		pubname NAME,
		schemaname NAME,
		tablename NAME
	)`
}

type pg_range struct{}

func (pg_range) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_range (
		rngtypid OID,
		rngsubtype OID,
		rngcollation OID,
		rngsubopc OID,
		rngcanonical OID,
		rngsubdiff OID
	)`
}

type pg_roles struct{}

func (pg_roles) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_roles (
		oid OID,
		rolname NAME,
		rolsuper BOOL,
		rolinherit BOOL,
		rolcreaterole BOOL,
		rolcreatedb BOOL,
		rolcatupdate BOOL,
		rolcanlogin BOOL,
		rolreplication BOOL,
		rolconnlimit INT4,
		rolpassword STRING,
		rolvaliduntil TIMESTAMPTZ,
		rolbypassrls BOOL,
		rolconfig STRING[]
	)`
}

type pg_settings struct{}

func (pg_settings) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_settings (
		name STRING,
		setting STRING,
		unit STRING,
		category STRING,
		short_desc STRING,
		extra_desc STRING,
		context STRING,
		vartype STRING,
		source STRING,
		min_val STRING,
		max_val STRING,
		enumvals STRING,
		boot_val STRING,
		reset_val STRING,
		sourcefile STRING,
		sourceline INT4,
		pending_restart BOOL
	)`
}

type pg_stats struct{}

func (pg_stats) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_stats (
		schemaname NAME,
		tablename NAME,
		attname NAME,
		inherited BOOL,
		null_frac FLOAT4,
		avg_width INT4,
		n_distinct FLOAT4,
		most_common_vals STRING[],
		most_common_freqs FLOAT4[],
		histogram_bounds STRING[],
		correlation FLOAT4,
		most_common_elems STRING[],
		most_common_elem_freqs FLOAT4[],
		elem_count_histogram FLOAT4[]
	)`
}

type pg_subscription struct{}

func (pg_subscription) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_subscription (
		oid OID,
		subdbid OID,
		subname NAME,
		subowner OID,
		subenabled BOOL,
		subconninfo STRING,
		subslotname NAME,
		subsynccommit STRING,
		subpublications STRING[]
	)`
}

type pg_subscription_rel struct{}

func (pg_subscription_rel) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_subscription_rel (
		srsubid OID,
		srrelid OID,
		srsubstate "char",
		srsublsn STRING
	)`
}

type pg_tables struct{}

func (pg_tables) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_tables (
		schemaname NAME,
		tablename NAME,
		tableowner NAME,
		tablespace NAME,
		hasindexes BOOL,
		hasrules BOOL,
		hastriggers BOOL,
		rowsecurity BOOL
	)`
}

type pg_tablespace struct{}

func (pg_tablespace) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_tablespace (
		oid OID,
		spcname NAME,
		spcowner OID,
		spclocation TEXT,
		spcacl TEXT[],
		spcoptions TEXT[]
	)`
}

type pg_type struct{}

func (pg_type) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_type (
		oid OID NOT NULL,
		typname NAME NOT NULL,
		typnamespace OID,
		typowner OID,
		typlen INT2,
		typbyval BOOL,
		typtype "char",
		typcategory "char",
		typispreferred BOOL,
		typisdefined BOOL,
		typdelim "char",
		typrelid OID,
		typelem OID,
		typarray OID,
		typinput REGPROC,
		typoutput REGPROC,
		typreceive REGPROC,
		typsend REGPROC,
		typmodin REGPROC,
		typmodout REGPROC,
		typanalyze REGPROC,
		typalign "char",
		typstorage "char",
		typnotnull BOOL,
		typbasetype OID,
		typtypmod INT4,
		typndims INT4,
		typcollation OID,
		typdefaultbin STRING,
		typdefault STRING,
		typacl STRING[],
	  INDEX(oid)
	)`
}

type pg_views struct{}

func (pg_views) DDL() string {
	return `CREATE TABLE IF NOT EXISTS pg_catalog.pg_views (
		schemaname NAME,
		viewname NAME,
		viewowner NAME,
		definition STRING
	)`
}