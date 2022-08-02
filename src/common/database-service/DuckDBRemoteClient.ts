import { DuckDBClient } from "$common/database-service/DuckDBClient";
import type { DatabaseConfig } from "$common/config/DatabaseConfig";
import axios from "axios";
import type { ChildProcess } from "node:child_process";
import { spawn } from "node:child_process";
import { isPortOpen } from "$common/utils/isPortOpen";
import { waitUntil } from "$common/utils/waitUtils";
import os from "node:os";

export class DuckDBRemoteClient extends DuckDBClient {
  private goServer: ChildProcess;

  public async init(): Promise<void> {
    if (this.databaseConfig.skipDatabase || this.goServer) return;
    const { args, env } = getGoRunCommand();
    this.goServer = spawn(
      "go",
      [
        ...args,
        "runtime/main.go",
        this.databaseConfig.databaseName,
        "" + this.databaseConfig.goServerPort,
      ],
      {
        env: {
          ...process.env,
          ...env,
        },
        stdio: "inherit",
        shell: true,
      }
    );
    await waitUntil(() => isPortOpen(this.databaseConfig.goServerPort));
  }

  public async destroy(): Promise<void> {
    this.goServer?.kill();
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

function getGoRunCommand() {
  const [LIB_EXT, LIBRARY_PATH, LIB_PATH] = getLibExtAndPath();
  return {
    args: ["run", `-ldflags="-r ${LIB_PATH}"`],
    env: {
      LIB: `libduckdb.${LIB_EXT}`,
      CGO_LDFLAGS: `-L${LIB_PATH}`,
      [LIBRARY_PATH]: LIB_PATH,
      CGO_CFLAGS: `-I${LIB_PATH}`,
    },
  };
}

function getLibExtAndPath() {
  const LIB_PATH = `${process.cwd()}/lib`;
  switch (os.platform()) {
    case "darwin":
      return ["dylib", "DYLD_LIBRARY_PATH", LIB_PATH];
    default:
      return ["so", "LD_LIBRARY_PATH", LIB_PATH];
  }
}
