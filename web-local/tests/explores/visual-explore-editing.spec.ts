import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { gotoNavEntry } from "../utils/waitHelpers";

test.describe("visual explore editing", () => {
  test.use({ project: "AdBids" });

  test("visual explore editor runthrough", async ({ page }) => {
    test.setTimeout(45_000); // Note: we should make this test smaller!

    await page.getByLabel("/dashboards").click();
    await page.waitForTimeout(1000);
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    await page.getByRole("button", { name: "switch to code editor" }).click();

    await page.getByRole("button", { name: "Subset" }).first().click();
    await page.getByRole("button", { name: "Subset" }).nth(1).click();

    await page.getByRole("button", { name: "Custom" }).first().click();
    await page.getByRole("button", { name: "Custom" }).nth(1).click();
    await page.getByRole("button", { name: "Custom" }).nth(2).click();

    let text = await page
      .getByRole("textbox", { name: "codemirror editor" })
      .textContent();

    expect(text).toEqual(
      ' 5212345678910111213141516171819202122232425262728293031323334353637383940414243444546474849505152# Explore YAML# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboardstype: exploretitle: "Adbids dashboard"metrics_view: AdBids_metricsdimensions:  - timestamp  - publisher  - domainmeasures:  - total_records  - bid_price_sumtime_ranges:  - PT6H  - PT24H  - P7D  - P14D  - P4W  - P12M  - rill-TD  - rill-WTD  - rill-MTD  - rill-QTD  - rill-YTD  - rill-PDC  - rill-PWC  - rill-PMC  - rill-PQC  - rill-PYC  - inftime_zones:  - UTC  - America/Los_Angeles  - America/Chicago  - America/New_York  - Europe/London  - Europe/Paris  - Asia/Jerusalem  - Europe/Moscow  - Asia/Kolkata  - Asia/Shanghai  - Asia/Tokyo  - Australia/Sydneytheme:  light:    primary: hsl(180, 100%, 50%)    secondary: lightgreen',
    );

    await page.getByRole("button", { name: "Expression" }).first().click();
    await page.getByRole("button", { name: "Expression" }).nth(1).click();

    await page.getByRole("button", { name: "Default" }).first().click();
    await page.getByRole("button", { name: "Presets" }).click();

    text = await page
      .getByRole("textbox", { name: "codemirror editor" })
      .textContent();

    expect(text).toEqual(
      ' 45123456789101112131415161718192021222324252627282930313233343536373839404142434445# Explore YAML# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboardstype: exploretitle: "Adbids dashboard"metrics_view: AdBids_metricsdimensions:  expr: "*"measures:  expr: "*"time_ranges:  - PT6H  - PT24H  - P7D  - P14D  - P4W  - P3M  - P12M  - rill-TD  - rill-WTD  - rill-MTD  - rill-QTD  - rill-YTD  - rill-PDC  - rill-PWC  - rill-PMC  - rill-PQC  - rill-PYCtime_zones:  - UTC  - America/Los_Angeles  - America/Chicago  - America/New_York  - Europe/London  - Europe/Paris  - Asia/Jerusalem  - Europe/Moscow  - Asia/Kolkata  - Asia/Shanghai  - Asia/Tokyo  - Australia/Sydney',
    );
  });
});
