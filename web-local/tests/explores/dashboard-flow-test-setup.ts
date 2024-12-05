import { createExploreFromModel } from "web-local/tests/utils/exploreHelpers";
import { createAdBidsModel } from "web-local/tests/utils/dataSpecifcHelpers";
import { test } from "../utils/test";

export function useDashboardFlowTestSetup(navigateToPreview = false) {
  test.beforeEach(async ({ page }) => {
    await createAdBidsModel(page);
    await createExploreFromModel(page, navigateToPreview);
  });
}
