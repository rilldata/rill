import { exec } from "child_process";
import { promisify } from "util";

const execAsync = promisify(exec);

const skipGlobalSetup = process.env.E2E_SKIP_GLOBAL_SETUP === "true";

export default async function globalTeardown() {
  if (skipGlobalSetup) return;

  // Stop the cloud services
  await execAsync(
    "docker compose -f ../cli/cmd/devtool/data/cloud-deps.docker-compose.yml down --volumes",
  );

  // Remove the test repositories
  await execAsync("rm -rf tests/setup/git/repos");
}
