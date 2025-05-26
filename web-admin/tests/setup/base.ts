import { test as base, type Page } from "@playwright/test";
import { getOpenPort } from "@rilldata/web-common/tests/utils/get-open-port";
import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils";
import { spawnAndMatch } from "@rilldata/web-common/tests/utils/spawn";
import axios from "axios";
import { spawn } from "node:child_process";
import { cpSync, existsSync, mkdirSync, rmSync } from "node:fs";
import { join } from "node:path";
import treeKill from "tree-kill";
import { ADMIN_STORAGE_STATE, VIEWER_STORAGE_STATE } from "./constants";
import { cliLogin, cliLogout } from "./fixtures/cli";
import path from "path";
import { fileURLToPath } from "url";
import {
  RILL_EMBED_SERVICE_TOKEN_FILE,
  RILL_ORG_NAME,
  RILL_PROJECT_NAME,
} from "./constants";
import fs from "fs";
import { generateEmbed } from "../utils/generate-embed";

export const TestTempDirectory = "playwright";
const TEST_PROJECTS = "tests/setup/projects";

type MyFixtures = {
  adminPage: Page;
  viewerPage: Page;
  anonPage: Page;
  embedPage: Page;

  cli: void;
  cliHome: string | undefined;
  rillDevPage: Page;
  rillDevProject: string | undefined;
};

export const test = base.extend<MyFixtures>({
  cliHome: [undefined, { option: true }],
  rillDevProject: [undefined, { option: true }],

  // Note: the `e2e` project uses the admin auth file by default, so it's likely that
  // this fixture won't be used often.
  adminPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: ADMIN_STORAGE_STATE,
    });
    const adminPage = await context.newPage();
    await use(adminPage);
    await context.close();
  },

  viewerPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: VIEWER_STORAGE_STATE,
    });
    const viewerPage = await context.newPage();
    await use(viewerPage);
    await context.close();
  },

  anonPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: { cookies: [], origins: [] },
    });
    const anonPage = await context.newPage();
    await use(anonPage);
    await context.close();
  },

  cli: async ({ page }, use) => {
    await cliLogin(page);
    await use();
    await cliLogout();
  },

  rillDevPage: async ({ anonPage, cliHome, rillDevProject }, use) => {
    const TEST_PORT = await getOpenPort();
    const TEST_PORT_GRPC = await getOpenPort();
    const TEST_PROJECT_DIRECTORY = join(
      TestTempDirectory,
      "temp",
      "" + TEST_PORT,
    );
    // Add a default home if cliHome is not provided so that tests always have a different home than the user's home.
    // This will make sure that login status won't conflicts with dev's login status when run locally.
    const TEST_HOME_DIRECTORY = cliHome ?? join(TestTempDirectory, "home");

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

    rmSync(TEST_PROJECT_DIRECTORY, { force: true, recursive: true });

    if (!existsSync(TEST_PROJECT_DIRECTORY)) {
      mkdirSync(TEST_PROJECT_DIRECTORY, { recursive: true });
    }

    if (rillDevProject) {
      cpSync(join(TEST_PROJECTS, rillDevProject), TEST_PROJECT_DIRECTORY, {
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
        const response = await axios.get(
          `http://localhost:${TEST_PORT}/v1/ping`,
        );
        return response.status === 200;
      } catch {
        return false;
      }
    });

    await anonPage.goto(`http://localhost:${TEST_PORT}`);

    // Seems to help with issues related to DOM elements not being ready
    await anonPage.waitForTimeout(1500);

    await use(anonPage);

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

  embedPage: [
    async ({ browser }, use) => {
      const __dirname = path.dirname(fileURLToPath(import.meta.url));
      const readPath = path.join(
        __dirname,
        "..",
        "..",
        RILL_EMBED_SERVICE_TOKEN_FILE,
      );
      const rillServiceToken = fs.readFileSync(readPath, "utf-8");

      await generateEmbed(
        RILL_ORG_NAME,
        RILL_PROJECT_NAME,
        "bids_explore",
        rillServiceToken,
      );
      const filePath = "file://" + path.resolve(__dirname, "..", "embed.html");

      const context = await browser.newContext();
      const embedPage = await context.newPage();
      await embedPage.goto(filePath);

      await use(embedPage);

      await context.close();
    },
    { scope: "test" },
  ],
});
