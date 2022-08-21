import {DatabaseActions} from "./DatabaseActions";
import {guidGenerator} from "$lib/util/guid";
import type {DatabaseMetadata} from "$common/database-service/DatabaseMetadata";
import * as prql from "prql-js/dist/node/";

export class DatabaseTableActions extends DatabaseActions {
    public async createViewOfQuery(
        metadata: DatabaseMetadata,
        tableName: string,
        query: string
    ): Promise<void> {
        let prql_query = query;
        if (query.startsWith('from')) {
            try {
                const ptos = query.replace('\\n', '|').replace('||', '|');
                prql_query = prql.to_sql(ptos);
            } catch (err) {
                console.log('prql_err --> createViewOfQuery --> ', err);
                return Promise.reject(404);
            }
        }
        const q = `-- wrapQueryAsTemporaryView
            CREATE OR REPLACE TEMPORARY VIEW ${tableName} AS (${prql_query});`;
        console.log(`'query --> ${q}`);
        await this.databaseClient.execute(q);
    }

    public async getFirstNOfTable(
        metadata: DatabaseMetadata,
        tableName: string,
        n = 1
    ): Promise<unknown[]> {
        // FIXME: sort out the type here
        try {
            return await this.databaseClient.execute(
                `SELECT * from '${tableName}' LIMIT ${n};`
            );
        } catch (err) {
            throw Error(err);
        }
    }

    public async getCardinalityOfTable(
        metadata: DatabaseMetadata,
        tableName: string
    ): Promise<number> {
        const [cardinality] = await this.databaseClient.execute<{ count: number }>(
            `select count(*) as count FROM '${tableName}';`
        );
        return cardinality.count;
    }

    public async getProfileColumns(
        metadata: DatabaseMetadata,
        tableName: string
    ): Promise<unknown[]> {
        const guid = guidGenerator().replace(/-/g, "_");
        await this.databaseClient.execute(`-- parquetToDBTypes
	        CREATE TEMP TABLE tbl_${guid} AS (
                SELECT * from '${tableName}' LIMIT 1
            );
	    `);
        const tableDef = await this.databaseClient.execute(`-- parquetToDBTypes
            PRAGMA table_info(tbl_${guid});`);
        await this.databaseClient.execute(`DROP TABLE tbl_${guid};`);
        return tableDef;
    }

    public async validateQuery(
        metadata: DatabaseMetadata,
        query: string
    ): Promise<void> {

        if (query.trim().toLowerCase().startsWith('from')) {
            try {
                const sql = prql.to_sql(query);
                return this.databaseClient.prepare(sql);
            }
            catch (err) {
                throw new Error(
                    err.toString()
                        .replace(/\n/g, "<br>")
                        .replace(/\t/g, "&nbsp;&nbsp;&nbsp;&nbsp;")
                        .replace(/\s/g, "&nbsp;")
                        .replace('"Error: ', '')
                );
            }
        } else {
            return this.databaseClient.prepare(query);
        }

    }

    public async renameTable(
        metadata: DatabaseMetadata,
        tableName: string,
        newTableName: string
    ): Promise<void> {
        await this.databaseClient.execute(
            `ALTER TABLE ${tableName} RENAME TO ${newTableName};`
        );
    }

    public async dropTable(
        metadata: DatabaseMetadata,
        tableName: string
    ): Promise<void> {
        await this.databaseClient.execute(`DROP TABLE ${tableName}`);
    }
}
