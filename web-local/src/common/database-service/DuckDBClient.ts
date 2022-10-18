import fetch from "isomorphic-unfetch";
import type { ChildProcess } from "node:child_process";
import childProcess from "node:child_process";
import { URL } from "url";
import type { RootConfig } from "../config/RootConfig";
import { isPortOpen } from "../utils/isPortOpen";
import { asyncWaitUntil } from "../utils/waitUtils";
import { getBinaryRuntimePath } from "./getBinaryRuntimePath";

/**
 * Spawns or connects to a runtime and uses it to proxy DuckDB queries.
 * Runtime and database details can be configured {@link DatabaseConfig}
 *
 * There is only one runtime connection right now.
 * But in the future we can easily add an interface to this and have different implementations.
 */
export class DuckDBClient {
  // this is a singleton class because
  // duckdb doesn't work well with multiple connections to same db from same process
  // if we ever need to have different connections modify this to have a map of database to instance
  private static instance: DuckDBClient;
  protected runtimeProcess: ChildProcess;
  protected instanceID: string;
  protected onCallback: () => void;

  protected offCallback: () => void;

  private constructor(private readonly config: RootConfig) {}
  public static getInstance(config: RootConfig) {
    if (!this.instance) this.instance = new DuckDBClient(config);
    return this.instance;
  }

  public async init(): Promise<void> {
    if (this.config.database.skipDatabase) return;
    await this.spawnRuntime();
    await this.connectRuntime();
  }

  public async destroy(): Promise<void> {
    this.runtimeProcess?.kill();
    this.runtimeProcess = undefined;
    this.instanceID = undefined;
  }

  public async execute<Row = Record<string, unknown>>(
    query: string,
    log = false,
    dry_run = false
  ): Promise<Array<Row>> {
    this.onCallback?.();
    if (log) console.log(query);

    try {
      const resp = await this.request(
        `/v1/instances/${this.instanceID}/query/direct`,
        {
          sql: query,
          priority: 0,
          dry_run: dry_run,
        }
      );
      if (log) console.log(resp.data);
      return resp.data;
    } catch (err) {
      if (log) console.error(err);
      throw err;
    }
  }

  public async prepare(query: string): Promise<void> {
    await this.execute(query, false, true);
  }

  public async requestToInstance(path: string, data: any): Promise<any> {
    return this.request(`/v1/instances/${this.instanceID}/${path}`, data);
  }

  public getInstanceId(): string {
    return this.instanceID;
  }

  protected async spawnRuntime() {
    if (!this.config.database.spawnRuntime) {
      return;
    }

    if (this.runtimeProcess) {
      console.log("Already spawned runtime");
      // do not throw error. this is the case when the same instance is reused
      return;
    }

    const httpPort = this.config.database.spawnRuntimePort;
    const grpcPort = httpPort + 1000; // Hack to prevent port collision when spawning many runtimes

    if ((await isPortOpen(httpPort)) || (await isPortOpen(grpcPort))) {
      // TODO: once isDev is merged, throw error when isDev=false
      // throw Error(`Ports ${httpPort} or ${grpcPort} already in use.`);
      console.warn(`Ports ${httpPort} or ${grpcPort} already in use.`);
      return;
    }

    this.runtimeProcess = childProcess.spawn(
      getBinaryRuntimePath(this.config.local.version),
      [],
      {
        env: {
          ...process.env,
          RILL_RUNTIME_ENV: "production",
          RILL_RUNTIME_LOG_LEVEL: "warn",
          RILL_RUNTIME_HTTP_PORT: httpPort.toString(),
          RILL_RUNTIME_GRPC_PORT: grpcPort.toString(),
        },
        stdio: "inherit",
        shell: true,
      }
    );
    this.runtimeProcess.on("error", console.log);

    await asyncWaitUntil(() =>
      isPortOpen(this.config.database.spawnRuntimePort)
    );
  }

  protected async connectRuntime() {
    if (this.instanceID) {
      console.log("Already connected to runtime");
      return;
    }

    let databaseName = this.config.database.databaseName;
    if (databaseName === ":memory:") {
      databaseName = "";
    }

    const res = await this.request("/v1/instances", {
      driver: "duckdb",
      dsn: databaseName,
      exposed: true,
      embed_catalog: true,
    });

    this.instanceID = res["instanceId"];

    await this.execute(`
      INSTALL 'json';
      INSTALL 'parquet';
      INSTALL 'httpfs';
      LOAD 'json';
      LOAD 'parquet';
      LOAD 'httpfs';
    `);

    await this.execute(
      "PRAGMA threads=32;PRAGMA log_query_path='./log';",
      false
    );
  }

  private async request(path: string, data: any): Promise<any> {
    const url = new URL(path, this.config.database.runtimeUrl).toString();

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
      throw new Error(msg);
    }

    return json;
  }
}
