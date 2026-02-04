import { expect, type Page } from "@playwright/test";
import { test } from "../setup/base";
import { gotoNavEntry } from "../utils/waitHelpers";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions";

async function setupDashboard(page: Page) {
  await page.getByLabel("/dashboards").click();
  await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");

  const bigNumber = page
    .locator(".big-number")
    .filter({ hasText: "Total records" });
  await expect(bigNumber).toBeVisible({ timeout: 10_000 });

  await page.getByRole("button", { name: "Preview" }).click();
  await expect(bigNumber).toBeVisible({ timeout: 10_000 });

  await interactWithTimeRangeMenu(page, async () => {
    await page.getByRole("menuitem", { name: "All Time" }).click();
  });
  await page.waitForTimeout(1000);

  const valueLocator = bigNumber.locator('div[role="button"]');
  await expect(valueLocator).toBeVisible({ timeout: 5000 });

  const chartSvg = page.locator('svg[aria-label*="Measure Chart"]').first();
  await expect(chartSvg).toBeVisible({ timeout: 5000 });
  const box = await chartSvg.boundingBox();
  expect(box).toBeTruthy();

  return { valueLocator, chartSvg, box: box! };
}

/** Drag across startPct–endPct (0–1) of chart width to create a scrub selection. */
async function scrub(
  page: Page,
  box: { x: number; y: number; width: number; height: number },
  startPct: number,
  endPct: number,
) {
  const startX = box.x + box.width * startPct;
  const endX = box.x + box.width * endPct;
  const centerY = box.y + box.height / 2;

  await page.mouse.move(startX, centerY);
  await page.mouse.down();
  await page.mouse.move(endX, centerY, { steps: 15 });
  await page.mouse.up();
  await page.waitForTimeout(1500);
}

test.describe("chart scrub and zoom", () => {
  test.use({ project: "AdBids" });

  test("scrub selection updates big number, zoom changes time range", async ({
    page,
  }) => {
    const { valueLocator, box } = await setupDashboard(page);
    const initialValue = await valueLocator.textContent();

    await scrub(page, box, 0.2, 0.6);

    const scrubValue = await valueLocator.textContent();
    expect(scrubValue).toBeTruthy();
    expect(scrubValue).not.toBe(initialValue);

    await expect(page.getByLabel("Zoom")).toBeVisible({ timeout: 3000 });

    await page.keyboard.press("z");
    await page.waitForTimeout(1500);

    await expect(page.getByLabel("Undo zoom")).toBeVisible({ timeout: 3000 });
    const timeRangeText = await page
      .getByLabel("Select time range")
      .textContent();
    expect(timeRangeText).toContain("Custom");

    const zoomedValue = await valueLocator.textContent();
    expect(zoomedValue).toBe(scrubValue);
  });

  test("move scrub range updates big number", async ({ page }) => {
    const { valueLocator, box } = await setupDashboard(page);

    await scrub(page, box, 0.2, 0.5);
    const scrubValue = await valueLocator.textContent();
    expect(scrubValue).toBeTruthy();

    // Grab center of selection (35%) and drag right to 65%
    const grabX = box.x + box.width * 0.35;
    const dropX = box.x + box.width * 0.65;
    const centerY = box.y + box.height / 2;

    await page.mouse.move(grabX, centerY);
    await page.mouse.down();
    await page.mouse.move(dropX, centerY, { steps: 15 });
    await page.mouse.up();
    await page.waitForTimeout(1500);

    const movedValue = await valueLocator.textContent();
    expect(movedValue).toBeTruthy();
    expect(movedValue).not.toBe(scrubValue);
    await expect(page.getByLabel("Zoom")).toBeVisible({ timeout: 3000 });
  });

  test("resize scrub start edge updates big number", async ({ page }) => {
    const { valueLocator, box } = await setupDashboard(page);

    await scrub(page, box, 0.3, 0.7);
    const scrubValue = await valueLocator.textContent();
    expect(scrubValue).toBeTruthy();

    // Drag left edge from 30% to 10%
    const edgeX = box.x + box.width * 0.3;
    const newEdgeX = box.x + box.width * 0.1;
    const centerY = box.y + box.height / 2;

    await page.mouse.move(edgeX, centerY);
    await page.mouse.down();
    await page.mouse.move(newEdgeX, centerY, { steps: 15 });
    await page.mouse.up();
    await page.waitForTimeout(1500);

    const resizedValue = await valueLocator.textContent();
    expect(resizedValue).toBeTruthy();
    expect(resizedValue).not.toBe(scrubValue);
    await expect(page.getByLabel("Zoom")).toBeVisible({ timeout: 3000 });
  });

  test("resize scrub end edge updates big number", async ({ page }) => {
    const { valueLocator, box } = await setupDashboard(page);

    await scrub(page, box, 0.2, 0.5);
    const scrubValue = await valueLocator.textContent();
    expect(scrubValue).toBeTruthy();

    // Drag right edge from 50% to 80%
    const edgeX = box.x + box.width * 0.5;
    const newEdgeX = box.x + box.width * 0.8;
    const centerY = box.y + box.height / 2;

    await page.mouse.move(edgeX, centerY);
    await page.mouse.down();
    await page.mouse.move(newEdgeX, centerY, { steps: 15 });
    await page.mouse.up();
    await page.waitForTimeout(1500);

    const resizedValue = await valueLocator.textContent();
    expect(resizedValue).toBeTruthy();
    expect(resizedValue).not.toBe(scrubValue);
    await expect(page.getByLabel("Zoom")).toBeVisible({ timeout: 3000 });
  });
});
