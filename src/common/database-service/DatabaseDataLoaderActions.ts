import {DatabaseActions} from "./DatabaseActions";
import { existsSync, mkdirSync } from "fs";
import type {DatabaseMetadata} from "$common/database-service/DatabaseMetadata";
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";

/**
 * Abstraction around loading data into duck db.
 * WASM version will change how the file is read.
 */
export class DatabaseDataLoaderActions extends DatabaseActions {
    private async createTableWithQuery(metadata: DatabaseMetadata, tableName: string, query: string): Promise<any> {
        // check if table exists.
        const tables = await this.databaseClient.execute('SHOW TABLES');
        const tableIsPresent = tables.some(table => table.name === tableName);

        // if table does exist, let's put it in a temporary place during import.
        // if there is an error, we should rename the temp table back to its original one.
        if (tableIsPresent) {
            await this.databaseClient.execute(`ALTER TABLE ${tableName} RENAME TO ${tableName}___;`);    
        }
        let outcome;
        try {
            outcome = await this.databaseClient.execute(`CREATE TABLE ${tableName} AS ${query};`);
            if (tableIsPresent) {
                await this.databaseClient.execute(`DROP TABLE IF EXISTS ${tableName}___;`);
            }
        } catch (error) {
            if (tableIsPresent) {
                await this.databaseClient.execute(`ALTER TABLE ${tableName}___ RENAME TO ${tableName};`);
            }
            return ActionResponseFactory.getEntityError(error);
        }
        return outcome;
    }
    
    public async importParquetFile(metadata: DatabaseMetadata, parquetFile: string, tableName: string): Promise<any> {
        return this.createTableWithQuery(metadata, tableName, `SELECT * FROM '${parquetFile}'`)
    }
    
    public async importCSVFile(metadata: DatabaseMetadata, csvFile: string,
                               tableName: string, delimiter: string): Promise<void> {
        return this.createTableWithQuery(metadata, tableName, `SELECT * FROM 
        read_csv_auto('${csvFile}', header=true ${delimiter ? `,delim='${delimiter}'`: ""})`);
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
