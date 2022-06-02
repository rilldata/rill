import "../../src/moduleAlias";
import { DuckDBClient } from "$common/database-service/DuckDBClient";
import { DatabaseConfig } from "$common/config/DatabaseConfig";

(async () => {
  const duckDbClient = DuckDBClient.getInstance(
    new DatabaseConfig({ databaseName: process.argv[2] })
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
})();
