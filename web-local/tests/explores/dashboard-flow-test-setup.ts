import { createMetricsViewFromModel } from "web-local/tests/utils/metricsViewHelpers";
import { createAdBidsModel } from "web-local/tests/utils/dataSpecifcHelpers";
import { test } from "../utils/test";
import { waitForFileNavEntry } from "../utils/waitHelpers";

export function useDashboardFlowTestSetup() {
  test.beforeEach(async ({ page }) => {
    test.setTimeout(45000);
    await createAdBidsModel(page);

    await Promise.all([
      waitForFileNavEntry(
        page,
        `/dashboards/AdBids_model_dashboard.yaml`,
        true,
      ),
      createMetricsViewFromModel(page, "/models/AdBids_model.sql"),
    ]);
  });
}
