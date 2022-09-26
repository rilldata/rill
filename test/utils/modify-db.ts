import "../../src/moduleAlias";
import { DatabaseConfig } from "$common/config/DatabaseConfig";
import { RootConfig } from "$common/config/RootConfig";
import { DuckDBClient } from "$common/database-service/DuckDBClient";

(async () => {
  const duckDbClient = DuckDBClient.getInstance(
    new RootConfig({
      database: new DatabaseConfig({
        databaseName: process.argv[2],
        spawnRuntime: false,
        runtimeUrl: `http://localhost:8081`,
      }),
    })
  );
  await duckDbClient.init();

  await duckDbClient.execute(`ALTER TABLE AdBids DROP domain`);
  await duckDbClient.execute(
    `ALTER TABLE Impressions RENAME country to r_country`
  );
  // create temporary tables and views. this will not be picked up
  await duckDbClient.execute(
    `CREATE TEMP TABLE AdBids_row as (select * from AdBids limit 1);`
  );
  await duckDbClient.execute(
    `CREATE OR REPLACE TEMPORARY VIEW FullTable AS (select * from AdBids b join Impressions i on b.id=i.id);`
  );

  await duckDbClient.destroy();
})();
