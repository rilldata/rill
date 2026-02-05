import type { Page } from "@playwright/test";
import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import axios from "axios";
import { spawn } from "node:child_process";
import { cpSync, existsSync, mkdirSync, rmSync } from "node:fs";
import { join } from "node:path";
import { test as base, expect } from "playwright/test";
import treeKill from "tree-kill";
import { getOpenPort } from "@rilldata/web-common/tests/utils/get-open-port.ts";
import { makeTempDir } from "@rilldata/web-common/tests/utils/make-temp-dir.ts";
import { spawnAndMatch } from "@rilldata/web-common/tests/utils/spawn.ts";

type MyFixtures = {
  cliHomeDir: string;
  project: string | undefined;
  projectDir: string | undefined;
  rillDevPage: Page;
  rillDevBrowserState: string | undefined;
};

export const rillDev = base.extend<MyFixtures>({
  // Add a default home if cliHome is not provided so that tests always have a different home than the user's home.
  // This will make sure that login status won't conflicts with dev's login status when run locally.
  cliHomeDir: [makeTempDir("home"), { option: true }],
  project: [undefined, { option: true }],
  // We default to using a randomly created temporary directory for project.
  // This can be used to get a consistent
  projectDir: [undefined, { option: true }],
  // If set, used to create the context used to create the rillDevPage.
  // A fresh context is used if not provided.
  rillDevBrowserState: [undefined, { option: true }],

  rillDevPage: async (
    { browser, project, projectDir, cliHomeDir, rillDevBrowserState },
    use,
  ) => {
    const TEST_PORT = await getOpenPort();
    const TEST_GRPC_PORT = await getOpenPort();
    const TEST_PROJECT_DIRECTORY =
      projectDir ?? makeTempDir(`projects-${TEST_PORT}`);

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
      const sourceProjectDir = join(
        import.meta.dirname,
        "../projects",
        project,
      );
      cpSync(sourceProjectDir, TEST_PROJECT_DIRECTORY, {
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

    const context = await browser.newContext(
      rillDevBrowserState
        ? {
            storageState: rillDevBrowserState,
          }
        : {
            storageState: { cookies: [], origins: [] },
          },
    );
    const page = await context.newPage();

    await page.goto(`http://localhost:${TEST_PORT}`);

    // Seems to help with issues related to DOM elements not being ready
    await page.waitForTimeout(1500);

    await use(page);

    // Close browser context to release any connections/resources first
    await context.close();

    const processExit = new Promise((resolve) => {
      childProcess.on("exit", resolve);
    });

    if (childProcess.pid) treeKill(childProcess.pid);

    await processExit;

    // Remove the test project directory after the dev process has fully exited.
    // Use expect.poll with exponential intervals to handle transient FS errors.
    await expect
      .poll(
        () => {
          try {
            rmSync(TEST_PROJECT_DIRECTORY, { force: true, recursive: true });
            return true;
          } catch (err) {
            const code = (err as NodeJS.ErrnoException)?.code;
            const isTransient =
              code === "ENOTEMPTY" || code === "EBUSY" || code === "EPERM";
            if (isTransient) return false;
            throw err;
          }
        },
        {
          intervals: [200, 400, 800, 1600, 3200],
          timeout: 7000,
        },
      )
      .toBe(true);
  },
});
