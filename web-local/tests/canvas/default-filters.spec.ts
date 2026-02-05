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
      ` 68123456789101112131415161718192021222324252627282930313233343536# Explore YAML# Reference documentation: https://docs.rilldata.com/reference/project-files/canvas-dashboardstype: canvasdisplay_name: "Adbids Canvas Dashboard"defaults:  time_range: PT24H  comparison_mode: time  filters:    AdBids_metrics: (domain IN ('facebook.com', 'google.com', 'msn.com'))rows:  - items:      - kpi_grid:          metrics_view: AdBids_metrics          measures:            - total_records            - bid_price_sum          comparison:            - delta            - percent_change        width: 12    height: 128px  - items:      - stacked_bar:          metrics_view: AdBids_metrics          x:            type: temporal            field: timestamp            sort: -y            limit: 20          y:            type: quantitative            field: total_records            zeroBasedOrigin: true          color: hsl(240,100%,67%)        width: 12`,
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

  test("legacy filters without prefix still work", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");
    const currentUrl = new URL(page.url());

    // Clear search params from current url
    currentUrl.search = "";

    await page.goto(
      `${currentUrl}?tr=PT24H&compare_tr=rill-PP&f=domain+IN+%28%5B%27facebook.com%27%2C%27google.com%27%2C%27msn.com%27%5D%29`,
    );

    // check that a filter pill exists with the correct text
    await expect(page.getByText("Domain facebook.com +2 others")).toBeVisible();

    // check that filters are applied
    await expect(page.locator(".kpi-wrapper").getByText("797")).toBeVisible();
  });
});
