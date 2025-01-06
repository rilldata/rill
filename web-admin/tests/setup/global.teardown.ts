import { execAsync } from "../utils/spawn";

export default async function globalTeardown() {
  // Stop the cloud services
  await execAsync(
    "docker compose -f ../cli/cmd/devtool/data/cloud-deps.docker-compose.yml down --volumes",
  );

  // Remove the test repositories
  await execAsync("rm -rf tests/setup/git/repos");
}
