import { expect, type Page } from "@playwright/test";
import { createSourceV2 } from "./utils/sourceHelpers";
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
    await page.getByLabel("Start with an empty project").click();
    await expect(
      page.getByText("Connect your data", { exact: true }),
    ).toBeVisible();

    await page.getByRole("link", { name: "rill.yaml" }).click();
    // Wait for navigation to complete
    await page.waitForURL("**/files/rill.yaml");
    await expectRillYAMLToContainOlapConnector(page, "duckdb");
  });

  test("Should set default olap_connector to duckdb for local file upload", async ({
    page,
  }) => {
    await page.getByLabel("Start with an empty project").click();
    await expect(
      page.getByText("Connect your data", { exact: true }),
    ).toBeVisible();

    await createSourceV2(page, "AdBids.csv", "/models/AdBids.yaml");

    await page.getByRole("link", { name: "rill.yaml" }).click();
    // Wait for navigation to complete
    await page.waitForURL("**/files/rill.yaml");
    await expectRillYAMLToContainOlapConnector(page, "duckdb");
  });
});
