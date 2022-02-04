import {DuckDBAPI} from "./DuckDBAPI";

/**
 * Abstraction around loading data into duck db.
 * WASM version will change how the file is read.
 */
export class DuckDBDataLoaderAPI extends DuckDBAPI {
    public async loadData(parquetFile: string, tableName: string): Promise<void> {
        await this.duckDBClient.all(`CREATE TABLE '${tableName}' AS SELECT * FROM '${parquetFile}';`);
    }
}
