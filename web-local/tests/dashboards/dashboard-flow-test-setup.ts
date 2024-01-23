import { expect, test } from "@playwright/test";
import { createDashboardFromModel } from "web-local/tests/utils/dashboardHelpers";
import { createAdBidsModel } from "web-local/tests/utils/dataSpecifcHelpers";

export function useDashboardFlowTestSetup() {
  test.beforeEach(async ({ page }) => {
    test.setTimeout(60000);
    await page.goto("/");
    // disable animations
    await page.addStyleTag({
      content: `
        *, *::before, *::after {
          animation-duration: 0s !important;
          transition-duration: 0s !important;
        }
      `,
    });
    await createAdBidsModel(page);
    await createDashboardFromModel(page, "AdBids_model");

    // Change time zone to UTC
    await page.getByLabel("Timezone selector").click();
    await page
      .getByRole("menuitem", { name: "UTC GMT +00:00 Etc/UTC" })
      .click();
    // Wait for menu to close
    await expect(
      page.getByRole("menuitem", { name: "UTC GMT +00:00 Etc/UTC" }),
    ).not.toBeVisible();
  });
}
