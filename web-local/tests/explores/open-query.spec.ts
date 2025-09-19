import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("Query-to-Explore routing", () => {
  test.use({ project: "AdBids" });

  test("Can open a complex metrics view query", async ({ page }) => {
    // Get the current URL to construct the correct baseURL with the dynamic port
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    // Open the metrics view query
    await page.goto(
      `${baseUrl}/-/open-query?query=%7B%22dimensions%22%3A%5B%7B%22name%22%3A%22publisher%22%7D%5D%2C%22measures%22%3A%5B%7B%22name%22%3A%22bid_price_sum%22%7D%2C%7B%22name%22%3A%22total_records%22%7D%5D%2C%22metrics_view%22%3A%22AdBids_metrics%22%2C%22time_range%22%3A%7B%22end%22%3A%222022-03-30T23%3A59%3A54.56Z%22%2C%22start%22%3A%222022-01-01T00%3A02%3A39.041Z%22%7D%2C%22where%22%3A%7B%22cond%22%3A%7B%22exprs%22%3A%5B%7B%22name%22%3A%22publisher%22%7D%2C%7B%22val%22%3A%22Facebook%22%7D%5D%2C%22op%22%3A%22eq%22%7D%7D%7D`,
    );

    // Wait a bit for any redirects to happen
    await page.waitForTimeout(2000);

    // Expect to get redirected to the explore page with a rich stateful URL
    expect(page.url()).toContain(
      "explore/AdBids_metrics_explore?tr=2022-01-01T00%3A02%3A39.041Z%2C2022-03-30T23%3A59%3A54.560Z&f=publisher+IN+%28%27Facebook%27%29&measures=bid_price_sum%2Ctotal_records&dims=publisher&expand_dim=publisher",
    );
  });

  // Note that this is not exhaustive. This will be iterated over quite a bit. So more tests would be added to the mapper directly.
});
