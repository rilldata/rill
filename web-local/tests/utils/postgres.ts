import path from "path";
import { GenericContainer, type StartedTestContainer } from "testcontainers";
import { fileURLToPath } from "url";
import { Client } from "pg";
import { expect } from "@playwright/test";

export class PostgresTestContainer {
  private httpPort = 5432;
  private user = "default";
  private password = "password";
  private container: StartedTestContainer;
  private client: Client;

  async start(): Promise<void> {
    const __filename = fileURLToPath(import.meta.url);
    const __dirname = path.dirname(__filename);
    const testDataDir = path.join(__dirname, "../data");

    try {
      this.container = await new GenericContainer("postgres:16")
        .withExposedPorts(this.httpPort)
        .withBindMounts([
          {
            source: testDataDir,
            target: "/tmp/files",
          },
        ])
        .withEnvironment({
          POSTGRES_USER: this.user,
          POSTGRES_PASSWORD: this.password,
        })
        .start();
    } catch (error) {
      console.error("Failed to start Postgres server:", error);
      throw error;
    }
  }

  async stop(): Promise<void> {
    try {
      if (this.client) await this.client.end();
      if (this.container) await this.container.stop();
    } catch (error) {
      console.error("Failed to stop ClickHouse server:", error);
      throw error;
    }
  }

  async getClient(): Client {
    if (!this.client) {
      const host = this.container.getHost();
      const port = this.container.getMappedPort(this.httpPort);
      this.client = new Client({
        host,
        port,
        user: this.user,
        password: this.password,
        database: "postgres",
      });
      // Wait for postgres server to be ready.
      await expect
        .poll(
          async () => {
            try {
              await this.client.connect();
              return true;
            } catch {
              return false;
            }
          },
          { timeout: 10_000 },
        )
        .toBeTruthy();
    }
    return this.client;
  }

  /**
   * Seeds the ClickHouse server with an AdBids table.
   */
  async seedAdBids(): Promise<void> {
    const client = await this.getClient();

    try {
      await client.query("DROP TABLE IF EXISTS ad_bids");

      await client.query(`CREATE TABLE ad_bids (
        id INTEGER,
        timestamp TIMESTAMP,
        publisher VARCHAR(255),
        domain VARCHAR(255),
        bid_price DECIMAL(10, 6)
      )`);

      await client.query(
        `COPY ad_bids (id, timestamp, publisher, domain, bid_price)
FROM '/tmp/files/AdBids.csv'
DELIMITER ','
CSV HEADER;`,
      );
    } catch (error) {
      console.error("Failed to seed AdBids into ClickHouse server:", error);
      throw error;
    }
  }

  async seedAdImpressions(): Promise<void> {
    const client = await this.getClient();

    try {
      await client.query("DROP TABLE IF EXISTS ad_impressions");

      // id	city	country	user_id

      await client.query(`CREATE TABLE ad_impressions (
        id INTEGER,
        city VARCHAR(255),
        country VARCHAR(255),
        user_id INTEGER
      )`);

      await client.query(
        `COPY ad_impressions (id, city, country, user_id)
FROM '/tmp/files/AdImpressions.tsv'
DELIMITER '\t'
CSV HEADER;`,
      );
    } catch (error) {
      console.error(
        "Failed to seed AdImpressions into ClickHouse server:",
        error,
      );
      throw error;
    }
  }

  getHost(): string {
    return this.container.getHost();
  }

  getPort(): number {
    return this.container.getMappedPort(this.httpPort);
  }
}
