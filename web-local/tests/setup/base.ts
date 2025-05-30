import { test as base } from "@playwright/test";
import {
  startRillDev,
  TestTempDirectory,
} from "@rilldata/web-common/tests/utils/start-rill-dev";
import { join } from "node:path";

type MyFixtures = {
  project: string | undefined;
};

export const test = base.extend<MyFixtures>({
  project: [undefined, { option: true }],

  page: async ({ page, project }, use) => {
    await startRillDev(page, use, {
      ...(project
        ? {
            projectDir: join(TestTempDirectory, "projects", project),
          }
        : {}),
    });
  },
});
