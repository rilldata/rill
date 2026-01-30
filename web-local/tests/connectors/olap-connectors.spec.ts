import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { ClickHouseTestContainer } from "../utils/clickhouse";

test.describe("ClickHouse connector", () => {
  /*
   * NOTE: These tests are flaky due to a race condition:
   * 1. When navigation to the new connector file is in progress
   * 2. An edit to `rill.yaml` triggers `invalidate("init")`
   * 3. This re-runs the root load function with its own navigation logic
   *
   * Note: This issue occurs during automated test runs (both CI and local),
   * but has not been reproduced when performing the steps manually.
   */
  test.describe.configure({ retries: 3 });

  test.use({ project: "Blank" });

  const clickhouse = new ClickHouseTestContainer();

  test.beforeAll(async () => {
    await clickhouse.start();
    await clickhouse.seed();
  });

  test.afterAll(async () => {
    await clickhouse.stop();
  });

  // Flaky
  test.skip("Create connector using individual fields", async ({ page }) => {
    // Open the Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select ClickHouse
    await page.locator("#clickhouse").click();

    // Verify form validation - empty host
    await page
      .getByRole("dialog", { name: "ClickHouse" })
      .getByRole("button", {
        name: "Connect",
        exact: true,
      })
      .click();
    await expect(page.getByText("Host is required")).toBeVisible();

    // Verify form validation - invalid host with protocol prefix
    await page.getByRole("textbox", { name: "Host" }).click();
    await page.getByRole("textbox", { name: "Host" }).fill("http://localhost");
    await page.getByRole("textbox", { name: "Host" }).press("Tab");
    await expect(
      page.getByText("Do not prefix the host with `http(s)://`"),
    ).toBeVisible();

    // Now, fill in the form correctly
    await page
      .getByRole("textbox", { name: "Host" })
      .fill(clickhouse.getHost());
    await page.getByRole("textbox", { name: "Host" }).press("Tab");
    await page
      .getByRole("textbox", { name: "Port (optional)" })
      .fill(clickhouse.getPort().toString());
    await page.getByRole("textbox", { name: "Port (optional)" }).press("Tab");
    await page
      .getByRole("textbox", { name: "Username (optional)" })
      .fill("default");
    await page
      .getByRole("textbox", { name: "Password (optional)" })
      .fill("password");

    // Submit the form
    await page
      .getByRole("dialog", { name: "ClickHouse" })
      .getByRole("button", { name: "Connect", exact: true })
      .click();

    // Wait for navigation to the new file
    await page.waitForURL(`**/files/connectors/clickhouse.yaml`);

    // Assert that the file contains key properties
    const codeEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(codeEditor).toContainText("type: connector");
    await expect(codeEditor).toContainText("driver: clickhouse");
    await expect(codeEditor).toContainText(`host: "${clickhouse.getHost()}"`);
    await expect(codeEditor).toContainText(
      `port: ${clickhouse.getPort().toString()}`,
    );
    await expect(codeEditor).toContainText('username: "default"');
    await expect(codeEditor).toContainText(
      'password: "{{ .env.connector.clickhouse.password }}"',
    );

    // Assert that the connector explorer now has a ClickHouse connector
    await expect(
      page.getByRole("region", { name: "Data explorer" }).getByRole("button", {
        name: "clickhouse",
        exact: true,
      }),
    ).toBeVisible();
  });

  // Flaky
  test.skip("Create connector using DSN", async ({ page }) => {
    // Open the Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select ClickHouse
    await page.locator("#clickhouse").click();

    // Switch to the DSN tab
    await page.getByRole("button", { name: "Use connection string" }).click();

    // Fill in the form correctly
    await page
      .getByRole("textbox", { name: "Connection string" })
      .fill(
        `http://${clickhouse.getHost()}:${clickhouse.getPort().toString()}?username=default&password=password`,
      );

    // Submit the form
    await page
      .getByRole("dialog", { name: "ClickHouse" })
      .getByRole("button", { name: "Connect", exact: true })
      .click();

    // Wait for navigation to the new file
    await page.waitForURL(`**/files/connectors/clickhouse.yaml`);

    // Assert that the file contains key properties
    const codeEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(codeEditor).toContainText("type: connector");
    await expect(codeEditor).toContainText("driver: clickhouse");
    await expect(codeEditor).toContainText(
      'dsn: "{{ .env.connector.clickhouse.dsn }}"',
    );

    // Go to the `.env` file and verify the connector.clickhouse.dsn is set
    await page.getByRole("link", { name: ".env" }).click();
    const envEditor = page.getByLabel("codemirror editor").getByRole("textbox");
    await expect(envEditor).toContainText(
      `connector.clickhouse.dsn=http://${clickhouse.getHost()}:${clickhouse.getPort().toString()}?username=default&password=password`,
    );

    // Go to the `rill.yaml` and verify the OLAP connector is set
    await page.getByRole("link", { name: "rill.yaml" }).click();
    const rillYamlEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(rillYamlEditor).toContainText("olap_connector: clickhouse");

    // Assert that the connector explorer now has a ClickHouse connector
    await expect(
      page.getByRole("region", { name: "Data explorer" }).getByRole("button", {
        name: "clickhouse",
        exact: true,
      }),
    ).toBeVisible();
  });
});
