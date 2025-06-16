import type { Page } from "@playwright/test";
import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils";
import { ADMIN_STORAGE_STATE } from "@rilldata/web-integration/tests/constants";
import axios from "axios";
import { spawn } from "node:child_process";
import { cpSync, existsSync, mkdirSync, rmSync } from "node:fs";
import { join } from "node:path";
import { test as base } from "playwright/test";
import treeKill from "tree-kill";
import { getOpenPort } from "../utils/get-open-port";
import { makeTempDir } from "../utils/make-temp-dir";
import { spawnAndMatch } from "../utils/spawn";

type MyFixtures = {
  cliHomeDir: string;
  project: string | undefined;
  rillDevPage: Page;
};

export const rillDev = base.extend<MyFixtures>({
  // Add a default home if cliHome is not provided so that tests always have a different home than the user's home.
  // This will make sure that login status won't conflicts with dev's login status when run locally.
  cliHomeDir: [makeTempDir("home"), { option: true }],
  project: [undefined, { option: true }],

  rillDevPage: async ({ browser, project, cliHomeDir }, use) => {
    const TEST_PORT = await getOpenPort();
    const TEST_GRPC_PORT = await getOpenPort();
    const TEST_PROJECT_DIRECTORY = makeTempDir(`projects-${TEST_PORT}`);

    // Switch env to "dev" so that this points to the locally started rill cloud.
    // For tests that involve a local cloud this will point to it.
    // Otherwise, when running in a dev's machine, it will avoid pointing to prod cloud and bombard prod.
    await spawnAndMatch(
      "../rill",
      "devtool switch-env dev".split(" "),
      /Set default env to "dev"/,
      {
        additionalEnv: {
          // Override home so that the instance is isolated for the provided cliHome.
          HOME: cliHomeDir,
        },
      },
    );

    rmSync(TEST_PROJECT_DIRECTORY, { force: true, recursive: true });

    if (!existsSync(TEST_PROJECT_DIRECTORY)) {
      mkdirSync(TEST_PROJECT_DIRECTORY, { recursive: true });
    }

    if (project) {
      const projectDir = join(import.meta.dirname, "../data/projects", project);
      cpSync(projectDir, TEST_PROJECT_DIRECTORY, {
        recursive: true,
        force: true,
      });
    }

    const cmd = `start --no-open --port ${TEST_PORT} --port-grpc ${TEST_GRPC_PORT} ${TEST_PROJECT_DIRECTORY}`;

    const childProcess = spawn("../rill", cmd.split(" "), {
      stdio: "inherit",
      shell: true,
      env: {
        ...process.env,
        // Override home so that the instance is isolated for the provided cliHome.
        // Login status will be siloed for tests using the same cliHome.
        HOME: cliHomeDir,
      },
    });

    childProcess.on("error", console.log);

    // Ping runtime until it's ready
    await asyncWaitUntil(async () => {
      try {
        const response = await axios.get(
          `http://localhost:${TEST_PORT}/v1/ping`,
        );
        return response.status === 200;
      } catch {
        return false;
      }
    });

    const context = await browser.newContext({
      storageState: { cookies: [], origins: [] },
    });
    const page = await context.newPage();

    await page.goto(`http://localhost:${TEST_PORT}`);

    // Seems to help with issues related to DOM elements not being ready
    await page.waitForTimeout(1500);

    await use(page);

    rmSync(TEST_PROJECT_DIRECTORY, {
      force: true,
      recursive: true,
    });

    const processExit = new Promise((resolve) => {
      childProcess.on("exit", resolve);
    });

    if (childProcess.pid) treeKill(childProcess.pid);

    await processExit;
  },
});
