import type { V1ListResourcesResponse } from "@rilldata/web-common/runtime-client";
import { spawnAndMatch } from "@rilldata/web-common/tests/utils/spawn";
import { rmSync } from "fs";
import { join } from "node:path";
import treeKill from "tree-kill";
import { getOpenPort } from "web-local/tests/utils/getOpenPort";
import {
  BASE_PROJECT_DIRECTORY,
  test as setup,
} from "web-local/tests/setup/base";
import { cpSync, readdirSync } from "node:fs";
import { expect } from "@playwright/test";

const TEST_PROJECTS = "tests/data/projects";

setup("should prep projects", async () => {
  await Promise.all(readdirSync(TEST_PROJECTS).map(prepProject));
});

/**
 * Prep a project by pre-ingesting and building the tmp folder.
 */
async function prepProject(name: string) {
  const TEST_PROJECT_SRC_DIRECTORY = join(TEST_PROJECTS, name);
  const TEST_PROJECT_DIRECTORY = join(BASE_PROJECT_DIRECTORY, name);
  const TEST_PORT = await getOpenPort();
  const TEST_PORT_GRPC = await getOpenPort();

  rmSync(TEST_PROJECT_DIRECTORY, { force: true, recursive: true });
  cpSync(TEST_PROJECT_SRC_DIRECTORY, TEST_PROJECT_DIRECTORY, {
    recursive: true,
    force: true,
  });

  const { process } = await spawnAndMatch(
    "../rill",
    [
      "start",
      "--no-open",
      "--port",
      "" + TEST_PORT,
      "--port-grpc",
      "" + TEST_PORT_GRPC,
      TEST_PROJECT_DIRECTORY,
    ],
    new RegExp("Serving Rill on: http://localhost"),
  );

  // Wait for all resources to complete reconcile.
  // TODO: Verify this in the UI instead of hitting the ListResources API
  await expect
    .poll(
      async () => {
        const resp = await fetch(
          "http://localhost:" + TEST_PORT + "/v1/instances/default/resources",
        );
        if (!resp.ok) return false;
        const json = (await resp.json()) as V1ListResourcesResponse;
        const relevantResources = json.resources?.filter(
          (r) => r.meta?.name?.kind !== "rill.runtime.v1.ProjectParser",
        );
        return (
          relevantResources?.every(
            (r) => r.meta?.reconcileStatus === "RECONCILE_STATUS_IDLE",
          ) ?? false
        );
      },
      {
        timeout: 5 * 60 * 1000,
      },
    )
    .toBeTruthy();

  const processExit = new Promise((resolve) => {
    process.on("exit", resolve);
  });
  if (process.pid) treeKill(process.pid);
  await processExit;
}
