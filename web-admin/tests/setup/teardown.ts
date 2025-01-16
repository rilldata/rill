import fs from "fs";
import { execAsync } from "../utils/spawn";
import { test as teardown } from "./base";

teardown("should stop all services", async () => {
  // Stop the admin and runtime services:
  // 1. Get the process ID from the file
  // 2. Get the process group ID
  // 3. Kill the whole process group
  const processID = fs.readFileSync("rill-devtool-pid.txt", "utf8");
  const { stdout: processGroupID } = await execAsync(
    `ps -o pgid= -p ${processID}`,
  );
  await execAsync(`kill -TERM -${processGroupID}`);

  // Stop the cloud services
  await execAsync(
    "docker compose -f ../cli/cmd/devtool/data/cloud-deps.docker-compose.yml down --volumes",
  );

  // Remove the test repositories
  await execAsync("rm -rf tests/setup/git/repos");
});
