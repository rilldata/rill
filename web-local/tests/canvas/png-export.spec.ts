import { expect } from "@playwright/test";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";

test.describe("canvas PNG export", () => {
  test.use({ project: "AdBids" });

  test("downloads a component as a PNG from its toolbar menu", async ({
    page,
  }) => {
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");

    const card = page.locator("#AdBids_metrics_canvas--component-1-0");
    await card.locator("canvas").waitFor({ state: "visible" });

    // The toolbar only shows while the card is hovered.
    await card.hover();
    await card.getByRole("button", { name: "Component menu" }).click();
    await page.getByRole("menuitem", { name: "Download as PNG" }).click();

    const dialog = page.getByRole("dialog");
    await expect(dialog.getByText("Download as PNG")).toBeVisible();
    // The preview renders the component a second time, framed by the
    // dashboard's title.
    await expect(
      dialog.locator("#canvas-png-export-frame canvas"),
    ).toBeVisible();

    // Start waiting for the download before clicking. Note no await.
    const downloadPromise = page.waitForEvent("download");
    await dialog.getByRole("button", { name: "Download PNG" }).click();
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/^.+-\d{8}-\d{6}\.png$/);
  });
});
