import type {DuckDBClient} from "$common/duckdb/DuckDBClient";

export class DuckDBAPI {
    public constructor(protected readonly duckDBClient: DuckDBClient) {}
}
