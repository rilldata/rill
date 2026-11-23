import { expect, type Page } from "@playwright/test";
import { test } from "../setup/base";
import { waitForReconciliation } from "../utils/wait-for-reconciliation.ts";

const METRICS_VIEW = "AdBids_precision_metrics";

// The AdBids_precision_metrics metrics view defines `publishers` with a default
// expression (`count(DISTINCT publisher) + 1000`) and an exact variant
// (`count(DISTINCT publisher)`). The two deliberately differ so the test can
// tell which one answered: the default value is 1,004, the exact value 4
// (Facebook, Google, Yahoo, Microsoft; nulls are not counted).

// Waits for the next aggregation query against the metrics view whose JSON
// body satisfies `matches`. The explore calls the runtime's Connect endpoint,
// which carries the metrics view in the body, not in the URL.
function waitForAggregation(page: Page, matches: (body: string) => boolean) {
  return page.waitForRequest((request) => {
    const body = request.postData() ?? "";
    return (
      request.url().includes("QueryService/MetricsViewAggregation") &&
      body.includes(`"metricsView":"${METRICS_VIEW}"`) &&
      matches(body)
    );
  });
}

test.describe("exact measures", () => {
  test.use({ project: "AdBids" });

  test("switches a measure to exact on click and back on a context change", async ({
    page,
  }) => {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    await waitForReconciliation(page);
    await page.goto(
      `${baseUrl}/explore/AdBids_precision_metrics_explore?measures=publishers,total_records`,
    );

    const card = page.locator(".big-number").filter({ hasText: "Publishers" });
    const chip = card.getByRole("button", {
      name: "Toggle exact value for Publishers",
    });

    // Approximate by default.
    await expect(chip).toHaveText("≈");
    await expect(chip).toHaveAttribute("aria-pressed", "false");
    await expect(card.getByText("1,004")).toBeVisible();

    // One click requests the exact value for the measure, and only for it.
    const exactRequest = waitForAggregation(page, (body) =>
      body.includes('"name":"publishers","exact":true'),
    );
    await chip.click();
    const request = await exactRequest;
    expect(request.postData()).not.toContain(
      '"name":"total_records","exact":true',
    );

    await expect(chip).toHaveText("=");
    await expect(chip).toHaveAttribute("aria-pressed", "true");
    await expect(card.getByText("4", { exact: true })).toBeVisible();
    await expect(card.getByText("1,004")).toBeHidden();

    // The exact selection is tab memory: it is not written to the URL.
    expect(new URL(page.url()).searchParams.get("exact")).toBeNull();

    // A query context change (opening the dimension table) resets the measure
    // to approximate; the chip in the dimension table header shows it.
    const approximateRequest = waitForAggregation(
      page,
      (body) =>
        body.includes('"name":"publishers"') && !body.includes('"exact":true'),
    );
    await page
      .getByRole("button", { name: "Open dimension details" })
      .first()
      .click();
    await approximateRequest;

    // The chip's aria label is part of the accessible name of the card and of
    // the column header button around it, hence the exact name match.
    const tableChip = page
      .getByRole("table", { name: "Dimension table" })
      .getByRole("button", {
        name: "Toggle exact value for Publishers",
        exact: true,
      });
    await expect(tableChip).toHaveText("≈");
    await expect(tableChip).toHaveAttribute("aria-pressed", "false");
  });

  // `records_per_publisher` is calculated from `publishers`. Derived measures
  // have no exact variant of their own, so the card shows no chip; the ratio is
  // computed from the measures as the query computes them.
  test("a derived measure has no precision chip", async ({ page }) => {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    await waitForReconciliation(page);
    await page.goto(
      `${baseUrl}/explore/AdBids_precision_metrics_explore?measures=publishers,total_records,records_per_publisher`,
    );

    const publishersCard = page
      .locator(".big-number")
      .filter({ hasText: "Publishers" })
      .first();
    await expect(
      publishersCard.getByRole("button", {
        name: "Toggle exact value for Publishers",
      }),
    ).toHaveText("≈");

    const ratioCard = page
      .locator(".big-number")
      .filter({ hasText: "Records per publisher" });
    await expect(ratioCard).toBeVisible();
    await expect(
      ratioCard.getByRole("button", {
        name: "Toggle exact value for Records per publisher",
      }),
    ).toHaveCount(0);
  });

  // Every leaderboard, the big numbers, the dimension table and the pivot are
  // separate surfaces: a chip only switches the surface it sits in.
  test("a leaderboard chip switches only that leaderboard", async ({
    page,
  }) => {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    await waitForReconciliation(page);
    await page.goto(
      `${baseUrl}/explore/AdBids_precision_metrics_explore?measures=publishers,total_records&leaderboard_measures=publishers&sort_by=publishers`,
    );

    const leaderboard = (dimension: string) =>
      page.locator("table", {
        has: page.locator("[data-dimension-header]", { hasText: dimension }),
      });
    const publisherChip = leaderboard("Publisher").getByRole("button", {
      name: "Toggle exact value for Publishers",
    });
    const domainChip = leaderboard("Domain").getByRole("button", {
      name: "Toggle exact value for Publishers",
    });
    const cardChip = page
      .locator(".big-number")
      .filter({ hasText: "Publishers" })
      .first()
      .getByRole("button", { name: "Toggle exact value for Publishers" });

    await expect(publisherChip).toHaveText("≈");

    const exactRequest = waitForAggregation(
      page,
      (body) =>
        body.includes('"name":"publishers","exact":true') &&
        body.includes('"dimensions":[{"name":"publisher"}]'),
    );
    await publisherChip.click();
    await exactRequest;

    await expect(publisherChip).toHaveText("=");
    await expect(publisherChip).toHaveAttribute("aria-pressed", "true");
    await expect(domainChip).toHaveText("≈");
    await expect(cardChip).toHaveText("≈");
    await expect(cardChip).toHaveAttribute("aria-pressed", "false");
  });
});
