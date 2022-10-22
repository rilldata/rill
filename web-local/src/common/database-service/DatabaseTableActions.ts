import type { DatabaseMetadata } from "./DatabaseMetadata";
import type { ProfileColumn } from "@rilldata/web-local/lib/types";
import { guidGenerator } from "@rilldata/web-local/lib/util/guid";
import { DatabaseActions } from "./DatabaseActions";

function escapeColumn(columnName: string): string {
  return columnName.replace(/"/g, "'");
}

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
    tableDef = tableDef.map((column: ProfileColumn) => {
      // get string rep length value to estimate preview table column sizes
      column.largestStringLength = characterLengths[column.name];
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
    // template in the column mins and maxes.
    // treat categoricals a little differently; all they have is length.
    const minAndMax = columns
      .map((column) => {
        const escapedColumn = escapeColumn(column.name);
        return (
          `min(length('${column.name}')) as "min_${escapedColumn}",` +
          `max(length('${column.name}')) as "max_${escapedColumn}"`
        );
      })
      .join(", ");
    const largestStrings = columns
      .map((column) => {
        const escapedColumn = escapeColumn(column.name);
        return (
          `CASE WHEN "min_${escapedColumn}" > "max_${escapedColumn}" THEN "min_${escapedColumn}" ` +
          `ELSE "max_${escapedColumn}" END AS "${escapedColumn}"`
        );
      })
      .join(",");
    return (
      await this.databaseClient.execute(
        `
      WITH strings AS (SELECT ${minAndMax} from "${table}")
      SELECT ${largestStrings} from strings;
    `
      )
    )[0];
  }
}
