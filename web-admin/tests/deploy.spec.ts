import { expect } from "@playwright/test";
import { exec } from "child_process";
import { promisify } from "util";
import { test } from "./setup/test";

const execAsync = promisify(exec);

test("deploy", async ({ page }) => {
  // Execute the deploy command
  const { stdout: deployStdout } = await execAsync(
    "rill deploy --path tests/setup/git/repos/rill-openrtb-prog-ads --github true --interactive false",
  );

  // Assert on CLI output
  expect(deployStdout).toContain(`Created project "e2e/rill-openrtb-prog-ads"`);
  expect(deployStdout).toContain(`Opening project in browser...`);

  // Expect to see the successful deployment
  await page.goto("/e2e/rill-openrtb-prog-ads");
  await expect(page.getByText("Your trial expires in 30 days")).toBeVisible(); // Billing banner
  await expect(page.getByText("e2e")).toBeVisible(); // Organization breadcrumb
  await expect(page.getByText("Free trial")).toBeVisible(); // Billing status
  await expect(page.getByText("rill-open-rtb-prog-ads")).toBeVisible(); // Project breadcrumb
  await expect(page.getByText("2 dashboards")).toBeVisible(); // Dashboard count
  await expect(page.getByText("Programmatic Ads Auction")).toBeVisible(); // Dashboard name
  await expect(page.getByText("Programmatic Ads Bids")).toBeVisible(); // Dashboard name
});

test.afterAll(async () => {
  await execAsync("rill project delete rill-openrtb-prog-ads --force");
  await execAsync(
    "rm -rf tests/setup/git/repos/rill-openrtb-prog-ads/.rillcloud",
  );
});
