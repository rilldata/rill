import { test as base } from "@playwright/test";
import { rmSync, existsSync, mkdirSync, cpSync } from "node:fs";
import { spawn } from "node:child_process";
import { join } from "node:path";
import treeKill from "tree-kill";
import { getOpenPort } from "web-local/tests/utils/getOpenPort";
import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils";
import axios from "axios";

export const BASE_PROJECT_DIRECTORY = "temp/test-project";

type ProjectInitArgs = { name?: string } | undefined;
type MyFixtures = {
  project: ProjectInitArgs;
};

export const test = base.extend<MyFixtures>({
  project: [undefined, { option: true }],

  page: async ({ page, project }, use) => {
    const TEST_PORT = await getOpenPort();
    const TEST_PORT_GRPC = await getOpenPort();
    const TEST_PROJECT_DIRECTORY = join(BASE_PROJECT_DIRECTORY, "" + TEST_PORT);

    rmSync(TEST_PROJECT_DIRECTORY, { force: true, recursive: true });

    if (!existsSync(TEST_PROJECT_DIRECTORY)) {
      mkdirSync(TEST_PROJECT_DIRECTORY, { recursive: true });
    }

    if (project?.name) {
      cpSync(
        join(BASE_PROJECT_DIRECTORY, project.name),
        TEST_PROJECT_DIRECTORY,
        {
          recursive: true,
          force: true,
        },
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
