import { execAsync } from "../utils/spawn";

const skipGlobalSetup = Boolean(process.env.E2E_SKIP_GLOBAL_SETUP);

export default async function globalTeardown() {
  if (skipGlobalSetup) return;

  // Stop the cloud services
  await execAsync(
    "docker compose -f ../cli/cmd/devtool/data/cloud-deps.docker-compose.yml down --volumes",
  );

  // Remove the test repositories
  await execAsync("rm -rf tests/setup/git/repos");
}
