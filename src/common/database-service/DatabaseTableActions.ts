import { DatabaseActions } from "./DatabaseActions";
import { guidGenerator } from "$lib/util/guid";
import type { DatabaseMetadata } from "$common/database-service/DatabaseMetadata";
import type { ProfileColumn } from "$lib/types";
import { CATEGORICALS } from "$lib/duckdb-data-types"
import type { Table } from "$lib/components/table";

export class DatabaseTableActions extends DatabaseActions {
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

    /**
     * Estimate size (in bytes) of output table.
     *
     * @param {DatabaseMetadata} metadata - DB metadata.
     * @param {Table} table - Table metadata.
     * @param {Array<ProfileColumn>} - Column metadata from profiler.
     * @returns Estimated size of output table, in bytes.
     */
    public async getDestinationSize(metadata: DatabaseMetadata, table: Table, profileColumns: Array<ProfileColumn>): Promise<number> {
        const args = [];
        const sizes = [];

        for (const column of profileColumns) {
            if (CATEGORICALS.has(column.type)) {
                // If this is a categorical type such as VARCHAR, we will need to query to get a good estimate.
                args.push(`sum(bit_length(${column.name}))`);
            } else {
                // Otherwise, use the column type and row count (removing any nulls).

                // FIXME provide sizes for various types e.g. https://duckdb.org/docs/sql/data_types/numeric
                // FIXME this could probably go into `duckdb-data-types`
                enum SIZES {
                    TINYINT = 1, SMALLINT = 1, INTEGER = 1, BIGINT = 1, HUGEINT = 1, UTINYINT = 1,
                    USMALLINT = 1, UINTEGER = 1, UBIGINT = 1, INT16 = 2, INT32 = 4, INT64 = 8, INT128 = 16,
                    REAL = 4, DOUBLE = 8,
                    TIMESTAMP = 32
                };
                const profile = table.profile.find((a: { name: string; }) => a.name === column.name);
                let nullCount = 0;
                if (profile && profile.nullCount) {
                    nullCount = profile.nullCount;
                }
                sizes.push(SIZES[column.type] * (table.cardinality - nullCount));
            }
        }

        if (args.length > 0) {
            const [result] = await this.databaseClient.execute(`SELECT ${args} from ${table.tableName} USING SAMPLE 10000`);
            sizes.push(...Object.values(result));
        }
        const sum = sizes.reduce((a, b) => a + b, 0);

        // Return size in bytes.
        return sum / 8;
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
