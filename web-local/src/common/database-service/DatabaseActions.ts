import type { DuckDBClient } from "./DuckDBClient";
import type { DatabaseConfig } from "../config/DatabaseConfig";

export class DatabaseActions {
  public constructor(
    protected readonly databaseConfig: DatabaseConfig,
    protected readonly databaseClient: DuckDBClient
  ) {}
}
