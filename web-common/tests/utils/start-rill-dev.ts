import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils";
import { getOpenPort } from "@rilldata/web-common/tests/utils/get-open-port";
import { spawnAndMatch } from "@rilldata/web-common/tests/utils/spawn";
import axios from "axios";
import { spawn } from "node:child_process";
import { cpSync, existsSync, mkdirSync, rmSync } from "node:fs";
import { join } from "node:path";
import type { Page } from "playwright";
import treeKill from "tree-kill";

export const TestTempDirectory = "playwright";

export async function startRillDev(
  page: Page,
  use: (page: Page) => Promise<void>,
  {
    cliHome,
    switchEnv,
    projectDir,
  }: {
    cliHome?: string;
    switchEnv?: boolean;
    projectDir?: string;
  },
) {
  const TEST_PORT = await getOpenPort();
  const TEST_PORT_GRPC = await getOpenPort();
  const TEST_PROJECT_DIRECTORY = join(
    TestTempDirectory,
    "projects",
    "" + TEST_PORT,
  );
  // Add a default home if cliHome is not provided so that tests always have a different home than the user's home.
  // This will make sure that login status won't conflicts with dev's login status when run locally.
  const TEST_HOME_DIRECTORY = cliHome ?? join(TestTempDirectory, "home");

  if (switchEnv) {
    // Switch env to "dev" so that this points to the locally started rill cloud
    await spawnAndMatch(
      "../rill",
      "devtool switch-env dev".split(" "),
      /Set default env to "dev"/,
      {
        additionalEnv: {
          // Override home so that the instance is isolated for the provided cliHome.
          HOME: TEST_HOME_DIRECTORY,
        },
      },
    );
  }

  rmSync(TEST_PROJECT_DIRECTORY, { force: true, recursive: true });

  if (!existsSync(TEST_PROJECT_DIRECTORY)) {
    mkdirSync(TEST_PROJECT_DIRECTORY, { recursive: true });
  }

  if (projectDir) {
    cpSync(projectDir, TEST_PROJECT_DIRECTORY, {
      recursive: true,
      force: true,
    });
  }

  const cmd = `start --no-open --port ${TEST_PORT} --port-grpc ${TEST_PORT_GRPC} ${TEST_PROJECT_DIRECTORY}`;

  const childProcess = spawn("../rill", cmd.split(" "), {
    stdio: "inherit",
    shell: true,
    env: {
      ...process.env,
      // Override home so that the instance is isolated for the provided cliHome.
      // Login status will be siloed for tests using the same cliHome.
      HOME: TEST_HOME_DIRECTORY,
    },
  });

  childProcess.on("error", console.log);

  // Ping runtime until it's ready
  await asyncWaitUntil(async () => {
    try {
      const response = await axios.get(`http://localhost:${TEST_PORT}/v1/ping`);
      return response.status === 200;
    } catch {
      return false;
    }
  });

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
}
