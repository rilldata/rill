import { test } from "@playwright/test";
import { createDashboardFromModel } from "web-local/test/utils/dashboardHelpers";
import { createAdBidsModel } from "web-local/test/utils/dataSpecifcHelpers";

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
  });
}
