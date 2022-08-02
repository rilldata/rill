import { DatabaseActions } from "./DatabaseActions";
import { guidGenerator } from "$lib/util/guid";
import type { DatabaseMetadata } from "$common/database-service/DatabaseMetadata";

export class DatabaseTableActions extends DatabaseActions {
  public async createViewOfQuery(
    metadata: DatabaseMetadata,
    tableName: string,
    query: string
  ): Promise<void> {
    await this.databaseClient.execute(`-- wrapQueryAsTemporaryView
            CREATE OR REPLACE TEMPORARY VIEW ${tableName} AS (${query});`);
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
    return this.databaseClient.execute(`PRAGMA table_info(${tableName});`);
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
