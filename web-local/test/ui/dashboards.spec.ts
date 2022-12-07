import { describe, it } from "@jest/globals";
import { createDashboardFromSource } from "./utils/dashboardHelpers";
import { TestEntityType } from "./utils/helpers";
import { useRegisteredServer } from "./utils/serverConfigs";
import { createOrReplaceSource } from "./utils/sourceHelpers";
import { waitForEntity } from "./utils/waitHelpers";

describe.skip("dashboards", () => {
  const testBrowser = useRegisteredServer("dashboards");

  it("autogenerate dashboard", async () => {
    await createOrReplaceSource(testBrowser.page, "AdBids.csv", "AdBids");
    await createDashboardFromSource(testBrowser.page, "AdBids");
    await waitForEntity(
      testBrowser.page,
      TestEntityType.Dashboard,
      "AdBids_dashboard",
      true
    );
  });
});
