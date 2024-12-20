import { expect } from "@playwright/test";
import { exec } from "child_process";
import { promisify } from "util";

const execAsync = promisify(exec);

export async function orgCreate() {
  const { stdout: orgCreateStdout } = await execAsync("rill org create e2e");
  expect(orgCreateStdout).toContain("Created organization");
}

export async function orgDelete() {
  const { stdout: orgDeleteStdout } = await execAsync(
    "rill org delete e2e --force",
  );
  expect(orgDeleteStdout).toContain("Deleted organization");
}
