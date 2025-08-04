import { type Page } from "@playwright/test";
import {
  execAsync,
  spawnAndMatch,
  type SpawnAndMatchResult,
} from "@rilldata/web-common/tests/utils/spawn.ts";

export async function cliLogin(
  page: Page,
  maybeLoginInPage: () => Promise<void> = () => Promise.resolve(),
  homeDir?: string,
) {
  // Run the login command and capture the verification URL
  const { process, match }: SpawnAndMatchResult = await spawnAndMatch(
    "rill",
    ["login", "--interactive=false"],
    /Open this URL in your browser to confirm the login: (.*)\n/,
    homeDir
      ? {
          additionalEnv: { HOME: homeDir },
        }
      : undefined,
  );

  const verificationUrl = match[1];

  // Manually navigate to the verification URL
  await page.goto(verificationUrl);

  await maybeLoginInPage();

  // Click the confirm button
  await page.getByRole("button", { name: /confirm/i }).click();

  // Wait for the process to complete
  await new Promise((resolve, reject) => {
    process.on("close", (code) => {
      if (code === 0) resolve(null);
      else reject(new Error(`Process exited with code ${code}`));
    });
  });
}

export async function cliLogout(homeDir?: string) {
  const homePrefix = homeDir ? `HOME=${homeDir} ` : "";
  await execAsync(`${homePrefix}rill logout`);
}
