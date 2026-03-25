import { expect } from "playwright/test";
import { test } from "./setup/base";
import { createSourceV2 } from "./utils/sourceHelpers";

test.describe("Breadcrumbs", () => {
  test.use({ project: "Blank" });

  test.describe("Breadcrumb interactions", () => {
    test.describe.configure({ retries: 3 });
    test.setTimeout(120_000);
    test("breadcrumb navigation", async ({ page }) => {
      await createSourceV2(page, "AdBids.csv", "/models/AdBids.yaml");

      let link = page.getByRole("link", {
        name: "AdBids",
        exact: true,
      });

      await expect(link).toBeVisible();
      await expect(link).toHaveClass(/selected/g);

      await page.getByText("Generate metrics view").click();
      await page.getByText("Start simple").click();

      link = page.getByRole("link", {
        name: "AdBids_metrics",
        exact: true,
      });

      await expect(link).toBeVisible();
      await expect(link).toHaveClass(/selected/g);

      await page.getByText("Generate Explore Dashboard").click();

      await page.waitForURL("**/files/dashboards/AdBids_metrics_explore.yaml");

      link = page.getByRole("link", {
        name: "AdBids_metrics_explore",
        exact: true,
      });

      await expect(link).toBeVisible();
      await expect(link).toHaveClass(/selected/g);

      await page
        .getByRole("link", {
          name: "AdBids_metrics",
          exact: true,
        })
        .click();

      await page.getByRole("button", { name: "Create resource menu" }).click();
      await page
        .getByRole("menuitem", { name: "Generate Explore Dashboard" })
        .click();

      await page.waitForURL(
        "**/files/dashboards/AdBids_metrics_explore_1.yaml",
      );

      await page.getByRole("link", { name: "AdBids", exact: true }).click();

      await page.waitForURL("**/files/models/AdBids.yaml");

      await expect(
        page.getByRole("link", {
          name: "AdBids",
          exact: true,
        }),
      ).toBeVisible();

      await expect(
        page.getByRole("link", {
          name: "AdBids_metrics",
          exact: true,
        }),
      ).toBeVisible();

      await expect(
        page.getByRole("button", {
          name: "2 dashboards",
          exact: true,
        }),
      ).toBeVisible();

      await page
        .getByRole("link", { name: "AdBids_metrics", exact: true })
        .click();

      await page.waitForURL("**/files/metrics/AdBids_metrics.yaml");

      await page
        .getByRole("button", { name: "2 dashboards", exact: true })
        .click();
      await page
        .getByRole("menuitem", {
          name: "AdBids_metrics_explore",
          exact: true,
        })
        .click();

      await page.waitForURL("**/files/dashboards/AdBids_metrics_explore.yaml");
    });
  });
});
