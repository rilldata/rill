import { execAsync } from "../utils/spawn";
import { test as teardown } from "./base";

teardown("should remove the test data", async ({ cli: _cli }) => {
  // Delete the organization and its projects
  await execAsync("rill org delete e2e --force");

  // Delete the project's Rill Cloud metadata
  await execAsync(
    "rm -rf tests/setup/git/repos/rill-examples/rill-openrtb-prog-ads/.rillcloud",
  );
});
