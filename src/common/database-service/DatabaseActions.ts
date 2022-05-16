import type { DuckDBClient } from "$common/database-service/DuckDBClient";
import type { DatabaseConfig } from "$common/config/DatabaseConfig";

export class DatabaseActions {
  public constructor(
    protected readonly databaseConfig: DatabaseConfig,
    protected readonly databaseClient: DuckDBClient
  ) {}
}
