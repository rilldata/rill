import { expect, type Page } from "@playwright/test";
import { uploadFile } from "./utils/sourceHelpers";
import { test } from "./setup/base";

async function expectRillYAMLToContainOlapConnector(page: Page, text: string) {
  // Wait for the editor to be visible first
  const rillYamlEditor = page
    .getByLabel("codemirror editor")
    .getByRole("textbox");

  // Wait for the editor to be visible with a longer timeout
  await expect(rillYamlEditor).toBeVisible({ timeout: 10000 });

  // Wait a bit for the content to load
  await page.waitForTimeout(1000);

  // Check for the text with a longer timeout
  await expect(rillYamlEditor).toContainText(`olap_connector: ${text}`, {
    timeout: 10000,
  });
}

test.describe("Default olap_connector behavior", () => {
  test.describe.configure({ retries: 2 }); // Add retries for flaky tests
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

    // Wait for the connector to be created and reconciled
    // The backend will spin up a new clickhouse instance, download the binary, and start the server
    // We wait for the connector file to be created and the reconciliation to complete
    await page.waitForURL(`**/files/connectors/clickhouse.yaml`);

    await page.getByRole("link", { name: "rill.yaml" }).click();
    // Wait for navigation to complete
    await page.waitForURL("**/files/rill.yaml");
    await expectRillYAMLToContainOlapConnector(page, "clickhouse");
  });
});
