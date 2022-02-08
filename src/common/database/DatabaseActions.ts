import type {DuckDBClient} from "$common/database/DuckDBClient";

export class DatabaseActions {
    public constructor(protected readonly dbClient: DuckDBClient) {}
}
