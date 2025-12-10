import { expect } from "@playwright/test";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";

test.describe("canvas time filters", () => {
  test.use({ project: "AdBids" });

  test("save default filters button works", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");

    await page.getByRole("button", { name: "Options" }).click();
    await page.getByRole("button", { name: "Add filter button" }).click();
    await page.getByRole("menuitem", { name: "Domain" }).click();

    await page.getByLabel("domain results").getByText("facebook.com").click();

    await page
      .getByLabel("domain results")
      .getByText("google.com", { exact: true })
      .click();
    await page.getByLabel("domain results").getByText("msn.com").click();

    await page
      .getByRole("button", { name: "Open domain filter" })
      .first()
      .click();

    await page.getByRole("button", { name: "Save as default" }).click();

    await page.waitForSelector('button:has-text("Saved default filters")');

    await page.waitForSelector('button:has-text("Viewing default state")');

    expect(page.url()).toContain(
      "?tr=PT24H&compare_tr=rill-PP&f.AdBids_metrics=domain+IN+%28%5B%27facebook.com%27%2C%27google.com%27%2C%27msn.com%27%5D%29",
    );

    // navigate to code view
    await page.getByRole("button", { name: "switch to code editor" }).click();

    const text = await page
      .getByRole("textbox", { name: "codemirror editor" })
      .textContent();

    expect(text).toEqual(
      ' 5212345678910111213141516171819202122232425262728293031323334353637383940414243444546474849505152# Explore YAML# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboardstype: exploretitle: \"Adbids dashboard\"metrics_view: AdBids_metricsdimensions:  - timestamp  - publisher  - domainmeasures:  - total_records  - bid_price_sumtime_ranges:  - PT6H  - PT24H  - P7D  - P14D  - P4W  - P12M  - rill-TD  - rill-WTD  - rill-MTD  - rill-QTD  - rill-YTD  - rill-PDC  - rill-PWC  - rill-PMC  - rill-PQC  - rill-PYC  - inftime_zones:  - UTC  - America/Los_Angeles  - America/Chicago  - America/New_York  - Europe/London  - Europe/Paris  - Asia/Jerusalem  - Europe/Moscow  - Asia/Kolkata  - Asia/Shanghai  - Asia/Tokyo  - Australia/Sydneytheme:  light:    primary: hsl(180, 100%, 50%)    secondary: lightgreen',
    );
  });

  test("default filters load", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");

    await page.getByRole("button", { name: "Options" }).click();
    await page.getByRole("button", { name: "Add filter button" }).click();
    await page.getByRole("menuitem", { name: "Domain" }).click();

    await page.getByLabel("domain results").getByText("facebook.com").click();

    await page
      .getByLabel("domain results")
      .getByText("google.com", { exact: true })
      .click();
    await page.getByLabel("domain results").getByText("msn.com").click();

    await page
      .getByRole("button", { name: "Open domain filter" })
      .first()
      .click();

    await page.getByRole("button", { name: "Save as default" }).click();

    await page.waitForSelector('button:has-text("Saved default filters")');

    await page.waitForSelector('button:has-text("Viewing default state")');
    const currentUrl = new URL(page.url());

    // Clear search params from current url
    currentUrl.search = "";

    await page.goto(`${currentUrl}`);

    await page.waitForTimeout(1000);
    expect(page.url()).toContain(
      "?tr=PT24H&compare_tr=rill-PP&f.AdBids_metrics=domain+IN+%28%5B%27facebook.com%27%2C%27google.com%27%2C%27msn.com%27%5D%29",
    );
  });
});
