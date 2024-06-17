import { createDashboardFromModel } from "web-local/tests/utils/dashboardHelpers";
import { createAdBidsModel } from "web-local/tests/utils/dataSpecifcHelpers";
import { test } from "../utils/test";
import { waitForFileNavEntry } from "../utils/waitHelpers";

export function useDashboardFlowTestSetup() {
  test.beforeEach(async ({ page }) => {
    test.setTimeout(30000);
    await createAdBidsModel(page);

    await Promise.all([
      waitForFileNavEntry(
        page,
        `/dashboards/AdBids_model_dashboard.yaml`,
        true,
      ),
      createDashboardFromModel(page, "/models/AdBids_model.sql"),
    ]);
  });
}
