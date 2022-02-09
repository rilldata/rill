import {DatabaseActions} from "./DatabaseActions";
import {guidGenerator} from "$lib/util/guid";
import type {DatabaseMetadata} from "$common/database-service/DatabaseMetadata";

export class DatabaseTableActions extends DatabaseActions {
    public async materializeTable(metadata: DatabaseMetadata, tableName: string, query: string): Promise<any> {
        await this.databaseClient.execute(`-- wrapQueryAsTemporaryView
            DROP TABLE IF EXISTS '${tableName}'`);
        return await this.databaseClient.execute(`-- wrapQueryAsTemporaryView
            CREATE TABLE '${tableName}' AS ${query}`);
    }

    public async createViewOfQuery(metadata: DatabaseMetadata, tableName: string, query: string): Promise<any> {
        await this.databaseClient.execute(`-- wrapQueryAsTemporaryView
            CREATE OR REPLACE TEMPORARY VIEW ${tableName} AS (${query});`);
    }

    public async getFirstNOfTable(metadata: DatabaseMetadata, tableName: string, n = 1): Promise<any[]> {
        // FIXME: sort out the type here
        try {
            return await this.databaseClient.execute(`SELECT * from '${tableName}' LIMIT ${n};`);
        } catch (err) {
            throw Error(err);
        }
    }

    public async getCardinalityOfTable(metadata: DatabaseMetadata, tableName: string): Promise<number> {
        const [cardinality] = await this.databaseClient.execute(`select count(*) as count FROM '${tableName}';`);
        return cardinality.count;
    }

    public async getProfileColumns(metadata: DatabaseMetadata, tableName: string): Promise<any[]> {
        const guid = guidGenerator().replace(/-/g, '_');
        await this.databaseClient.execute(`-- parquetToDBTypes
	        CREATE TEMP TABLE tbl_${guid} AS (
                SELECT * from '${tableName}' LIMIT 1
            );
	    `);
        const tableDef = await this.databaseClient.execute(`-- parquetToDBTypes
            PRAGMA table_info(tbl_${guid});`)
        await this.databaseClient.execute(`DROP TABLE tbl_${guid};`);
        return tableDef;
    }

    public async validateQuery(metadata: DatabaseMetadata, query: string): Promise<void> {
        return this.databaseClient.prepare(query);
    }
}
