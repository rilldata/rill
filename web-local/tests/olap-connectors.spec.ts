import { expect } from "@playwright/test";
import { test } from "./setup/base";

/**
 * These tests focus on form validation and YAML generation for OLAP connectors.
 * They don't test connections to live OLAP databases - that's handled by runtime integration tests.
 * Future work: Add tests using ClickHouse testcontainers for end-to-end validation.
 */

test.describe("ClickHouse connector", () => {
  test.use({ project: "Blank" });

  test("Create connector using individual fields", async ({ page }) => {
    // Open the Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select ClickHouse
    await page.locator("#clickhouse").click();

    // Verify form validation - empty host
    await page
      .getByRole("dialog", { name: "ClickHouse" })
      .getByRole("button", {
        name: "Add data",
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
    await page.getByRole("textbox", { name: "Host" }).fill("localhost");
    await page.getByRole("textbox", { name: "Host" }).press("Tab");
    await page.getByRole("textbox", { name: "Port (optional)" }).fill("8123");
    await page.getByRole("textbox", { name: "Port (optional)" }).press("Tab");
    await page
      .getByRole("textbox", { name: "Username (optional)" })
      .fill("default");

    // Submit the form
    await page
      .getByRole("dialog", { name: "ClickHouse" })
      .getByRole("button", { name: "Add data" })
      .click();

    // Assert that the file contains key properties
    const codeEditor = page.getByLabel("Code editor").getByRole("textbox");
    await expect(codeEditor).toContainText("type: connector");
    await expect(codeEditor).toContainText("driver: clickhouse");
    await expect(codeEditor).toContainText('host: "localhost"');
    await expect(codeEditor).toContainText("port: 8123");
    await expect(codeEditor).toContainText('username: "default"');

    // Assert that the connector explorer now has a ClickHouse connector
    await page.pause();
    await expect(
      page
        .getByRole("region", { name: "Connector explorer" })
        .getByRole("button", {
          name: "clickhouse",
          exact: true,
        }),
    ).toBeVisible();
  });

  test("Create connector using DSN", async ({ page }) => {
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
      .fill("http://localhost:8123?username=default&password=password");

    // Submit the form
    await page
      .getByRole("dialog", { name: "ClickHouse" })
      .getByRole("button", { name: "Add data" })
      .click();

    // Assert that the file contains key properties
    const codeEditor = page.getByLabel("Code editor").getByRole("textbox");
    await expect(codeEditor).toContainText("type: connector");
    await expect(codeEditor).toContainText("driver: clickhouse");
    await expect(codeEditor).toContainText(
      'dsn: "{{ .env.connector.clickhouse.dsn }}"',
    );

    // Go to the `.env` file and verify the connector.clickhouse.dsn is set
    await page.getByRole("link", { name: ".env" }).click();
    const envEditor = page.getByLabel("Code editor").getByRole("textbox");
    await expect(envEditor).toContainText(
      "connector.clickhouse.dsn=http://localhost:8123?username=default&password=password",
    );

    // Go to the `rill.yaml` and verify the OLAP connector is set
    await page.getByRole("link", { name: "rill.yaml" }).click();
    const rillYamlEditor = page.getByLabel("Code editor").getByRole("textbox");
    await expect(rillYamlEditor).toContainText("olap_connector: clickhouse");

    // Assert that the connector explorer now has a ClickHouse connector
    await expect(
      page
        .getByRole("region", { name: "Connector explorer" })
        .getByRole("button", {
          name: "clickhouse",
          exact: true,
        }),
    ).toBeVisible();
  });
});
