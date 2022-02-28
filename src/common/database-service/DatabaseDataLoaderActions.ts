import { DatabaseActions } from "./DatabaseActions";
import { existsSync, mkdirSync } from "fs";
import type { DatabaseMetadata } from "$common/database-service/DatabaseMetadata";

/**
 * Abstraction around loading data into duck db.
 * WASM version will change how the file is read.
 */
export class DatabaseDataLoaderActions extends DatabaseActions {
    public async importParquetFile(metadata: DatabaseMetadata, parquetFile: string, tableName: string): Promise<any> {
        await this.databaseClient.execute(`DROP TABLE IF EXISTS ${tableName};`);
        return await this.databaseClient.execute(`CREATE TABLE ${tableName} AS SELECT * FROM '${parquetFile}';`);
    }

    public async importCSVFile(metadata: DatabaseMetadata, csvFile: string,
                               tableName: string, delimiter: string): Promise<void> {
        await this.databaseClient.execute(`DROP TABLE IF EXISTS ${tableName};`);
        return await this.databaseClient.execute(`CREATE TABLE ${tableName} AS SELECT * FROM ` +
            `read_csv_auto('${csvFile}', header=true ${delimiter ? `,delim='${delimiter}'` : ""});`);
    }

    public async exportToParquet(metadata: DatabaseMetadata, query: string, exportFile: string): Promise<any> {
        if (!existsSync(this.databaseConfig.exportFolder)) {
            mkdirSync(this.databaseConfig.exportFolder);
        }
        const exportPath = `${this.databaseConfig.exportFolder}/${exportFile}`;
        const exportQuery = `COPY (${query}) TO '${exportPath}' (FORMAT 'parquet')`;
        await this.databaseClient.execute(exportQuery);
        return exportPath;
    }
}
