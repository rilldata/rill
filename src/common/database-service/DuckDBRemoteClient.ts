import { DuckDBClient } from "$common/database-service/DuckDBClient";
import type { DatabaseConfig } from "$common/config/DatabaseConfig";
import axios from "axios";

export class DuckDBRemoteClient extends DuckDBClient {
  public async init(): Promise<void> {}

  public static getInstance(databaseConfig: DatabaseConfig) {
    if (!this.instance) this.instance = new DuckDBRemoteClient(databaseConfig);
    return this.instance;
  }

  public async execute<Row = Record<string, unknown>>(
    query: string,
    log = false
  ): Promise<Array<Row>> {
    const resp = await axios.post("http://localhost:3100/query", { query });
    if (log) console.log(query, resp.data);
    if (resp.status === 200) {
      return resp.data.data;
    } else {
      console.error(resp.data.message);
      return Promise.reject(new Error(resp.data.message));
    }
  }

  public async prepare(query: string): Promise<void> {
    const resp = await axios.post("http://localhost:3100/prepare", { query });
    if (resp.status === 200) {
      return resp.data;
    } else {
      return Promise.reject(new Error(resp.data.message));
    }
  }
}
