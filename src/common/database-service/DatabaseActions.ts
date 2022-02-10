import type {DuckDBClient} from "$common/database-service/DuckDBClient";

export class DatabaseActions {
    public constructor(protected readonly dbClient: DuckDBClient) {}
}
