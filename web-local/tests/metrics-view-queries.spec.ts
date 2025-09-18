import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Metrics view queries", () => {
  test.use({ project: "AdBids" });

  test("Can open a complex metrics view query", async ({ page }) => {
    // Open the metrics view query
    await page.goto(
      "/-/open-query?query=%7B%22metrics_view%22%3A%22AdBids_metrics%22%2C%22filters%22%3A%5B%7B%22column%22%3A%22bid_price%22%2C%22operator%22%3A%22%3E%22%2C%22value%22%3A%22100%22%7D%5D%7D",
    );

    // Expect to get redirected to the explore page with a rich stateful URL
    expect(page.url()).toContain(
      "/explore/AdBids_explore?filters=bid_price%3E100",
    );
  });

  // Note that this is not exhaustive. This will be iterated over quite a bit. So more tests would be added to the mapper directly.
});
