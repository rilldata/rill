import {DatabaseActions} from "./DatabaseActions";
import { existsSync, mkdirSync } from "fs";
import type {DatabaseMetadata} from "$common/database-service/DatabaseMetadata";

/**
 * Abstraction around loading data into duck db.
 * WASM version will change how the file is read.
 */
export class DatabaseDataLoaderActions extends DatabaseActions {
    private async importWithQuery(metadata: DatabaseMetadata, tableName: string, query: string): Promise<any> {
        // check if table exists.
        const tables = await this.databaseClient.execute('SHOW TABLES');
        const tableIsPresent = tables.some(table => table.name = tableName);
        // if table does exist, let's put it in a temporary place during import.
        // if there is an error, we should rename the temp table back to its original one.
        if (tableIsPresent) {
            await this.databaseClient.execute(`ALTER TABLE ${tableName} RENAME TO ${tableName}___;`);    
        }
        try {
            const outcome = await this.databaseClient.execute(query);
            // we are finished with that temporary table, so let's drop it.
            if (tableIsPresent) {
                await this.databaseClient.execute(`DROP TABLE IF EXISTS ${tableName}___;`);
            }
            return outcome;
        } catch (err) {
            // let's make sure the table is renamed back to the original if there's an error.
            if (tableIsPresent) {
                await this.databaseClient.execute(`ALTER TABLE ${tableName}___ RENAME TO ${tableName};`);
            }
            throw err;
        }
    }

    public async importParquetFile(metadata: DatabaseMetadata, parquetFile: string, tableName: string): Promise<any> {
        return this.importWithQuery(metadata, tableName, `CREATE TABLE ${tableName} AS SELECT * FROM '${parquetFile}';`)
    }

    public async importCSVFile(metadata: DatabaseMetadata, csvFile: string,
                               tableName: string, delimiter: string): Promise<void> {
        return this.importWithQuery(metadata, tableName, `CREATE TABLE ${tableName} AS SELECT * FROM 
        read_csv_auto('${csvFile}', header=true ${delimiter ? `,delim='${delimiter}'`: ""});`);
    }

    public async getDestinationSize(metadata: DatabaseMetadata, path: string): Promise<number> {
        // Being worked on to handle this in a better way.
        // if (existsSync(path)) {
        //     const size = await this.databaseClient.all(`SELECT total_compressed_size from parquet_metadata('${path}')`) as any[];
        //     return size.reduce((acc: number, v: Record<string, any>) => acc + v.total_compressed_size, 0);
        // }
        return undefined;
    }

    public async exportToParquet(metadata: DatabaseMetadata, query: string, exportFile: string): Promise<any> {
        return this.exportToFile(query, exportFile, "FORMAT PARQUET");
    }

    public async exportToCsv(metadata: DatabaseMetadata, query: string, exportFile: string): Promise<any> {
        return this.exportToFile(query, exportFile, "FORMAT CSV, HEADER")
    }

    private async exportToFile(query: string, exportFile: string, exportString: string): Promise<any> {
        if (!existsSync(this.databaseConfig.exportFolder)) {
            mkdirSync(this.databaseConfig.exportFolder);
        }
        const exportPath = `${this.databaseConfig.exportFolder}/${exportFile}`;
        const exportQuery = `COPY (${query}) TO '${exportPath}' (${exportString})`;
        await this.databaseClient.execute(exportQuery);
        return exportPath;
    }
}
