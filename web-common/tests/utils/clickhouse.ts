import { ClickHouseClient, createClient } from "@clickhouse/client";
import { GenericContainer, type StartedTestContainer } from "testcontainers";
import { expect, type Page } from "@playwright/test";
import { TestDataPath } from "@rilldata/web-common/tests/utils/source-helpers.ts";

export class ClickHouseTestContainer {
  private httpPort = 8123;
  private user = "default";
  private password = "password";
  private container: StartedTestContainer;
  private client: ClickHouseClient;

  async start(): Promise<void> {
    try {
      this.container = await new GenericContainer(
        "clickhouse/clickhouse-server",
      )
        .withExposedPorts(this.httpPort)
        .withBindMounts([
          {
            source: TestDataPath,
            target: "/var/lib/clickhouse/user_files",
          },
        ])
        .withEnvironment({
          CLICKHOUSE_USER: this.user,
          CLICKHOUSE_PASSWORD: this.password,
        })
        .start();
    } catch (error) {
      console.error("Failed to start ClickHouse server:", error);
      throw error;
    }
  }

  async stop(): Promise<void> {
    try {
      if (this.container) {
        await this.container.stop();
      }
    } catch (error) {
      console.error("Failed to stop ClickHouse server:", error);
      throw error;
    }
  }

  getClient(): ClickHouseClient {
    if (!this.client) {
      const host = this.container.getHost();
      const port = this.container.getMappedPort(this.httpPort);
      this.client = createClient({
        url: `http://${host}:${port}`,
        username: this.user,
        password: this.password,
      });
    }
    return this.client;
  }

  /**
   * Seeds the ClickHouse server with an AdBids table.
   */
  async seedAdBids(): Promise<void> {
    const client = this.getClient();

    try {
      await client.command({
        query: "DROP TABLE IF EXISTS ad_bids",
      });

      await client.command({
        query: `CREATE TABLE ad_bids (
        id UInt32,
        timestamp DateTime64(3),
        publisher String,
        domain String,
        bid_price Float64
      ) ENGINE = MergeTree()
      ORDER BY id`,
      });

      await client.command({
        query: `INSERT INTO ad_bids
  SELECT
    id,
    parseDateTime64BestEffort(replaceAll(replaceAll(timestamp, 'T', ' '), 'Z', '')) AS timestamp,
    publisher,
    domain,
    bid_price
  FROM file('/var/lib/clickhouse/user_files/AdBids.csv', 'CSVWithNames',
    'id UInt32, timestamp String, publisher String, domain String, bid_price Float64');`,
      });
    } catch (error) {
      console.error("Failed to seed AdBids into ClickHouse server:", error);
      throw error;
    }
  }

  async seedAdImpressions(): Promise<void> {
    const client = this.getClient();

    try {
      await client.command({
        query: "DROP TABLE IF EXISTS ad_impressions",
      });

      // id	city	country	user_id

      await client.command({
        query: `CREATE TABLE ad_impressions (
        id UInt32,
        city String,
        country String,
        user_id UInt32
      ) ENGINE = MergeTree()
      ORDER BY id`,
      });

      await client.command({
        query: `INSERT INTO ad_impressions
  SELECT
    id,
    city,
    country,
    user_id
  FROM file('/var/lib/clickhouse/user_files/AdImpressions.tsv', 'CSVWithNames',
    'id UInt32, city String, country String, user_id UInt32');`,
      });
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

export async function enterClickhouseCredentials(
  page: Page,
  clickhouse: ClickHouseTestContainer,
  assertYaml = true,
) {
  // Switch to "self-managed", "cloud" options does not allow non-ssl connections.
  await page.getByLabel("Connection type").click();
  await page.getByRole("option").getByText("Self Managed").click();

  // Fill in the form correctly
  await page.getByRole("textbox", { name: "Host" }).fill(clickhouse.getHost());
  await page.getByRole("textbox", { name: "Host" }).press("Tab");
  await page
    .getByRole("textbox", { name: "Port" })
    .fill(clickhouse.getPort().toString());
  await page.getByRole("textbox", { name: "Port" }).press("Tab");
  await page.getByRole("textbox", { name: "Username" }).fill("default");
  await page.getByRole("textbox", { name: "Password" }).fill("password");
  await page.getByRole("checkbox").scrollIntoViewIfNeeded();
  if (await page.getByRole("checkbox").isChecked()) {
    await page.getByRole("checkbox").click();
  }

  if (assertYaml) {
    // Assert that the yaml contains key properties
    const yamlPreview = page.getByLabel("Yaml preview");
    await expect(yamlPreview).toContainText("type: connector");
    await expect(yamlPreview).toContainText("driver: clickhouse");
    await expect(yamlPreview).toContainText(`host: "${clickhouse.getHost()}"`);
    await expect(yamlPreview).toContainText(
      `port: "${clickhouse.getPort().toString()}"`,
    );
    await expect(yamlPreview).toContainText('username: "default"');
    await expect(yamlPreview).toContainText(
      'password: "{{ .env.CLICKHOUSE_PASSWORD }}"',
    );
    await expect(yamlPreview).toContainText("ssl: false");
  }
}

export async function selectAdBidsAndSubmit(
  page: Page,
  metricsViewOnly: boolean,
) {
  // Wait for pick a table screen
  await expect(
    page.getByText("Pick a table to power your first dashboard"),
  ).toBeVisible();

  // Select ad_bids table
  await page
    .getByLabel("Import Table Form")
    .getByLabel("default.default")
    .click();
  await page
    .getByLabel("Import Table Form")
    .getByLabel("clickhouse-default.ad_bids")
    .click();

  // Click the primary submit button
  const submitLabel = metricsViewOnly
    ? "Generate metrics with AI"
    : "Generate dashboard with AI";
  await page.getByRole("button", { name: submitLabel }).click();

  // Creating metrics view is triggered.
  await expect(page.getByText("Creating Metrics View...")).toBeVisible();

  if (metricsViewOnly) {
    // Wait for navigation to the new file
    await page.waitForURL(/\/files\/metrics\/ad_bids_metrics.yaml/, {
      timeout: 10_000,
    });
    return;
  }

  // Metrics view is created.
  await expect
    .poll(async () => page.getByText("Created Metrics View").isVisible(), {
      timeout: 10_000,
    })
    .toBeTruthy();

  // Wait for navigation to the new file
  await page.waitForURL(/\/files\/dashboards\/ad_bids_metrics_canvas.yaml/, {
    timeout: 10_000,
  });
}

export async function selectAdImpressionsAndSubmit(
  page: Page,
  connectorName: string,
) {
  // Wait for pick a table screen
  await expect(
    page.getByText("Pick a table to power your first dashboard"),
  ).toBeVisible();

  // Select `ad_impressions` from the second connector
  await page
    .getByLabel("Import Table Form")
    .getByLabel("default.default")
    .click();
  await page
    .getByLabel("Import Table Form")
    .getByLabel(`${connectorName}-default.ad_impressions`)
    .click();

  // Click the primary submit button (metrics-view-only flow).
  await page.getByRole("button", { name: "Generate metrics with AI" }).click();

  // Wait for navigation to the new file
  await page.waitForURL(/\/files\/metrics\/ad_impressions_metrics.yaml/, {
    timeout: 10_000,
  });
}
