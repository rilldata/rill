import { expect } from "@playwright/test";
import fs from "fs";
import { execAsync } from "../utils/spawn";
import { test as teardown } from "./base";
import { RILL_DEVTOOL_BACKGROUND_PROCESS_PID_FILE } from "./constants";

teardown.describe("global teardown", () => {
  teardown("should clean up the test organization", async ({ cli: _ }) => {
    await execAsync("rill org delete e2e --force");

    // Wait for the organization to be deleted
    // This includes deleting the org from Orb and Stripe, which we'd like to do to keep those environments clean.
    await expect
      .poll(async () => await isOrgDeleted("e2e"), {
        intervals: [1_000],
        timeout: 15_000,
      })
      .toBeTruthy();
  });

  teardown("should stop all services", async () => {
    // Stop the admin and runtime services:
    // 1. Get the process ID from the file
    // 2. Get the process group ID
    // 3. Kill the whole process group
    // 4. Delete the process ID file
    const processID = fs.readFileSync(
      RILL_DEVTOOL_BACKGROUND_PROCESS_PID_FILE,
      "utf8",
    );
    const { stdout: processGroupID } = await execAsync(
      `ps -o pgid= -p ${processID}`,
    );
    await execAsync(`kill -TERM -${processGroupID}`);
    await execAsync(`rm ${RILL_DEVTOOL_BACKGROUND_PROCESS_PID_FILE}`);

    // Stop the cloud services
    await execAsync(
      "docker compose -f ../cli/cmd/devtool/data/cloud-deps.docker-compose.yml down --volumes",
    );

    // Remove the test repositories
    await execAsync("rm -rf tests/setup/git/repos");
  });
});

async function isOrgDeleted(orgName: string): Promise<boolean> {
  try {
    // This command throws an exit code of 1 along with the "Org not found." message when the org is not found.
    await execAsync(`rill org show ${orgName}`);
    // If it doesn't throw, the org still exists.
    return false;
  } catch (error: any) {
    // Check for the expected error code.
    return error.code === 1;
  }
}
