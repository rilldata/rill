import { expect } from "@playwright/test";
import { exec } from "child_process";
import { promisify } from "util";
import { test } from "./setup/base";
import { spawnAndMatch } from "./utils/spawn";

const execAsync = promisify(exec);

test.describe("Projects", () => {
  test("should deploy a project", async ({ page, organization: _ }) => {
    // Execute the deploy command and capture the auth URL
    const { match } = await spawnAndMatch(
      "rill",
      [
        "deploy",
        "--path",
        "tests/setup/git/repos/rill-examples",
        "--subpath",
        "rill-openrtb-prog-ads",
        "--project",
        "openrtb",
        "--github",
        "true",
      ],
      /https?:\/\/[^\s]+/,
    );

    // Manually navigate to the auth URL
    const url = match[0];
    await page.goto(url);

    // Expect to see the successful deployment
    await page.goto("/e2e/openrtb");
    await expect(page.getByText("Your trial expires in 30 days")).toBeVisible(); // Billing banner
    await expect(page.getByText("e2e")).toBeVisible(); // Organization breadcrumb
    await expect(page.getByText("Free trial")).toBeVisible(); // Billing status
    await expect(page.getByText("openrtb")).toBeVisible(); // Project breadcrumb
    await expect(
      page.getByRole("link", { name: "Programmatic Ads Auction" }).first(),
    ).toBeVisible(); // Link to dashboard
    await expect(
      page.getByRole("link", { name: "Programmatic Ads Bids" }),
    ).toBeVisible(); // Link to dashboard

    // Clean up quickly
    await execAsync("rill project delete openrtb --force");
  });
});

test.afterAll(async () => {
  await execAsync(
    "rm -rf tests/setup/git/repos/rill-examples/rill-openrtb-prog-ads/.rillcloud",
  );
});
