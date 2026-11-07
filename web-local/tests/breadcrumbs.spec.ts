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

      // The generated metrics view defines its explore dashboard inline, so the
      // metrics view crumb and the dashboard crumb both point at this file.
      await expect(link.first()).toBeVisible();
      await expect(link.first()).toHaveClass(/selected/g);

      await page
        .getByRole("link", { name: "AdBids", exact: true })
        .first()
        .click();

      await page.waitForURL("**/files/models/AdBids.yaml");

      await expect(
        page.getByRole("link", {
          name: "AdBids",
          exact: true,
        }),
      ).toBeVisible();

      await expect(
        page
          .getByRole("link", {
            name: "AdBids_metrics",
            exact: true,
          })
          .first(),
      ).toBeVisible();

      await page
        .getByRole("link", { name: "AdBids_metrics", exact: true })
        .first()
        .click();

      await page.waitForURL("**/files/metrics/AdBids_metrics.yaml");
    });
  });
});
