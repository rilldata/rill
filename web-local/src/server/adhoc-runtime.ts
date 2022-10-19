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

  console.log(
    await duckdb.requestToInstance("migrate/single", {
      sql:
        "create source AdBidsS3 with connector = 's3', path = 's3://rill-developer.rilldata.io/AdBids.csv'," +
        "'aws.region' = 'us-east-1'",
    })
  );
  console.log(
    await duckdb.requestToInstance("query/direct", {
      sql: "select * from AdBidsS3 limit 5",
    })
  );

  console.log(
    await duckdb.requestToInstance("migrate/single", {
      sql:
        "create source AdBidsGS with connector = 'gcs', path = 's3://scratch.rilldata.com/rill-developer/AdBids.csv'," +
        "'gcp.region' = 'us-east-1'",
    })
  );
  console.log(
    await duckdb.requestToInstance("query/direct", {
      sql: "select * from AdBidsGS limit 5",
    })
  );
})();
