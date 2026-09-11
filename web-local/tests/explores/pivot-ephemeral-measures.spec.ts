import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { validateTableContents } from "../utils/tableHelpers";
import { waitForReconciliation } from "../utils/wait-for-reconciliation.ts";

// ephemeral measures are encoded in the `ephemeral` URL param, so a
// shared pivot URL reproduces them without any project or YAML changes.
// Definitions can be created in the pivot UI or arrive via a shared URL.
test.describe("pivot ephemeral measures from URL state", () => {
  test.use({ project: "AdBids" });

  test("renders an ephemeral measure column from a shared URL", async ({
    page,
  }) => {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    await waitForReconciliation(page);

    await page.goto(
      `${baseUrl}/explore/AdBids_metrics_explore?view=pivot&rows=publisher&cols=total_records,doubled&ephemeral=doubled:Doubled:total_records*2`,
    );

    // The ephemeral measure renders as a chip in the columns zone...
    await expect(
      page.getByLabel("Doubled pivot chip", { exact: true }),
    ).toBeVisible();

    // ...and as a computed column in the table.
    // Raw counts: total 100000, null 32897, Facebook 19341, Google 18763,
    // Yahoo 18593, Microsoft 10406; "Doubled" is each count * 2, humanized.
    await validateTableContents(page, "table", [
      [], // dummy row added for virtualization
      ["Total", "100.0k", "200.0k"],
      ["null", "32.9k", "65.8k"],
      ["Facebook", "19.3k", "38.7k"],
      ["Google", "18.8k", "37.5k"],
      ["Yahoo", "18.6k", "37.2k"],
      ["Microsoft", "10.4k", "20.8k"],
      [], // dummy row added for virtualization
    ]);

    // The definition survives in the URL for sharing.
    const url = new URL(page.url());
    expect(url.searchParams.get("ephemeral")).toBe(
      "doubled:Doubled:total_records*2",
    );
  });

  test("renders an ephemeral measure in the explore view", async ({ page }) => {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    await waitForReconciliation(page);

    await page.goto(
      `${baseUrl}/explore/AdBids_metrics_explore?measures=total_records,doubled&ephemeral=doubled:Doubled:total_records*2`,
    );

    // The big number renders the ephemeral measure with its computed total
    // (total_records is 100k, so Doubled is 200k)...
    await expect(
      page.getByRole("button", { name: "Doubled 200k" }),
    ).toBeVisible();
    // ...and the time-series chart renders via the timeseries API's
    // expression measures.
    await expect(
      page.getByRole("img", { name: "Measure Chart for doubled" }),
    ).toBeVisible();
  });

  test("creates an ephemeral measure from the pivot sidebar", async ({
    page,
  }) => {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    await waitForReconciliation(page);

    await page.goto(
      `${baseUrl}/explore/AdBids_metrics_explore?view=pivot&rows=publisher&cols=total_records`,
    );

    // The Measures section of the field sidebar offers a create CTA.
    await page.getByRole("button", { name: "Create adhoc measure" }).click();

    await page.getByLabel("Name", { exact: true }).fill("Doubled");
    await page
      .getByLabel("Expression", { exact: true })
      .fill("total_records * 2");
    await page.getByRole("button", { name: "Save" }).click();

    // Saving from the pivot view places the measure as a column chip...
    await expect(
      page.getByLabel("Doubled pivot chip", { exact: true }),
    ).toBeVisible();

    // ...renders its computed values...
    await validateTableContents(page, "table", [
      [], // dummy row added for virtualization
      ["Total", "100.0k", "200.0k"],
      ["null", "32.9k", "65.8k"],
      ["Facebook", "19.3k", "38.7k"],
      ["Google", "18.8k", "37.5k"],
      ["Yahoo", "18.6k", "37.2k"],
      ["Microsoft", "10.4k", "20.8k"],
      [], // dummy row added for virtualization
    ]);

    // ...and writes the definition into the shareable URL. Display name and
    // expression are URI-encoded per field by the ephemeral param grammar, so the
    // param value keeps that layer after URLSearchParams decoding.
    const url = new URL(page.url());
    expect(url.searchParams.get("ephemeral")).toBe(
      "doubled:Doubled:total_records%20*%202",
    );
  });

  test("drops invalid ephemeral measure definitions with an error", async ({
    page,
  }) => {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    await waitForReconciliation(page);

    await page.goto(
      `${baseUrl}/explore/AdBids_metrics_explore?view=pivot&rows=publisher&cols=total_records,bad&ephemeral=bad:Bad:unknown_measure*2`,
    );

    // The invalid definition and its column are dropped; the rest renders.
    await expect(
      page.getByLabel("Total records pivot chip", { exact: true }),
    ).toBeVisible();
    await expect(
      page.getByLabel("Bad pivot chip", { exact: true }),
    ).toHaveCount(0);
  });
});
