import type { DatabaseMetadata } from "$common/database-service/DatabaseMetadata";
import type { ProfileColumn } from "$lib/types";
import { guidGenerator } from "$lib/util/guid";
import { DatabaseActions } from "./DatabaseActions";

export class DatabaseTableActions extends DatabaseActions {
  public async createViewOfQuery(
    metadata: DatabaseMetadata,
    tableName: string,
    query: string
  ): Promise<void> {
    await this.databaseClient.execute(`-- wrapQueryAsTemporaryView
            CREATE OR REPLACE TEMPORARY VIEW "${tableName}" AS (${query});`);
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
      `select count(*) as count FROM '${tableName}';`,
      false,
      false
    );
    return cardinality.count;
  }

  private async getMinAndMaxStringLengthsOfAllColumns(
    table: string,
    columns: ProfileColumn[]
  ) {
    /** get columns */
    // template in the column mins and maxes.
    // treat categoricals a little differently; all they have is length.
    const minAndMax = columns
      .map(
        (column) => `min(length("${column.name}")) as "min_${column.name}", 
        max(length("${column.name}")) as "max_${column.name}"`
      )
      .join(", ");
    const largestStrings = columns
      .map(
        (column) => `
      CASE WHEN "min_${column.name}" > "max_${column.name}" THEN "min_${column.name}" ELSE "max_${column.name}" END AS "${column.name}"
    `
      )
      .join(",");
    return (
      await this.databaseClient.execute(`
      WITH strings AS (SELECT ${minAndMax} from "${table}")
      SELECT ${largestStrings} from strings;
    `)
    )[0];
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
    let tableDef = (await this.databaseClient.execute(`-- parquetToDBTypes
            PRAGMA table_info(tbl_${guid});`)) as ProfileColumn[];
    const characterLengths = (await this.getMinAndMaxStringLengthsOfAllColumns(
      tableName,
      tableDef
    )) as { [key: string]: number };
    tableDef = tableDef.map((column: ProfileColumn) => {
      // get string rep length value to estimate preview table column sizes
      column.largestStringLength = characterLengths[column.name];
      return column;
    });
    try {
      await this.databaseClient.execute(`DROP TABLE tbl_${guid};`);
    } catch (err) {
      console.error(err);
    }
    return tableDef;
  }

  public async validateQuery(
    metadata: DatabaseMetadata,
    query: string
  ): Promise<void> {
    return this.databaseClient.prepare(query);
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
