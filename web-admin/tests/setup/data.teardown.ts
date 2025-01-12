import { execAsync } from "../utils/spawn";
import { test as teardown } from "./base";

teardown("should remove the test data", async ({ cli: _ }) => {
  // Delete the organization and its projects
  await execAsync("rill org delete e2e --force");

  // Delete the user
  await execAsync(
    "PGPASSWORD=postgres psql -h localhost -p 6432 -U postgres -d postgres -c 'DELETE FROM users;'",
  );

  // Delete the project's Rill Cloud metadata
  await execAsync(
    "rm -rf tests/setup/git/repos/rill-examples/rill-openrtb-prog-ads/.rillcloud",
  );
});
