import { EXAMPLES } from "@rilldata/web-common/features/welcome/constants";
import { expect } from "playwright/test";
import { test } from "./setup/base";

test.describe("Example project initialization", () => {
  EXAMPLES.forEach((example) => {
    test.describe(`Example project: ${example.title}`, () => {
      test("should initialize new project", async ({ page }) => {
        await page.getByRole("link", { name: example.title }).click();

        await page.waitForURL(`**/files/dashboards/${example.firstFile}`);

        await expect(
          page.getByRole("heading", { name: example.firstFile }),
        ).toBeVisible();
      });
    });
  });

  test.describe("Empty project", () => {
    test("should initialize new project", async ({ page }) => {
      await page.getByRole("link", { name: "Empty Project" }).click();

      await expect(page.getByText("Getting started")).toBeVisible();

      await page.getByRole("link", { name: "rill.yaml" }).click();

      await expect(
        page.getByRole("heading", { name: "rill.yaml" }),
      ).toBeVisible();
    });
  });

  // Different behaviors
  // @source: https://www.notion.so/rilldata/Project-init-olap_connector-Behaviors-278ba33c8f5780a5bda1d44314bc306f
  test.describe("Default olap_connector behavior", () => {
    // Empty project
    test("should set default OLAP connector to duckdb for empty project", async ({
      page,
    }) => {
      await page.getByRole("link", { name: "Empty Project" }).click();
      await expect(page.getByText("Getting started")).toBeVisible();

      // Navigate to rill.yaml to verify default OLAP connector
      await page.getByRole("link", { name: "rill.yaml" }).click();
      const rillYamlEditor = page
        .getByLabel("codemirror editor")
        .getByRole("textbox");

      // Verify that the OLAP connector is set to duckdb by default
      await expect(rillYamlEditor).toContainText("olap_connector: duckdb");
    });

    // TODO: local file using csv
    // TODO: clickhouse managed
  });
});
