import {DatabaseActions} from "./DatabaseActions";
import fs from "fs";

/**
 * Abstraction around loading data into duck db.
 * WASM version will change how the file is read.
 */
export class DatabaseDataLoaderActions extends DatabaseActions {
    public async loadData(parquetFile: string, tableName: string): Promise<any> {
        return await this.dbClient.execute(`CREATE TABLE ${tableName} AS SELECT * FROM '${parquetFile}';`);
        // return await this.dbClient.execute(`SELECT * FROM parquet_schema('${parquetFile}');`);
    }

    public async getDestinationSize(path: string): Promise<number> {
        if (fs.existsSync(path)) {
            const size = await this.dbClient.execute(`SELECT total_compressed_size from parquet_metadata('${path}')`) as any[];
            return size.reduce((acc: number, v: Record<string, any>) => acc + v.total_compressed_size, 0);
        }
        return undefined;
    }
}
