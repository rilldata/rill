import {DuckDBAPI} from "./DuckDBAPI";
import {sanitizeQuery} from "$lib/util/sanitize-query";
import fs from "fs";

export class DuckDBTableAPI extends DuckDBAPI {
    public async materializeTable(tableName: string, query: string): Promise<void> {
        await this.duckDBClient.all(`-- wrapQueryAsTemporaryView
DROP TABLE IF EXISTS ${tableName}`);
        const sanitizedQuery = sanitizeQuery(query);
        await this.duckDBClient.all(`-- wrapQueryAsTemporaryView
CREATE TABLE ${tableName} AS ${sanitizedQuery}`);
    }

    public async createViewOfQuery(tableName: string, query: string): Promise<void> {
        await this.duckDBClient.all(`-- wrapQueryAsTemporaryView
DROP VIEW IF EXISTS ${tableName}`);
        const sanitizedQuery = sanitizeQuery(query);
        await this.duckDBClient.all(`-- wrapQueryAsTemporaryView
CREATE TEMP VIEW ${tableName} AS ${sanitizedQuery}`);
    }

    public async getFirstN(tableName: string, n = 1): Promise<any[]> {
        // FIXME: sort out the type here
        try {
            return await this.duckDBClient.all(`SELECT * from '${tableName}' LIMIT ${n};`);
        } catch (err) {
            throw Error(err);
        }
    }

    public async getCardinality(tableName: string): Promise<number> {
        const [cardinality] = await this.duckDBClient.all(`select count(*) as count FROM '${tableName}';`);
        return cardinality.count;
    }

    public async getDestinationSize(path: string): Promise<number> {
        if (fs.existsSync(path)) {
            const size = await this.duckDBClient.all(`SELECT total_compressed_size from parquet_metadata('${path}')`) as any[];
            return size.reduce((acc: number, v: Record<string, any>) => acc + v.total_compressed_size, 0);
        }
        return undefined;
    }

    public async getProfileColumns(tableName: string): Promise<any[]> {
        return await this.duckDBClient.all(`PRAGMA table_info('${tableName}');`);
    }
}
