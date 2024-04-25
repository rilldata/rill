import { createDashboardFromModel } from "web-local/tests/utils/dashboardHelpers";
import { createAdBidsModel } from "web-local/tests/utils/dataSpecifcHelpers";
import { test } from "../utils/test";

export function useDashboardFlowTestSetup() {
  test.beforeEach(async ({ page }) => {
    test.setTimeout(60000);
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
    await createDashboardFromModel(page, "/models/AdBids_model.sql");
  });
}
