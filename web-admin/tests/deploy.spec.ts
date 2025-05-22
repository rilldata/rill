import { test, TestTempDirectory } from "./setup/base";
import { join } from "node:path";

test.describe("Deploy journey", () => {
  test.use({
    cliHome: join(TestTempDirectory, "deploy_home"),
    rillDevProject: "adbids_lite",
  });

  test("Should create new org and deploy", async ({ rillDevPage }) => {
    await rillDevPage.getByRole("button", { name: "Deploy" }).click();
  });
});
