import {
  escapeColumn,
  escapeColumnAlias,
} from "@rilldata/web-local/common/database-service/columnUtils";
import type { ProfileColumn } from "@rilldata/web-local/lib/types";
import { guidGenerator } from "@rilldata/web-local/lib/util/guid";
import { DatabaseActions } from "./DatabaseActions";
import type { DatabaseMetadata } from "./DatabaseMetadata";

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
    tableDef = tableDef.map((column: ProfileColumn, index) => {
      // get string rep length value to estimate preview table column sizes
      column.largestStringLength = characterLengths[`col_${index}`];
      return column;
    });
    try {
      await this.databaseClient.execute(`DROP TABLE tbl_${guid};`);
    } catch (err) {
      // no-op
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

  private async getMinAndMaxStringLengthsOfAllColumns(
    table: string,
    columns: ProfileColumn[]
  ) {
    /** get columns */
    const columnNames = columns
      .map(
        (column, index) =>
          [escapeColumn(column.name), index] as [string, number]
      )
      .filter(([columnName]) => columnName !== "");
    // template in the column mins and maxes.
    // treat categoricals a little differently; all they have is length.
    const minAndMax = columnNames
      .map(([columnName]) => {
        const columnAlias = escapeColumnAlias(columnName);
        return (
          `min(length(${columnName})) as "min_${columnAlias}",` +
          `max(length(${columnName})) as "max_${columnAlias}"`
        );
      })
      .join(", ");
    const largestStrings = columnNames
      .map(([columnName, index]) => {
        const columnAlias = escapeColumnAlias(columnName);
        return (
          `CASE WHEN "min_${columnAlias}" > "max_${columnAlias}" THEN "min_${columnAlias}" ` +
          `ELSE "max_${columnAlias}" END AS col_${index}`
        );
      })
      .join(",");
    try {
      return (
        await this.databaseClient.execute(
          `
      WITH strings AS (SELECT ${minAndMax} from "${table}")
      SELECT ${largestStrings} from strings;
    `
        )
      )[0];
    } catch (err) {
      return {};
    }
  }
}
