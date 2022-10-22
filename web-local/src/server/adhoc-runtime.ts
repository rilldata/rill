import "../moduleAlias";
import { DatabaseConfig } from "@rilldata/web-local/common/config/DatabaseConfig";
import { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import { DuckDBClient } from "@rilldata/web-local/common/database-service/DuckDBClient";

/**
 * Use this file to test individual endpoints in runtime from outside.
 * Good to test APIs without UI
 *
 * Run "npm run dev -w runtime" to start runtime in a new terminal
 */
(async () => {
  const duckdb = DuckDBClient.getInstance(
    new RootConfig({
      database: new DatabaseConfig({
        spawnRuntime: false,
        spawnRuntimePort: 8081,
      }),
    })
  );
  await duckdb.init();
  await duckdb.execute("SET enable_profiling=QUERY_TREE");

  console.log(
    await duckdb.requestToInstance("migrate/single", {
      sql:
        "create source s3test with connector = 's3', path = 's3://datasets-epg/yammerevents.parquet'," +
        // WARNING: DO NOT CHECK IN KEY AND SECRET
        "'aws.region' = 'us-east-1', 'aws.access.key' = '', 'aws.access.secret' = ''  ",
    })
  );

  console.log(
    await duckdb.requestToInstance("query/direct", {
      sql: "select * from s3test limit 5",
    })
  );
})();
