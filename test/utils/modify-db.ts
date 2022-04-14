import "../../src/moduleAlias";
import {DuckDBClient} from "$common/database-service/DuckDBClient";
import {DatabaseConfig} from "$common/config/DatabaseConfig";

(async () => {
    const duckDbClient = new DuckDBClient(new DatabaseConfig(
        {databaseName: process.argv[2]}));
    await duckDbClient.init();

    await duckDbClient.execute(`ALTER TABLE AdBids DROP domain`);
    await duckDbClient.execute(`ALTER TABLE Impressions RENAME country to r_country`);
})();
