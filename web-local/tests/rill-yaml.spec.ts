import { expect, type Page } from "@playwright/test";
import { uploadFile } from "./utils/sourceHelpers";
import { test } from "./setup/base";

async function expectRillYAMLToContain(page: Page, text: string) {
  const rillYamlEditor = page
    .getByLabel("codemirror editor")
    .getByRole("textbox");
  await expect(rillYamlEditor).toContainText(`olap_connector: ${text}`);
}

test.describe("Default olap_connector behavior", () => {
  test("Should set default olap_connector to duckdb for empty project", async ({
    page,
  }) => {
    await page.getByRole("link", { name: "Empty Project" }).click();
    await expect(page.getByText("Getting started")).toBeVisible();

    await page.getByRole("link", { name: "rill.yaml" }).click();
    await expectRillYAMLToContain(page, "duckdb");
  });

  test("Should set default olap_connector to duckdb for local file upload", async ({
    page,
  }) => {
    await page.getByRole("link", { name: "Empty Project" }).click();
    await expect(page.getByText("Getting started")).toBeVisible();

    await uploadFile(page, "AdBids.csv");

    await page.getByText("View this source").click();

    await page.getByRole("link", { name: "rill.yaml" }).click();
    await expectRillYAMLToContain(page, "duckdb");
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

    // Wait for navigation to the connector file
    await page.waitForURL(`**/files/connectors/clickhouse.yaml`);

    await page.getByRole("link", { name: "rill.yaml" }).click();
    await expectRillYAMLToContain(page, "clickhouse");
  });
});
