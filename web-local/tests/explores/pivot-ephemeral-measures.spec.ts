import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { validateTableContents } from "../utils/tableHelpers";
import { waitForReconciliation } from "../utils/wait-for-reconciliation.ts";

// ephemeral measures are encoded in the `adhoc_m` URL param, so a
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
      `${baseUrl}/explore/AdBids_metrics_explore?view=pivot&rows=publisher&cols=total_records,doubled&adhoc_m=doubled:Doubled:total_records*2`,
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
    expect(url.searchParams.get("adhoc_m")).toBe(
      "doubled:Doubled:total_records*2",
    );
  });

  test("renders an ephemeral measure in the explore view", async ({ page }) => {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    await waitForReconciliation(page);

    await page.goto(
      `${baseUrl}/explore/AdBids_metrics_explore?measures=total_records,doubled&adhoc_m=doubled:Doubled:total_records*2`,
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
    expect(url.searchParams.get("adhoc_m")).toBe(
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
      `${baseUrl}/explore/AdBids_metrics_explore?view=pivot&rows=publisher&cols=total_records,bad&adhoc_m=bad:Bad:unknown_measure*2`,
    );

    // The invalid definition and its column are dropped; the rest renders.
    await expect(
      page.getByLabel("Total records pivot chip", { exact: true }),
    ).toBeVisible();
    await expect(
      page.getByLabel("Bad pivot chip", { exact: true }),
    ).toHaveCount(0);
  });

  test("inserts a measure from the @ picker in the expression input", async ({
    page,
  }) => {
    const currentUrl = new URL(page.url());
    const baseUrl = `${currentUrl.protocol}//${currentUrl.host}`;

    await waitForReconciliation(page);

    await page.goto(
      `${baseUrl}/explore/AdBids_metrics_explore?view=pivot&rows=publisher&cols=total_records`,
    );

    await page.getByRole("button", { name: "Create adhoc measure" }).click();
    await page.getByLabel("Name", { exact: true }).fill("Doubled");

    // Typing "@" opens a picker searchable by display name; the field name is
    // shown alongside so users learn it.
    const expression = page.getByLabel("Expression", { exact: true });
    await expression.pressSequentially("@records");
    const picker = page.getByRole("listbox", { name: "Measures" });
    await expect(
      picker.getByRole("option", { name: "Total records total_records" }),
    ).toBeVisible();
    await expect(
      picker.getByRole("option", { name: "Sum of Bid Price bid_price_sum" }),
    ).toHaveCount(0);
    // The "@" is not a valid expression character, but no error shows while
    // the picker is open.
    await expect(page.getByText('unexpected character "@"')).toHaveCount(0);

    // Enter replaces the "@query" with a chip showing the display name, which
    // serializes to the measure's field name in the saved expression.
    await expression.press("Enter");
    await expect(picker).toHaveCount(0);
    await expect(
      expression.locator(".measure-chip", { hasText: "Total records" }),
    ).toBeVisible();

    await expression.pressSequentially("* 2");
    await page.getByLabel("Description").fill("Twice");
    await page.getByRole("button", { name: "Save" }).click();

    await expect(
      page.getByLabel("Doubled pivot chip", { exact: true }),
    ).toBeVisible();
    // The description rides along in the URL after the (empty) format preset.
    const url = new URL(page.url());
    expect(url.searchParams.get("adhoc_m")).toBe(
      "doubled:Doubled:total_records%20*%202::Twice",
    );

    // A second measure with the same display name is rejected.
    await page.getByRole("button", { name: "Create adhoc measure" }).click();
    await page.getByLabel("Name", { exact: true }).fill("doubled");
    await expect(
      page.getByText('a measure named "doubled" already exists'),
    ).toBeVisible();
    await expect(page.getByRole("button", { name: "Save" })).toBeDisabled();
  });
});
