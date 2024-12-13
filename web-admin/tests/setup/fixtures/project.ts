import type { Page } from "@playwright/test";
import { exec } from "child_process";
import { promisify } from "util";

const execAsync = promisify(exec);

export async function projectDeploy(page: Page) {
  await execAsync(
    "rill deploy --path tests/setup/git/repos/rill-examples --subpath rill-openrtb-prog-ads --project openrtb --github true",
  );
  await page.goto("/e2e/openrtb");
}

export async function projectDelete() {
  await execAsync("rill project delete openrtb --force");
  await execAsync(
    "rm -rf tests/setup/git/repos/rill-examples/rill-openrtb-prog-ads/.rillcloud",
  );
}
