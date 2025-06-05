import { execAsync } from "web-integration/tests/utils/spawn";

export async function isOrgDeleted(orgName: string): Promise<boolean> {
  try {
    // This command throws an exit code of 1 along with the "Org not found." message when the org is not found.
    await execAsync(`rill org show ${orgName}`);
    // If it doesn't throw, the org still exists.
    return false;
  } catch (error: any) {
    return error.stdout.includes("Org not found.");
  }
}
