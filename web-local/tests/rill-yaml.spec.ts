import { expect } from "@playwright/test";
import { uploadFile } from "./utils/sourceHelpers";
import { test } from "./setup/base";

test.describe("Default olap_connector behavior", () => {
  test("Start with an empty project", async ({ page }) => {
    await page.getByRole("link", { name: "Empty Project" }).click();
    await expect(page.getByText("Getting started")).toBeVisible();

    await page.getByRole("link", { name: "rill.yaml" }).click();
    const rillYamlEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");

    await expect(rillYamlEditor).toContainText("olap_connector: duckdb");
  });

  test("Start with a new project with a local file", async ({ page }) => {
    await page.getByRole("link", { name: "Empty Project" }).click();
    await expect(page.getByText("Getting started")).toBeVisible();

    await uploadFile(page, "AdBids.csv");

    await page.getByText("View this source").click();

    await page.getByRole("link", { name: "rill.yaml" }).click();
    const rillYamlEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");

    await expect(rillYamlEditor).toContainText("olap_connector: duckdb");
  });
});
