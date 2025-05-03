---
note: GENERATED. DO NOT EDIT.
title: Model YAML
sidebar_position: 8
---
## Model YAML

Type: `object`

## Properties:
#### All of the following:
- Part 1:
  ## type

  Type: `object`

  ## Properties:

  - **type** _(required)_:
    Type: `%!s(<nil>)`

- Part 2:
  ## common_properties

  Type: `object`

  ## Properties:

  - **name**:
    Name is usually inferred from the filename, but can be specified manually.

    Type: `string`


  - **namespace**:
    Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`.

    Type: `string`


  - **refs**:
    List of resource references, each as a string or map.

    Type: `array`

    #### Array Items:
      Type: `%!s(<nil>)`

      #### One of the following:
      - Option 1:
        A string reference like 'resource-name' or 'Kind/resource-name'.

        Type: `string`

      - Option 2:
        An object reference with at least a 'name' and 'type'.

        Type: `object`

        ## Properties:

        - **name** _(required)_:
          Type: `string`


        - **type**:
          Type: `string`


  - **version**:
    Version of the parser to use for this file. Enables backwards compatibility for breaking changes.

    Type: `integer`

- Part 3:
  ## connector_properties

  Type: `object`

  ## Properties:
  #### One of the following:
  - Option 1:
    ## athena

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## athena_properties

      Type: `object`

      ## Properties:

      - **aws_access_token**:
        Optional AWS session token for temporary credentials

        Type: `string`


      - **aws_secret_access_key**:
        AWS Secret Access Key for Athena access

        Type: `string`


      - **external_id**:
        Optional External ID for assuming a role

        Type: `string`


      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **role_arn**:
        Optional AWS Role ARN to assume when accessing Athena

        Type: `string`


      - **role_session_name**:
        Optional Session name when assuming the role

        Type: `string`


      - **allow_host_access**:
        Allow access to host environment configuration

        Type: `boolean`


      - **aws_access_key_id**:
        AWS Access Key ID for Athena access

        Type: `string`

  - Option 2:
    ## azure

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## azure_properties

      Type: `object`

      ## Properties:

      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **allow_host_access**:
        Allow access to host environment configuration

        Type: `boolean`


      - **azure_storage_account**:
        Azure storage account name

        Type: `string`


      - **azure_storage_connection_string**:
        Optional azure connection string for storage account

        Type: `string`


      - **azure_storage_key**:
        Azure storage access key

        Type: `string`


      - **azure_storage_sas_token**:
        Optional azure SAS token for authentication

        Type: `string`

  - Option 3:
    ## bigquery

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## bigquery_properties

      Type: `object`

      ## Properties:

      - **allow_host_access**:
        Allow access to host environment configuration

        Type: `boolean`


      - **google_application_credentials**:
        Path to the Google Cloud credentials JSON file

        Type: `string`


      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:
  - Option 4:
    ## clickhouse

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## clickhouse_properties

      Type: `object`

      ## Properties:

      - **can_scale_to_zero**:
        Indicates if the database can scale to zero

        Type: `boolean`


      - **cluster**:
        Cluster name, required for running distributed queries

        Type: `string`


      - **dsn**:
        DSN(Data Source Name) for the ClickHouse connection

        Type: `string`


      - **password**:
        Password for authentication

        Type: `string`


      - **read_timeout**:
        Maximum time for a connection to read data

        Type: `string`


      - **settings_override**:
        override the default settings used in queries. example `readonly = 1, session_timezone = 'UTC'`

        Type: `string`


      - **ssl**:
        Indicates whether a secured SSL connection is required

        Type: `boolean`


      - **dial_timeout**:
        Timeout for dialing the ClickHouse server

        Type: `string`


      - **host**:
        Host where the ClickHouse instance is running

        Type: `string`


      - **log_queries**:
        Controls whether to log raw SQL queries

        Type: `boolean`


      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **max_idle_conns**:
        Maximum number of idle connections in the pool

        Type: `integer`


      - **database**:
        Name of the ClickHouse database within the cluster

        Type: `string`


      - **embed_port**:
        Port to run ClickHouse locally (0 for random port)

        Type: `integer`


      - **port**:
        Port where the ClickHouse instance is accessible

        Type: `integer`


      - **username**:
        Username for authentication

        Type: `string`


      - **conn_max_lifetime**:
        Maximum time a connection may be reused

        Type: `string`


      - **max_open_conns**:
        Maximum number of open connections to the database

        Type: `integer`

  - Option 5:
    ## druid

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## druid_properties

      Type: `object`

      ## Properties:

      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **max_open_conns**:
        Maximum number of open database connections (0 = default, -1 = unlimited)

        Type: `integer`


      - **password**:
        Password for authenticating with Druid

        Type: `string`


      - **port**:
        Port number of the Druid service

        Type: `integer`


      - **dsn**:
        Data Source Name (DSN) for connecting to Druid

        Type: `string`


      - **host**:
        Hostname of the Druid coordinator or broker

        Type: `string`


      - **log_queries**:
        Log raw SQL queries sent to Druid

        Type: `boolean`


      - **skip_version_check**:
        Skip checking Druid version compatibility

        Type: `boolean`


      - **ssl**:
        Enable SSL for secure connection

        Type: `boolean`


      - **username**:
        Username for authenticating with Druid

        Type: `string`

  - Option 6:
    ## duckdb

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## duckdb_properties

      Type: `object`

      ## Properties:

      - **init_sql**:
        SQL to run when initializing a new connection, after extensions and defaults

        Type: `string`


      - **log_queries**:
        Whether to log raw SQL queries executed through OLAP

        Type: `boolean`


      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **path**:
        File path to the DuckDB database file

        Type: `string`


      - **pool_size**:
        Number of concurrent connections and queries allowed

        Type: `integer`


      - **read_write_ratio**:
        Ratio of resources allocated to the read database; used to divide CPU and memory

        Type: `number`


      - **boot_queries**:
        SQL to run when initializing a new connection, before extensions and defaults

        Type: `string`


      - **cpu**:
        Number of CPU cores available to the database

        Type: `integer`


      - **memory_limit_gb**:
        Amount of memory in GB available to the database

        Type: `integer`


      - **allow_host_access**:
        Whether access to the local environment and file system is allowed

        Type: `boolean`

  - Option 7:
    ## gcs

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## gcs_properties

      Type: `object`

      ## Properties:

      - **google_application_credentials**:
        Google Cloud credentials JSON string

        Type: `string`


      - **key_id**:
        Optional S3-compatible Key ID when used in compatibility mode

        Type: `string`


      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **secret**:
        Optional S3-compatible Secret when used in compatibility mode

        Type: `string`


      - **allow_host_access**:
        Allow access to host environment configuration

        Type: `boolean`

  - Option 8:
    ## https

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## https_properties

      Type: `object`

      ## Properties:

      - **headers**:
        HTTP headers to include in the request

        Type: `object`

        ## Properties:

      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **path**:
        The full HTTPS URI to fetch data from

        Type: `string`

  - Option 9:
    ## local_file

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## local_file_properties

      Type: `object`

      ## Properties:

      - **allow_host_access**:
        Flag to indicate if access to host-level file paths is permitted

        Type: `boolean`


      - **dsn**:
        Data Source Name (DSN) indicating the file path or location of the local file

        Type: `string`


      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:
  - Option 10:
    ## motherduck

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## motherduck_properties

      Type: `object`

      ## Properties:

      - **dsn**:
        Data Source Name (DSN) specifying the MotherDuck connection endpoint

        Type: `string`


      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **token**:
        Authentication token for accessing MotherDuck (secret)

        Type: `string`

  - Option 11:
    ## mysql

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## mysql_properties

      Type: `object`

      ## Properties:

      - **dsn**:
        DSN(Data Source Name) for the mysql connection

        Type: `string`


      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:
  - Option 12:
    ## pinot

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## pinot_properties

      Type: `object`

      ## Properties:

      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **max_open_conns**:
        Maximum number of open connections to the Pinot database

        Type: `integer`


      - **ssl**:
        Enable SSL connection to Pinot

        Type: `boolean`


      - **broker_host**:
        Hostname of the Pinot broker

        Type: `string`


      - **password**:
        Password for authenticating with Pinot

        Type: `string`


      - **username**:
        Username for authenticating with Pinot

        Type: `string`


      - **broker_port**:
        Port number for the Pinot broker

        Type: `integer`


      - **controller_host**:
        Hostname of the Pinot controller

        Type: `string`


      - **controller_port**:
        Port number for the Pinot controller

        Type: `integer`


      - **dsn**:
        DSN(Data Source Name) for the Pinot connection

        Type: `string`


      - **log_queries**:
        Log raw SQL queries executed through Pinot

        Type: `boolean`

  - Option 13:
    ## postgres

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## postgres_properties

      Type: `object`

      ## Properties:

      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **dsn**:
        DSN(Data Source Name) for the postgres connection

        Type: `string`

  - Option 14:
    ## redshift

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## redshift_properties

      Type: `object`

      ## Properties:

      - **allow_host_access**:
        Allow access to host environment configuration

        Type: `boolean`


      - **aws_access_key_id**:
        AWS access key ID for authentication

        Type: `string`


      - **aws_access_token**:
        AWS session token for temporary credentials (optional)

        Type: `string`


      - **aws_secret_access_key**:
        AWS secret access key for authentication

        Type: `string`


      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:
  - Option 15:
    ## s3

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## s3_properties

      Type: `object`

      ## Properties:

      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **region**:
        AWS region of the S3 bucket

        Type: `string`


      - **retain_files**:
        Whether to retain intermediate files after processing

        Type: `boolean`


      - **allow_host_access**:
        Allow access to host environment configuration

        Type: `boolean`


      - **aws_access_key_id**:
        AWS Access Key ID used for authentication

        Type: `string`


      - **aws_access_token**:
        Optional AWS session token for temporary credentials

        Type: `string`


      - **aws_secret_access_key**:
        AWS Secret Access Key used for authentication

        Type: `string`


      - **endpoint**:
        Optional custom endpoint URL for S3-compatible storage

        Type: `string`

  - Option 16:
    ## salesforce

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## salesforce_properties

      Type: `object`

      ## Properties:

      - **endpoint**:
        Salesforce API endpoint URL

        Type: `string`


      - **key**:
        Authentication key for Salesforce (secret)

        Type: `string`


      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **password**:
        Salesforce account password (secret)

        Type: `string`


      - **username**:
        Salesforce account username

        Type: `string`


      - **client_id**:
        Client ID used for Salesforce OAuth authentication

        Type: `string`

  - Option 17:
    ## snowflake

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## snowflake_properties

      Type: `object`

      ## Properties:

      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:

      - **parallel_fetch_limit**:
        Maximum number of concurrent fetches during query execution

        Type: `integer`


      - **dsn**:
        DSN (Data Source Name) for the Snowflake connection

        Type: `string`

  - Option 18:
    ## sqlite

    Type: `object`

    ## Properties:
    #### All of the following:
    - Part 1:
      Type: `object`

      ## Properties:

      - **connector** _(required)_:
        Type: `%!s(<nil>)`

    - Part 2:
      ## sqlite_properties

      Type: `object`

      ## Properties:

      - **dsn**:
        DSN(Data Source Name) for the sqlite connection

        Type: `string`


      - **managed**:
        Boolean or map of properties

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `boolean`

        - Option 2:
          Type: `object`

          ## Properties:
- Part 4:
  ## model_properties

  Type: `object`

  ## Properties:

  - **incremental**:
    whether incremental modeling is required (optional)

    Type: `boolean`


  - **materialize**:
    models will be materialized in olap

    Type: `boolean`


  - **output**:
    to define the properties of output

    Type: `object`

    ## Properties:

    - **connector** _(required)_:
      Refers to the connector type for the output table

      Type: `string`


  - **partitions_concurrency**:
    Refers to the number of concurrent partitions that can be read at the same time (optional).

    Type: `integer`


  - **refresh**:
    Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying model data 

    Type: `object`

    ## Properties:

    - **every**:
      Run at a fixed interval using a Go duration string (e.g., '1h', '30m', '24h'). See: https://pkg.go.dev/time#ParseDuration

      Type: `string`


    - **ref_update**:
      If true, allows the resource to run when a dependency updates.

      Type: `boolean`


    - **run_in_dev**:
      If true, allows the schedule to run in development mode.

      Type: `boolean`


    - **time_zone**:
      Time zone to interpret the schedule in (e.g., 'UTC', 'America/Los_Angeles').

      Type: `string`


    - **cron**:
      A cron expression that defines the execution schedule

      Type: `string`


    - **disable**:
      If true, disables the resource without deleting it.

      Type: `boolean`


  - **state**:
    Refers to the explicitly defined state of your model, cannot be used with partitions (optional)

    Type: `object`

    ## Properties:
    #### One of the following:
    - Option 1:
      ## sql

      Type: `%!s(<nil>)`

    - Option 2:
      ## metrics_sql

      Type: `%!s(<nil>)`

    - Option 3:
      ## api

      Type: `%!s(<nil>)`

    - Option 4:
      ## glob

      Type: `%!s(<nil>)`

    - Option 5:
      ## resource_status

      Type: `%!s(<nil>)`


  - **timeout**:
    The maximum time to wait for model ingestion

    Type: `string`


  - **partitions**:
    Refers to the how your data is partitioned, cannot be used with state. (optional)

    Type: `object`

    ## Properties:
    #### One of the following:
    - Option 1:
      ## sql

      Type: `%!s(<nil>)`

    - Option 2:
      ## metrics_sql

      Type: `%!s(<nil>)`

    - Option 3:
      ## api

      Type: `%!s(<nil>)`

    - Option 4:
      ## glob

      Type: `%!s(<nil>)`

    - Option 5:
      ## resource_status

      Type: `%!s(<nil>)`


  - **partitions_watermark**:
    Refers to a customizable timestamp that can be set to check if an object has been updated (optional).

    Type: `string`


  - **sql**:
    Raw SQL query to run against source

    Type: `string`


  - **stage**:
    in the case of staging models, where an input source does not support direct write to the output and a staging table is required

    Type: `object`

    ## Properties:

    - **connector** _(required)_:
      Refers to the connector type for the staging table

      Type: `string`

- Part 5:
  ## environment_overrides

  Type: `%!s(<nil>)`

