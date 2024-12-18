import { expect } from "@playwright/test";
import { exec } from "child_process";
import { promisify } from "util";
import { test } from "./setup/base";

const execAsync = promisify(exec);

test.describe("Organizations", () => {
  test("should create an organization", async ({ cli: _, page }) => {
    // Create an organization
    const { stdout: orgCreateStdout } = await execAsync("rill org create e2e");
    expect(orgCreateStdout).toContain("Created organization");

    // Go to the organization's page
    await page.goto("/e2e");
    await expect(page.getByRole("heading", { name: "e2e" })).toBeVisible();

    // Clean up quickly
    await execAsync("rill org delete e2e --force");
  });
});
