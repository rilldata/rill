import { expect } from "@playwright/test";
import { createExploreFromModel } from "web-local/tests/utils/exploreHelpers";
import {
  clickMenuButton,
  openFileNavEntryContextMenu,
} from "../utils/commonHelpers";
import { createAdBidsModel } from "../utils/dataSpecifcHelpers";
import { test } from "../utils/test";

test.describe("visual explore editing", () => {
  test("visual explore editor runthrough", async ({ page }) => {
    await createAdBidsModel(page);
    await createExploreFromModel(page, false);

    await openFileNavEntryContextMenu(
      page,
      "/metrics/AdBids_model_metrics.yaml",
    );
    await clickMenuButton(page, "Generate dashboard");

    await page.getByRole("button", { name: "Edit" }).click();
    await page.getByRole("menuitem", { name: "Explore" }).click();
    await page.getByLabel("code").click();

    await page.getByRole("button", { name: "Subset" }).first().click();
    await page.getByRole("button", { name: "Subset" }).nth(1).click();

    await page.getByRole("button", { name: "Custom" }).first().click();
    await page.getByRole("button", { name: "Custom" }).nth(1).click();
    await page.getByRole("button", { name: "Custom" }).nth(2).click();

    const text = await page
      .getByRole("textbox", { name: "Code editor" })
      .textContent();

    expect(text).toEqual(
      ' 48123456789101112131415161718192021222324252627282930313233343536373839404142434445464748# Explore YAML# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboardstype: exploretitle: "Adbids Model dashboard"metrics_view: AdBids_model_metricsdimensions:  - publisher  - domainmeasures:  - total_records  - bid_price_sumtime_zones:  - America/Los_Angeles  - America/Chicago  - America/New_York  - Europe/London  - Europe/Paris  - Asia/Jerusalem  - Europe/Moscow  - Asia/Kolkata  - Asia/Shanghai  - Asia/Tokyo  - Australia/Sydneytime_ranges:  - PT6H  - PT24H  - P7D  - P14D  - P4W  - P12M  - rill-TD  - rill-WTD  - rill-MTD  - rill-QTD  - rill-YTD  - rill-PDC  - rill-PWC  - rill-PMC  - rill-PQC  - rill-PYCtheme:  colors:    primary: hsl(180, 100%, 50%)    secondary: lightgreen',
    );
  });
});
