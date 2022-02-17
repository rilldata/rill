import {DatabaseActions} from "./DatabaseActions";
import { existsSync, mkdirSync } from "fs";
import type {DatabaseMetadata} from "$common/database-service/DatabaseMetadata";

/**
 * Abstraction around loading data into duck db.
 * WASM version will change how the file is read.
 */
export class DatabaseDataLoaderActions extends DatabaseActions {
    public async importParquetFile(metadata: DatabaseMetadata, parquetFile: string, tableName: string): Promise<any> {
        await this.databaseClient.execute(`DROP TABLE IF EXISTS ${tableName};`);
        return await this.databaseClient.execute(`CREATE TABLE ${tableName} AS SELECT * FROM '${parquetFile}';`);
    }

    public async getDestinationSize(metadata: DatabaseMetadata, path: string): Promise<number> {
        if (existsSync(path)) {
            const size = await this.databaseClient.execute(`SELECT total_compressed_size from parquet_metadata('${path}')`) as any[];
            return size.reduce((acc: number, v: Record<string, any>) => acc + v.total_compressed_size, 0);
        }
        return undefined;
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
