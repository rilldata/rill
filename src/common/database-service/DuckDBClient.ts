import fetch from "isomorphic-unfetch";
import type { ChildProcess } from "node:child_process";
import { spawn } from "node:child_process";
import { URL } from "url";
import type { DatabaseConfig } from "$common/config/DatabaseConfig";
import { isPortOpen } from "$common/utils/isPortOpen";
import { asyncWaitUntil } from "$common/utils/waitUtils";

/**
 * Spawns or connects to a runtime and uses it to proxy DuckDB queries.
 * Runtime and database details can be configured {@link DatabaseConfig}
 *
 * There is only one runtime connection right now.
 * But in the future we can easily add an interface to this and have different implementations.
 */
export class DuckDBClient {
  protected runtimeProcess: ChildProcess;
  protected instanceID: string;

  protected onCallback: () => void;
  protected offCallback: () => void;

  // this is a singleton class because
  // duckdb doesn't work well with multiple connections to same db from same process
  // if we ever need to have different connections modify this to have a map of database to instance
  private static instance: DuckDBClient;
  private constructor(private readonly databaseConfig: DatabaseConfig) {}
  public static getInstance(databaseConfig: DatabaseConfig) {
    if (!this.instance) this.instance = new DuckDBClient(databaseConfig);
    return this.instance;
  }

  public async init(): Promise<void> {
    if (this.databaseConfig.skipDatabase) return;
    await this.spawnRuntime();
    await this.connectRuntime();
  }

  protected async spawnRuntime() {
    if (!this.databaseConfig.spawnRuntime) {
      return;
    }

    if (this.runtimeProcess) {
      throw Error("Already spawned runtime");
    }

    const httpPort = this.databaseConfig.spawnRuntimePort;
    const grpcPort = httpPort + 1000; // Hack to prevent port collision when spawning many runtimes

    this.runtimeProcess = spawn("./dist/runtime/runtime", [], {
      env: {
        ...process.env,
        RILL_RUNTIME_ENV: "production",
        RILL_RUNTIME_LOG_LEVEL: "warn",
        RILL_RUNTIME_HTTP_PORT: httpPort.toString(),
        RILL_RUNTIME_GRPC_PORT: grpcPort.toString(),
      },
      stdio: "inherit",
      shell: true,
    });

    this.runtimeProcess.on("exit", (code) => {
      process.exit(code);
    });

    await asyncWaitUntil(() =>
      isPortOpen(this.databaseConfig.spawnRuntimePort)
    );
  }

  protected async connectRuntime() {
    if (this.instanceID) {
      throw Error("Already connected to runtime");
    }

    let databaseName = this.databaseConfig.databaseName;
    if (databaseName === ":memory:") {
      databaseName = "";
    }

    const res = await this.request("/v1/instances", {
      driver: "duckdb",
      dsn: databaseName,
    });

    this.instanceID = res["instanceId"];

    await this.execute(`
      INSTALL 'json';
      INSTALL 'parquet';
      LOAD 'json';
      LOAD 'parquet';
    `);

    await this.execute(
      "PRAGMA threads=32;PRAGMA log_query_path='./log';",
      false
    );
  }

  public execute<Row = Record<string, unknown>>(
    query: string,
    log = false,
    dry_run = false
  ): Promise<Array<Row>> {
    this.onCallback?.();
    if (log) console.log(query);
    return new Promise((resolve, reject) => {
      this.request(`/v1/instances/${this.instanceID}/query/direct`, {
        sql: query,
        priority: 0,
        dry_run: dry_run,
      })
        .then((data) => {
          this.offCallback?.();
          resolve(data["data"]);
        })
        .catch((err) => {
          if (log) console.error(err);
          reject(err);
        });
    });
  }

  public async prepare(query: string): Promise<void> {
    await this.execute(query, false, true);
  }

  private async request(path: string, data: any): Promise<any> {
    let base = this.databaseConfig.runtimeUrl;
    if (!base && this.databaseConfig.spawnRuntime) {
      base = `http://localhost:${this.databaseConfig.spawnRuntimePort}`;
    }

    const url = new URL(path, base).toString();

    const res = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });

    const json = await res.json();
    if (!res.ok) {
      const msg = json["message"];
      const err = new Error(msg);
      throw err;
    }

    return json;
  }
}
