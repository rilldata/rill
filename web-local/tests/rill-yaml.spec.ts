import { expect, type Page } from "@playwright/test";
import { uploadFile } from "./utils/sourceHelpers";
import { waitForFileNavEntry } from "./utils/waitHelpers";
import { test } from "./setup/base";

async function expectRillYAMLToContainOlapConnector(page: Page, text: string) {
  // Wait for the editor to be visible first
  const rillYamlEditor = page
    .getByLabel("codemirror editor")
    .getByRole("textbox");

  // Check for the text with a longer timeout
  await expect(rillYamlEditor).toContainText(`olap_connector: ${text}`);
}

test.describe("Default olap_connector behavior", () => {
  test("Should set default olap_connector to duckdb for empty project", async ({
    page,
  }) => {
    await page.getByRole("link", { name: "Empty Project" }).click();
    await expect(page.getByText("Import data", { exact: true })).toBeVisible();

    await page.getByRole("link", { name: "rill.yaml" }).click();
    // Wait for navigation to complete
    await page.waitForURL("**/files/rill.yaml");
    await expectRillYAMLToContainOlapConnector(page, "duckdb");
  });

  test("Should set default olap_connector to duckdb for local file upload", async ({
    page,
  }) => {
    await page.getByRole("link", { name: "Empty Project" }).click();
    await expect(page.getByText("Import data", { exact: true })).toBeVisible();

    await uploadFile(page, "AdBids.csv");

    // Wait for the source file to be created in the file nav
    await waitForFileNavEntry(page, "/sources/AdBids.yaml", false);

    await page.getByText("View this source").click();

    await page.getByRole("link", { name: "rill.yaml" }).click();
    // Wait for navigation to complete
    await page.waitForURL("**/files/rill.yaml");
    await expectRillYAMLToContainOlapConnector(page, "duckdb");
  });

  test("Should set olap_connector to clickhouse for Rill-managed ClickHouse", async ({
    page,
  }) => {
    await page.getByRole("link", { name: "Empty Project" }).click();
    await expect(page.getByText("Import data", { exact: true })).toBeVisible();

    // Open Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select ClickHouse
    await page.locator("#clickhouse").click();

    // Select rill-managed from the connection type dropdown
    await page.getByRole("combobox").click();
    await page.getByRole("option", { name: "Rill Managed" }).click();

    // Submit the form
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Connect" })
      .click();

    // Wait for the connector file to be created
    await waitForFileNavEntry(page, "/connectors/clickhouse.yaml", false);

    // Verify rill.yaml has correct olap_connector
    await page.getByRole("link", { name: "rill.yaml" }).click();
    await page.waitForURL("**/files/rill.yaml");
    await expectRillYAMLToContainOlapConnector(page, "clickhouse");
  });
});
