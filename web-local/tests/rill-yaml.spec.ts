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
    await expect(page.getByText("Getting started")).toBeVisible();

    await page.getByRole("link", { name: "rill.yaml" }).click();
    // Wait for navigation to complete
    await page.waitForURL("**/files/rill.yaml");
    await expectRillYAMLToContainOlapConnector(page, "duckdb");
  });

  test("Should set default olap_connector to duckdb for local file upload", async ({
    page,
  }) => {
    await page.getByRole("link", { name: "Empty Project" }).click();
    await expect(page.getByText("Getting started")).toBeVisible();

    await uploadFile(page, "AdBids.csv");

    // Wait for the source file to be created in the file nav
    await waitForFileNavEntry(page, "/sources/AdBids.yaml", false);

    await page.getByText("View this source").click();

    await page.getByRole("link", { name: "rill.yaml" }).click();
    // Wait for navigation to complete
    await page.waitForURL("**/files/rill.yaml");
    await expectRillYAMLToContainOlapConnector(page, "duckdb");
  });

  test("Should set default olap_connector to clickhouse for Rill-managed ClickHouse", async ({
    page,
  }) => {
    await page.getByRole("link", { name: "Empty Project" }).click();
    await expect(page.getByText("Getting started")).toBeVisible();

    await page.getByRole("button", { name: "Add Data" }).click();
    await page.locator("#clickhouse").click();
    await page.locator("#managed").selectOption("rill-managed");
    await page
      .getByRole("dialog", { name: "ClickHouse" })
      .getByRole("button", {
        name: "Connect",
        exact: true,
      })
      .click();

    // Wait for the connector file to be created in the file nav
    await waitForFileNavEntry(page, "/connectors/clickhouse.yaml", false);

    await page.getByRole("link", { name: "rill.yaml" }).click();
    // Wait for navigation to complete
    await page.waitForURL("**/files/rill.yaml");
    await expectRillYAMLToContainOlapConnector(page, "clickhouse");
  });
});
