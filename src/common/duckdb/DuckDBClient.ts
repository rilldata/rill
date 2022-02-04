import duckdb from "duckdb";

interface DuckDB {
    // TODO: define concrete styles
    all: (...args: Array<any>) => any;
    exec: (...args: Array<any>) => any;
    run: (...args: Array<any>) => any;
    prepare: (...args: Array<any>) => any;
}

export class DuckDBClient {
    protected db: DuckDB;

    protected onCallback: () => void;
    protected offCallback: () => void;

    public async init(): Promise<void> {
        // we can later on swap this over to WASM and update data loader
        this.db = new duckdb.Database(":memory:");
        this.db.exec("PRAGMA threads=32;PRAGMA log_query_path='./log';");
    }

    public all(query: string): Promise<any> {
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

    public run(query: string): Promise<any> {
        return new Promise((resolve, reject) => {
            this.db.run(query, (err) => {
                if (err !== null) reject(false);
                else resolve(true);
            });
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
