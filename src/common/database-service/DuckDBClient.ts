import duckdb from "duckdb";
import type { RootConfig } from "$common/config/RootConfig";

interface DuckDB {
    // TODO: define concrete styles
    all: (...args: Array<any>) => any;
    exec: (...args: Array<any>) => any;
    prepare: (...args: Array<any>) => any;
}

const DUCK_MEMORY_DB = ":memory:";

/**
 * Runs a duckdb instance. Database name can be configured {@link DatabaseConfig}
 *
 * There is only one db right now.
 * But in the future we can easily add an interface to this and have different implementations.
 */
export class DuckDBClient {
    protected db: DuckDB;

    protected onCallback: () => void;
    protected offCallback: () => void;

    public constructor(private readonly config: RootConfig) {}

    public async init(): Promise<void> {
        if (this.config.database.skipDatabase) return;
        // we can later on swap this over to WASM and update data loader
        this.db = new duckdb.Database(
            this.config.database.databaseName === DUCK_MEMORY_DB ?
                this.config.database.databaseName:
                `${this.config.projectFolder}/${this.config.database.databaseName}`);
        this.db.exec("PRAGMA threads=32;PRAGMA log_query_path='./log';");
    }

    public execute(query: string): Promise<any> {
        this.onCallback?.();
        return new Promise((resolve, reject) => {
            try {
                this.db.all(query, (err, res) => {
                    if (err !== null) {
                        reject(err);
                    } else {
                        this.offCallback?.();
                        resolve(res);
                    }
                });
            } catch (err) {
                reject(err);
            }
        });
    }

    public prepare(query: string): Promise<any> {
        return new Promise((resolve, reject) => {
            this.db.prepare(query, (err, stmt) => {
                if (err !== null) reject(err);
                else resolve(stmt);
            });
        });
    }
}
