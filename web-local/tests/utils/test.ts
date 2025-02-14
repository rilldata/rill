import { test as base } from "@playwright/test";
import { rmSync, writeFileSync, existsSync, mkdirSync } from "fs";
import { spawn } from "node:child_process";
import treeKill from "tree-kill";
import { getOpenPort } from "./getOpenPort";
import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils";
import axios from "axios";

type MyFixtures = {
  includeRillYaml: boolean;
};

const BASE_PROJECT_DIRECTORY = "temp/test-project";

export const test = base.extend<MyFixtures>({
  includeRillYaml: [true, { option: true }],

  page: async ({ page, includeRillYaml }, use) => {
    const TEST_PORT = await getOpenPort();
    const TEST_PORT_GRPC = await getOpenPort();
    const TEST_PROJECT_DIRECTORY = `${BASE_PROJECT_DIRECTORY}-${TEST_PORT}`;

    rmSync(TEST_PROJECT_DIRECTORY, { force: true, recursive: true });

    if (!existsSync(TEST_PROJECT_DIRECTORY)) {
      mkdirSync(TEST_PROJECT_DIRECTORY, { recursive: true });
    }

    // Conditionally write the file
    if (includeRillYaml) {
      writeFileSync(
        `${TEST_PROJECT_DIRECTORY}/rill.yaml`,
        'compiler: rill-beta\ntitle: "Test Project"',
      );
    }

    const cmd = `start --no-open --port ${TEST_PORT} --port-grpc ${TEST_PORT_GRPC} ${TEST_PROJECT_DIRECTORY}`;

    const childProcess = spawn("../rill", cmd.split(" "), {
      stdio: "inherit",
      shell: true,
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

    await page.goto(`http://localhost:${TEST_PORT}`);

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
