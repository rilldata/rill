import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { waitForReconciliation } from "../utils/wait-for-reconciliation.ts";

// A filter the chips cannot show, such as a top level OR, falls back to the read only
// "Advanced (BETA)" pill. Clearing it has to work from the UI, and the cleared state has to be
// what the session restores afterwards, since an empty explore url loads the last state of the tab.
test.describe("advanced filters", () => {
  test.use({ project: "AdBids" });

  test("clears an advanced filter from a shared url and keeps it cleared on reload", async ({
    page,
  }) => {
    const { origin } = new URL(page.url());
    await waitForReconciliation(page);

    const filter = "publisher IN ('Facebook') OR domain IN ('google.com')";
    await page.goto(
      `${origin}/explore/AdBids_metrics_explore?f=${encodeURIComponent(filter)}`,
    );
    await expect(page.getByText("Advanced (BETA)")).toBeVisible();
    await expect(page.getByRole("button", { name: "Remove" })).toBeVisible();
    await expect(page.getByLabel("Add filter button")).toBeHidden();

    await page.getByRole("button", { name: "Clear filters" }).click();
    await expect(page.getByText("No filters selected")).toBeVisible();
    await expect(page.getByLabel("Add filter button")).toBeVisible();
    await page.waitForURL((url) => !url.searchParams.has("f"));

    await page.reload();
    await expect(page.getByText("No filters selected")).toBeVisible();
    expect(new URL(page.url()).searchParams.has("f")).toBe(false);
  });
});
