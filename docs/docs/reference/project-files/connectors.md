---
title: Connector YAML
sidebar_label: Connector YAML 
sidebar_position: 00
toc_max_heading_level: 4

---

Connector YAML files define how Rill connects to external data sources and OLAP engines. Each connector specifies a driver type and its required connection parameters. 

## Available Connectors

#### _OLAP Engines_

- **[DuckDB](#duckdb)** - Embedded DuckDB engine (_default_)
- **[Clickhouse](#clickhouse)** - ClickHouse analytical database
- **[MotherDuck](#motherduck)** - MotherDuck cloud database
- **[Druid](#druid)** - Apache Druid
- **[Pinot](#pinot)** - Apache Pinot

#### _Data Warehouses_
- **[Snowflake](#snowflake)** - Snowflake data warehouse
- **[BigQuery](#bigquery)** - Google BigQuery
- **[Redshift](#redshift)** - Amazon Redshift
- **[Athena](#athena)** - Amazon Athena

#### _Databases_
- **[PostgreSQL](#postgresql)** - PostgreSQL databases
- **[MySQL](#mysql)** - MySQL databases
- **[sqlite](#sqlite)** - SQLite databases

#### _Cloud Storage_
- **[GCS](#gcs)** - Google Cloud Storage
- **[S3](#s3)** - Amazon S3 storage
- **[Azure](#azure)** - Azure Blob Storage

#### _Other_
- **[https](#https)** - Public files via HTTP/HTTPS
- **[Salesforce](#salesforce)** - Salesforce data
- **[Slack](#slack)** - Slack data
- **[Google Sheets](#google-sheets)** - Public Google Sheets



## Required Properties

**`type`** - Must be `connector` (required)

**`driver`** - The type of connector, see [available connectors](#available-connectors) (required)


:::warning Security Recommendation
For all credential parameters (passwords, tokens, keys), use environment variables with the syntax `{{.env.connector.<connector_driver>.<parameter_name>}}`. This keeps sensitive data out of your YAML files and version control. See our [credentials documentation](/connect/credentials) for complete setup instructions.
:::

### Athena
```yaml
type: connector                                  # Must be `connector` (required)
driver: athena                                   # Must be `athena` _(required)_

aws_access_key_id: "myawsaccesskey"              # AWS Access Key ID for authentication  
aws_secret_access_key: "myawssecretkey"          # AWS Secret Access Key for authentication  
aws_access_token: "mytemporarytoken"             # AWS session token for temporary credentials  
role_arn: "arn:aws:iam::123456789012:role/MyRole" # ARN of the IAM role to assume  
role_session_name: "MySession"                   # Session name for STS AssumeRole  
external_id: "MyExternalID"                      # External ID for cross-account access  
workgroup: "primary"                             # Athena workgroup (defaults to 'primary')  
output_location: "s3://my-bucket/athena-output/" # S3 URI for query results  
aws_region: "us-east-1"                          # AWS region (defaults to 'us-east-1')  
allow_host_access: true                          # Allow host environment access _(default: true)_
```

### Azure
```yaml
type: connector                                  # Must be `connector` (required)
driver: azure                                    # Must be `azure` _(required)_

azure_storage_account: "mystorageaccount"        # Azure storage account name  
azure_storage_key: "credentialjsonstring"        # Azure storage access key  
azure_storage_sas_token: "optionaltoken"         # Optional SAS token for authentication  
azure_storage_connection_string: "optionalconnectionstring" # Optional connection string  
azure_storage_bucket: "mycontainer"              # Azure Blob Storage container name _(required)_  
allow_host_access: true                          # Allow host environment access
```

### BigQuery
```yaml
type: connector                                  # Must be `connector` (required)
driver: bigquery                                 # Must be `bigquery` _(required)_

google_application_credentials: "credentialjsonstring"     # Google Cloud service account JSON  
project_id: "my-project-id"                      # Google Cloud project ID  
allow_host_access: true                          # Allow host environment access _(default: true)_
```

### ClickHouse
```yaml
type: connector                                  # Must be `connector` (required)
driver: clickhouse                               # Must be `clickhouse` _(required)_

managed: false                                   # Enable automatic provisioning _(default: false)_  
mode: "read"                                     # Operation mode: `read` or `readwrite` _(default: read)_  
dsn: "clickhouse://user:password@localhost:9000/database" # ClickHouse connection DSN  
username: "default"                              # Username for authentication  
password: "mypassword"                           # Password for authentication  
host: "localhost"                                # ClickHouse instance hostname  
port: 9000                                       # ClickHouse instance port  
database: "default"                              # ClickHouse database name  
ssl: false                                       # Enable SSL connection  
cluster: "mycluster"                             # Cluster name for distributed queries  
log_queries: false                               # Log raw SQL queries  
settings_override: "max_memory_usage=1000000000" # Override default query settings  
embed_port: 0                                    # Port for local ClickHouse (0 for random)  
can_scale_to_zero: false                         # Database can scale to zero  
max_open_conns: 10                               # Maximum open connections  
max_idle_conns: 5                                # Maximum idle connections  
dial_timeout: "30s"                              # Connection dial timeout  
conn_max_lifetime: "1h"                          # Maximum connection reuse time  
read_timeout: "30s"                              # Maximum read timeout
```

### Druid
```yaml
type: connector                                  # Must be `connector` (required)
driver: druid                                    # Must be `druid` _(required)_

dsn: "http://localhost:8082"                     # Druid connection DSN _(required)_  
username: "admin"                                # Username for authentication  
password: "mypassword"                           # Password for authentication  
host: "localhost"                                # Druid coordinator/broker hostname  
port: 8082                                       # Druid service port  
ssl: false                                       # Enable SSL connection  
log_queries: false                               # Log raw SQL queries  
max_open_conns: 10                               # Maximum open connections (0=default, -1=unlimited)  
skip_version_check: false                        # Skip Druid version compatibility check
```

### DuckDB
```yaml
type: connector                                  # Must be `connector` (required)
driver: duckdb                                   # Must be `duckdb` _(required)_

pool_size: 4                                     # Number of concurrent connections and queries  
allow_host_access: true                          # Allow local environment and file system access  
cpu: 4                                           # Number of CPU cores available to database  
memory_limit_gb: 8                               # Memory in GB available to database  
read_write_ratio: 0.8                            # Resource allocation ratio for read database  
init_sql: "SET memory_limit='8GB'"               # SQL executed during database initialization  
conn_init_sql: "SET timezone='UTC'"              # SQL executed when new connection is initialized  
secrets: "s3,gcs"                                # Comma-separated list of connector names for temporary secrets  
log_queries: false                               # Log raw SQL queries through OLAP
```

### GCS
```yaml
type: connector                                  # Must be `connector` (required)
driver: gcs                                      # Must be `gcs` _(required)_

google_application_credentials: "credentialjsonstring" # Google Cloud credentials JSON string  
bucket: "my-bucket"                              # GCS bucket name _(required)_  
allow_host_access: true                          # Allow host environment access  
key_id: "myaccesskey"                            # Optional S3-compatible Key ID  
secret: "mysecret"                               # Optional S3-compatible Secret
```

### Google Sheets
```yaml
type: model                                      # Slightly different than the others `model`
connector: "duckdb"                              # connector will use default DuckDB engine
sql: "select * from read_csv_auto('https://docs.google.com/spreadsheets/d/<SPREADSHEET_ID>/export?format=csv&gid=<SHEET_ID>', normalize_names=True)"                           # Fill in the parameters with your public Google Sheet.
```

### HTTPS
```yaml
type: connector                                  # Must be `connector` (required)
driver: https                                    # Must be `https` _(required)_

path: "https://example.com/data.csv"             # Full HTTPS URI to fetch data from _(required)_  
headers: "Authorization: Bearer token"           # HTTP headers to include in the request
```

### MotherDuck

```yaml
---
type: connector                                  # Must be `connector` (required)
driver: duckdb                                   # Must be `duckdb` _(required)_


path: "md:my_db"                                # Path to your MD database

init_sql: |                                     # SQL executed during database initialization.
  INSTALL 'motherduck';                         # Install motherduck extension
  LOAD 'motherduck';                            # Load the extensions
  SET motherduck_token= '{{ .env.connector.motherduck.access_token }}' # Define the motherduck token
```

### MySQL
```yaml
type: connector                                  # Must be `connector` (required)
driver: mysql                                    # Must be `mysql` _(required)_

dsn: "user:password@tcp(localhost:3306)/database" # MySQL connection DSN _(required)_
```

### Pinot
```yaml
type: connector                                  # Must be `connector` (required)
driver: pinot                                    # Must be `pinot` _(required)_

dsn: "http://localhost:8099"                     # Pinot connection DSN _(required)_  
username: "admin"                                # Username for authentication  
password: "mypassword"                           # Password for authentication  
broker_host: "localhost"                         # Pinot broker hostname _(required)_  
broker_port: 8099                                # Pinot broker port  
controller_host: "localhost"                     # Pinot controller hostname _(required)_  
controller_port: 9000                            # Pinot controller port  
ssl: false                                       # Enable SSL connection  
log_queries: false                               # Log raw SQL queries  
max_open_conns: 10                               # Maximum open connections
```

### PostgreSQL
```yaml
type: connector                                  # Must be `connector` (required)
driver: postgres                                 # Must be `postgres` _(required)_

dsn: "postgres://user:password@localhost:5432/database" # PostgreSQL connection DSN _(required)_
```

### Redshift
```yaml
type: connector                                  # Must be `connector` (required)
driver: redshift                                 # Must be `redshift` _(required)_

aws_access_key_id: "myawsaccesskey"              # AWS Access Key ID _(required)_  
aws_secret_access_key: "myawssecretkey"          # AWS Secret Access Key _(required)_  
aws_access_token: "mytemporarytoken"             # AWS Session Token for temporary credentials  
region: "us-east-1"                              # AWS region  
database: "my_database"                          # Redshift database name _(required)_  
workgroup: "my-workgroup"                        # Workgroup name for Redshift Serverless  
cluster_identifier: "my-cluster"                 # Cluster identifier for provisioned clusters
```

### S3
```yaml
type: connector                                  # Must be `connector` (required)
driver: s3                                       # Must be `s3` _(required)_

aws_access_key_id: "myawsaccesskey"              # AWS Access Key ID  
aws_secret_access_key: "myawssecretkey"          # AWS Secret Access Key  
aws_access_token: "mytemporarytoken"             # AWS session token for temporary credentials  
bucket: "my-bucket"                              # S3 bucket name _(required)_  
endpoint: "https://s3.amazonaws.com"             # Custom endpoint for S3-compatible storage  
region: "us-east-1"                              # AWS region of the S3 bucket  
allow_host_access: true                          # Allow host environment access  
retain_files: false                              # Retain intermediate files after processing
```

### Salesforce
```yaml
type: connector                                  # Must be `connector` (required)
driver: salesforce                               # Must be `salesforce` _(required)_

username: "user@example.com"                     # Salesforce account username _(required)_  
password: "mypassword"                           # Salesforce account password  
key: "mysecuritykey"                             # Authentication key  
endpoint: "https://login.salesforce.com"         # Salesforce API endpoint URL _(required)_  
client_id: "myclientid"                          # Client ID for OAuth authentication
```

### Slack
```yaml
type: connector                                  # Must be `connector` (required)
driver: slack                                    # Must be `slack` _(required)_

bot_token: "xoxb-myslackbottoken"                # Bot token for Slack API authentication _(required)_
```

### Snowflake
```yaml
type: connector                                  # Must be `connector` (required)
driver: snowflake                                # Must be `snowflake` _(required)_

dsn: "user:password@account/database/schema?warehouse=warehouse&role=role" # Snowflake connection DSN _(required if not using individual parameters)_  
account: "myaccount"                             # Snowflake account identifier _(required if not using dsn)_  
user: "myuser"                                   # Snowflake username _(required if not using dsn)_  
password: "mypassword"                           # Snowflake password _(required if not using dsn and privateKey)_  
database: "mydatabase"                           # Snowflake database name _(required if not using dsn)_  
schema: "myschema"                               # Snowflake schema name  
warehouse: "mywarehouse"                         # Snowflake warehouse name  
role: "myrole"                                   # Snowflake role name  
authenticator: "SNOWFLAKE_JWT"                   # Authentication method (e.g., 'SNOWFLAKE_JWT')  
privateKey: "myprivatekey"                       # RSA private key for JWT authentication _(required if not using password)_  
parallel_fetch_limit: 5                          # Maximum concurrent fetches during query execution
```

### SQLite
```yaml
type: connector                                  # Must be `connector` (required)
driver: sqlite                                   # Must be `sqlite` _(required)_

dsn: "./data/database.db"                        # SQLite connection DSN _(required)_
```


### AI
```yaml
type: connector                                  # Must be `connector` (required)
driver: openai                                   # Must be `openai` _(required)_

api_key: '{{ .env.openai_api_key }}'             # Openai api key _(required)_
```
