import { useDashboardFlowTestSetup } from "web-local/test/ui/dashboard-flows/dashboard-flow-test-setup";
import { test, expect } from "@playwright/test";
import { startRuntimeForEachTest } from "../utils/startRuntimeForEachTest";

test.describe("~~~~~~~~~~~~~~~~~~~~FIXME RENAME THIS~~~~~~~~~~~~~~~~~~~~~~~", () => {
  startRuntimeForEachTest();
  // dashboard test setup
  useDashboardFlowTestSetup();

  test("~~~~~~~~~~~~~~~~~~~~FIXME RENAME THIS~~~~~~~~~~~~~~~~~~~~~~~", async ({
    page,
  }) => {
    // Delete this when your flow is ready.
    await page.pause();

    await expect(page.getByText("example expect - will fail")).toBeVisible();
  });
});
