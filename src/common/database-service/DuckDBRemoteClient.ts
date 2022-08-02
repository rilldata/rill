import { DuckDBClient } from "$common/database-service/DuckDBClient";
import type { DatabaseConfig } from "$common/config/DatabaseConfig";
import axios from "axios";
import type { ChildProcess } from "node:child_process";
import { spawn } from "node:child_process";
import { isPortOpen } from "$common/utils/isPortOpen";
import { waitUntil } from "$common/utils/waitUtils";

export class DuckDBRemoteClient extends DuckDBClient {
  private goServer: ChildProcess;

  public async init(): Promise<void> {
    if (this.databaseConfig.skipDatabase || this.goServer) return;
    this.goServer = spawn(
      "make",
      [
        `DB_PATH="${this.databaseConfig.databaseName}"`,
        `DB_PORT=${this.databaseConfig.goServerPort}`,
        "run",
      ],
      { stdio: "inherit" },
    );
    await waitUntil(() => isPortOpen(this.databaseConfig.goServerPort));
  }

  public async destroy(): Promise<void> {
    this.goServer?.kill("SIGKILL");
  }

  public static getInstance(databaseConfig: DatabaseConfig) {
    if (!this.instance) this.instance = new DuckDBRemoteClient(databaseConfig);
    return this.instance;
  }

  public async execute<Row = Record<string, unknown>>(
    query: string,
    log = false
  ): Promise<Array<Row>> {
    const resp = await axios.post(`${this.databaseConfig.goServerUrl}/query`, {
      query,
    });
    if (log) console.log(query, resp.data);
    if (resp.status === 200) {
      return resp.data.data;
    } else {
      console.error(resp.data.message);
      return Promise.reject(new Error(resp.data.message));
    }
  }

  public async prepare(query: string): Promise<void> {
    const resp = await axios.post(
      `${this.databaseConfig.goServerUrl}/prepare`,
      { query }
    );
    if (resp.status === 200) {
      return resp.data;
    } else {
      return Promise.reject(new Error(resp.data.message));
    }
  }
}
