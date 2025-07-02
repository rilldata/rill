import { execAsync } from "@rilldata/web-common/tests/utils/spawn.ts";

export async function isOrgDeleted(
  orgName: string,
  home?: string,
): Promise<boolean> {
  const envOverride = home ? `HOME=${home}` : "";

  try {
    // This command throws an exit code of 1 along with the "Org not found." message when the org is not found.
    await execAsync(`${envOverride} rill org show ${orgName}`);
    // If it doesn't throw, the org still exists.
    return false;
  } catch (error: any) {
    return error.stdout.includes("Org not found.");
  }
}
