import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { waitForReconciliation } from "@rilldata/web-common/tests/utils/wait-for-reconciliation";
import { gotoNavEntry } from "../utils/waitHelpers";

test.describe("visual explore editing", () => {
  test.use({ project: "AdBids" });

  test("visual explore editor runthrough", async ({ page }) => {
    test.setTimeout(45_000); // Note: we should make this test smaller!

    // Wait for all resources (including models) to fully reconcile
    // before interacting with the visual editor
    await waitForReconciliation(page);

    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "switch to code editor" }).click();

    await page.getByRole("button", { name: "Subset" }).first().click();
    await page.getByRole("button", { name: "Subset" }).nth(1).click();

    await page.getByRole("button", { name: "Custom" }).first().click();
    await page.getByRole("button", { name: "Custom" }).nth(1).click();
    await page.getByRole("button", { name: "Custom" }).nth(2).click();

    // Poll the editor content so the assertion waits for the visual-editor
    // edits to flush into the code editor rather than reading mid-update.
    await expect
      .poll(() =>
        page.getByRole("textbox", { name: "codemirror editor" }).textContent(),
      )
      .toContain(
        '# Explore YAML# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboardstype: exploretitle: "Adbids dashboard"metrics_view: AdBids_metricsdimensions:  - publisher  - domain  - timestamp  - offset_timestampmeasures:  - total_records  - bid_price_sumtime_ranges:  - PT6H  - PT24H  - P7D  - P14D  - P4W  - P12M  - rill-TD  - rill-WTD  - rill-MTD  - rill-QTD  - rill-YTD  - rill-PDC  - rill-PWC  - rill-PMC  - rill-PQC  - rill-PYC  - inftime_zones:  - UTC  - America/Los_Angeles  - America/Chicago  - America/New_York  - Europe/London  - Europe/Paris  - Asia/Jerusalem  - Europe/Moscow  - Asia/Kolkata  - Asia/Shanghai  - Asia/Tokyo  - Australia/Sydneytheme:  light:    primary: hsl(180, 100%, 50%)    secondary: lightgreen  dark:    primary: hsl(180, 100%, 50%)    secondary: lightgreen',
      );

    await page.getByRole("button", { name: "Expression" }).first().click();
    await page.getByRole("button", { name: "Expression" }).nth(1).click();

    await page.getByRole("button", { name: "Default" }).first().click();
    await page.getByRole("button", { name: "Presets" }).click();

    await expect
      .poll(() =>
        page.getByRole("textbox", { name: "codemirror editor" }).textContent(),
      )
      .toContain(
        '# Explore YAML# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboardstype: exploretitle: "Adbids dashboard"metrics_view: AdBids_metricsdimensions:  expr: "*"measures:  expr: "*"time_ranges:  - PT6H  - PT24H  - P7D  - P14D  - P4W  - P3M  - P12M  - rill-TD  - rill-WTD  - rill-MTD  - rill-QTD  - rill-YTD  - rill-PDC  - rill-PWC  - rill-PMC  - rill-PQC  - rill-PYCtime_zones:  - UTC  - America/Los_Angeles  - America/Chicago  - America/New_York  - Europe/London  - Europe/Paris  - Asia/Jerusalem  - Europe/Moscow  - Asia/Kolkata  - Asia/Shanghai  - Asia/Tokyo  - Australia/Sydney',
      );
  });
});
