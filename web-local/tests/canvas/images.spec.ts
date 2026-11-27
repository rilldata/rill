import { expect } from "@playwright/test";
import { waitForReconciliation } from "web-local/tests/utils/wait-for-reconciliation";
import { test } from "../setup/base";

test.describe("canvas images and banner", () => {
  test.use({ project: "CanvasAssets" });

  test("Should show images from a subfolder of the public folder, and the banner's formatting", async ({
    page,
  }) => {
    await waitForReconciliation(page);
    await page.goto(`${new URL(page.url()).origin}/canvas/branded`);

    // The image component and the image in the markdown component both load the SVG from the project.
    const images = page.locator("img[src^='blob:']");
    await expect(images).toHaveCount(2);
    for (const image of await images.all()) {
      await expect
        .poll(() => image.evaluate((img: HTMLImageElement) => img.naturalWidth))
        .toBeGreaterThan(0);
    }

    const banner = page.locator(".app-banner");
    await expect(banner.locator("strong")).toHaveText("Heads up:");
    await expect(
      banner.getByRole("link", { name: "status page" }),
    ).toHaveAttribute("target", "_blank");
  });
});
