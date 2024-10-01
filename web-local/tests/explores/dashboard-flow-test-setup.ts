import { createExploreFromModel } from "web-local/tests/utils/exploreHelpers";
import { createAdBidsModel } from "web-local/tests/utils/dataSpecifcHelpers";
import { test } from "../utils/test";

export function useDashboardFlowTestSetup() {
  test.beforeEach(async ({ page }) => {
    test.setTimeout(45000);
    await createAdBidsModel(page);
    await createExploreFromModel(page);
  });
}
