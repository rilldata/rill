import type { Page } from "@playwright/test";
import { asyncWaitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import {
  V1ReconcileStatus,
  type V1Resource,
} from "@rilldata/web-common/runtime-client/gen/index.schemas";
import axios from "axios";
import { spawn, type ChildProcess } from "node:child_process";
import {
  cpSync,
  existsSync,
  mkdirSync,
  readdirSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { join } from "node:path";
import { test as base, expect } from "playwright/test";
import treeKill from "tree-kill";
import { getOpenPort } from "@rilldata/web-common/tests/utils/get-open-port.ts";
import { makeTempDir } from "@rilldata/web-common/tests/utils/make-temp-dir.ts";
import { spawnAndMatch } from "@rilldata/web-common/tests/utils/spawn.ts";

// Matches cli/pkg/local.DefaultDBDir: the directory inside the project where the
// running runtime keeps its DuckDB and catalog. It must be preserved when
// resetting a reused runtime's project, or the open database is corrupted.
const DB_DIR = "tmp";
const RILL_YAML = "rill.yaml";
const ProjectParserKind = "rill.runtime.v1.ProjectParser";

// The long-lived runtime shared by every test on a worker.
type SharedRuntime = {
  port: number;
  projectDir: string;
  homeDir: string;
};

type MyFixtures = {
  cliHomeDir: string | undefined;
  project: string | undefined;
  projectDir: string | undefined;
  rillDevBrowserState: string | undefined;
  // Opt out of the shared per-worker runtime and get a pristine instance for
  // this test. Use for tests that mutate global runtime state (env changes,
  // controller restarts) and can't tolerate a reused process.
  freshInstance: boolean;
  rillDevPage: Page;
};

type MyWorkerFixtures = {
  sharedRuntime: SharedRuntime;
};

export const rillDev = base.extend<MyFixtures, MyWorkerFixtures>({
  // When set, the test gets its own pristine runtime using this home (rather
  // than the shared per-worker runtime). Tests that depend on CLI login/auth
  // state living in a specific home (e.g. the deploy journey) set this. When
  // unset, the shared runtime's isolated worker home is used.
  cliHomeDir: [undefined, { option: true }],
  project: [undefined, { option: true }],
  // We default to using a randomly created temporary directory for project.
  // This can be used to get a consistent
  projectDir: [undefined, { option: true }],
  // If set, used to create the context used to create the rillDevPage.
  // A fresh context is used if not provided.
  rillDevBrowserState: [undefined, { option: true }],
  freshInstance: [false, { option: true }],

  // Start the rill binary once per worker and reuse it across every test the
  // worker runs. Between tests the project is reset (see resetSharedProject)
  // rather than killing the process, which avoids ~one binary spawn per test
  // and the DuckDB teardown races that came with it.
  sharedRuntime: [
    // eslint-disable-next-line no-empty-pattern -- Playwright requires the fixtures arg even when unused
    async ({}, use) => {
      const port = await getOpenPort();
      const grpcPort = await getOpenPort();
      const homeDir = makeTempDir("home");
      const projectDir = makeTempDir("project");

      const childProcess = await startRuntime({
        port,
        grpcPort,
        projectDir,
        homeDir,
      });

      try {
        await use({ port, projectDir, homeDir });
      } finally {
        await stopRuntime(childProcess, projectDir);
      }
    },
    { scope: "worker" },
  ],

  rillDevPage: async (
    {
      browser,
      sharedRuntime,
      project,
      projectDir,
      cliHomeDir,
      rillDevBrowserState,
      freshInstance,
      timezoneId,
      locale,
    },
    use,
  ) => {
    // A test needs its own instance when it explicitly opts out, pins a
    // specific project directory the shared runtime doesn't watch, or depends on
    // a specific CLI home (the shared runtime uses its own isolated worker home).
    const needsOwnInstance =
      freshInstance || projectDir !== undefined || cliHomeDir !== undefined;

    let port: number;
    let ownProcess: ChildProcess | undefined;
    let ownProjectDir: string | undefined;

    if (needsOwnInstance) {
      port = await getOpenPort();
      const grpcPort = await getOpenPort();
      ownProjectDir = projectDir ?? makeTempDir(`projects-${port}`);

      rmSync(ownProjectDir, { force: true, recursive: true });
      mkdirSync(ownProjectDir, { recursive: true });
      if (project) {
        cpSync(projectFixtureDir(project), ownProjectDir, {
          recursive: true,
          force: true,
        });
      }

      ownProcess = await startRuntime({
        port,
        grpcPort,
        projectDir: ownProjectDir,
        homeDir: cliHomeDir ?? makeTempDir("home"),
      });
    } else {
      port = sharedRuntime.port;
      await resetSharedProject(sharedRuntime, project);
    }

    // Wait for the project to fully reconcile before any test interaction, so
    // tests never navigate to an explore/canvas that doesn't exist yet.
    await waitForProjectReady(port, projectHasResources(project));

    const context = await browser.newContext({
      storageState: rillDevBrowserState ?? { cookies: [], origins: [] },
      ...(timezoneId ? { timezoneId } : {}),
      ...(locale ? { locale } : {}),
    });
    const page = await context.newPage();

    await page.goto(`http://localhost:${port}`);

    await use(page);

    // Close browser context to release any connections/resources first.
    await context.close();

    // Only per-test instances are torn down here; the shared runtime lives for
    // the lifetime of the worker.
    if (ownProcess) await stopRuntime(ownProcess, ownProjectDir!);
  },
});

function projectFixtureDir(project: string) {
  return join(import.meta.dirname, "../projects", project);
}

/** Whether a project fixture defines resources (anything beyond rill.yaml). */
function projectHasResources(project: string | undefined): boolean {
  if (!project) return false;
  const stack = [projectFixtureDir(project)];
  while (stack.length > 0) {
    const dir = stack.pop()!;
    for (const entry of readdirSync(dir, { withFileTypes: true })) {
      if (entry.isDirectory()) {
        stack.push(join(dir, entry.name));
        continue;
      }
      if (entry.name === "rill.yaml") continue;
      if (/\.(sql|yaml|yml)$/.test(entry.name)) return true;
    }
  }
  return false;
}

type StartRuntimeOptions = {
  port: number;
  grpcPort: number;
  projectDir: string;
  homeDir: string;
};

async function startRuntime({
  port,
  grpcPort,
  projectDir,
  homeDir,
}: StartRuntimeOptions): Promise<ChildProcess> {
  // Switch env to "dev" so that this points to the locally started rill cloud.
  // For tests that involve a local cloud this will point to it. Otherwise, when
  // running on a dev's machine, it avoids pointing to prod cloud.
  await spawnAndMatch(
    "../rill",
    "devtool switch-env dev".split(" "),
    /Set default env to "dev"/,
    {
      // Override home so that the instance is isolated for the provided home.
      additionalEnv: { HOME: homeDir },
    },
  );

  rmSync(projectDir, { force: true, recursive: true });
  mkdirSync(projectDir, { recursive: true });

  const cmd = `start --no-open --port ${port} --port-grpc ${grpcPort} ${projectDir}`;
  const childProcess = spawn("../rill", cmd.split(" "), {
    stdio: "inherit",
    shell: true,
    env: {
      ...process.env,
      // Override home so that the instance is isolated for the provided home.
      HOME: homeDir,
    },
  });
  childProcess.on("error", console.log);

  // Ping runtime until it's ready.
  await asyncWaitUntil(async () => {
    try {
      const response = await axios.get(`http://localhost:${port}/v1/ping`);
      return response.status === 200;
    } catch {
      return false;
    }
  });

  return childProcess;
}

async function stopRuntime(childProcess: ChildProcess, projectDir: string) {
  const processExit = new Promise((resolve) => {
    childProcess.on("exit", resolve);
  });

  if (childProcess.pid) treeKill(childProcess.pid);

  await processExit;

  // Remove the project directory after the dev process has fully exited.
  // Use expect.poll with exponential intervals to handle transient FS errors.
  await expect
    .poll(
      () => {
        try {
          rmSync(projectDir, { force: true, recursive: true });
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
}

/**
 * Reset a reused runtime's project to the given fixture. Clears the previous
 * project's source files (preserving the runtime's open DB dir), waits for the
 * runtime to tear the old resources down, then copies in the new fixture. The
 * explicit empty barrier avoids racing the file watcher: without it, readiness
 * could observe the previous project still idle and return stale.
 */
async function resetSharedProject(
  runtime: SharedRuntime,
  project: string | undefined,
) {
  // Keep a rill.yaml present at all times. Deleting it puts the parser into an
  // error state where it stops removing the previous project's resources, so
  // the clear barrier below would never observe them disappear.
  const rillYamlPath = join(runtime.projectDir, RILL_YAML);
  if (!existsSync(rillYamlPath)) {
    writeFileSync(rillYamlPath, "compiler: rillv1\n");
  }

  // Remove the previous project's source files, preserving rill.yaml and the
  // running runtime's DB dir, then wait for its resources to be torn down.
  for (const entry of readdirSync(runtime.projectDir)) {
    if (entry === DB_DIR || entry === RILL_YAML) continue;
    rmSync(join(runtime.projectDir, entry), { force: true, recursive: true });
  }

  await waitForNoDataResources(runtime.port);

  // Copy the new project in; its own rill.yaml overwrites the placeholder.
  if (project) {
    cpSync(projectFixtureDir(project), runtime.projectDir, {
      recursive: true,
      force: true,
    });
  }
}

async function fetchResources(port: number): Promise<V1Resource[]> {
  const response = await axios.get(
    `http://localhost:${port}/v1/instances/default/resources`,
  );
  return (response.data?.resources ?? []) as V1Resource[];
}

function dataResourcesOf(resources: V1Resource[]) {
  return resources.filter((r) => r.meta?.name?.kind !== ProjectParserKind);
}

function isIdle(resource: V1Resource) {
  return (
    resource.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_IDLE
  );
}

/**
 * Wait until the previous project's data resources have been torn down. The
 * ProjectParser is intentionally ignored: it stays RUNNING while watching the
 * repo and never settles to idle.
 */
async function waitForNoDataResources(port: number, timeoutMs = 30_000) {
  const cleared = await asyncWaitUntil(async () => {
    try {
      const resources = await fetchResources(port);
      return dataResourcesOf(resources).length === 0;
    } catch {
      return false;
    }
  }, timeoutMs);

  if (!cleared) {
    throw new Error(
      `Project did not clear on port ${port} within ${timeoutMs}ms`,
    );
  }
}

/**
 * Wait until the project has fully reconciled: every data resource is idle and
 * the resource set is stable across consecutive polls (so we don't return
 * mid-reconcile while resources are still being created). When the project
 * defines resources, requires at least one to be present. The ProjectParser is
 * ignored: it stays RUNNING while watching the repo and never settles to idle.
 */
async function waitForProjectReady(
  port: number,
  expectResources: boolean,
  timeoutMs = 60_000,
) {
  let prevSignature = "";
  let stablePolls = 0;
  let reachedIdle = false;
  let lastResources: V1Resource[] = [];

  const ready = await asyncWaitUntil(async () => {
    let resources: V1Resource[];
    try {
      resources = await fetchResources(port);
    } catch {
      stablePolls = 0;
      return false;
    }
    lastResources = resources;

    const data = dataResourcesOf(resources);
    if (expectResources && data.length === 0) {
      stablePolls = 0;
      return false;
    }
    if (!data.every(isIdle)) {
      stablePolls = 0;
      return false;
    }

    // The project has fully reconciled at least once. Used as a fallback below
    // if the resource set never stabilizes (e.g. background cloud sync in the
    // deploy tests keeps churning it).
    reachedIdle = true;

    // Track the set of resource names only, not their stateVersion: a resource
    // that re-reconciles and returns to idle (common while a project is being
    // deployed) shouldn't reset stability and stall readiness. We only need the
    // resource set to stop growing/shrinking while everything is idle.
    const signature = data
      .map((r) => `${r.meta?.name?.kind}/${r.meta?.name?.name}`)
      .sort()
      .join("|");
    if (signature === prevSignature) {
      stablePolls += 1;
    } else {
      prevSignature = signature;
      stablePolls = 1;
    }
    // Several consecutive identical polls (~250ms apart) means reconciliation
    // has settled rather than still creating resources.
    return stablePolls >= 3;
  }, timeoutMs);

  if (ready) {
    const errors = dataResourcesOf(lastResources).filter(
      (r) => r.meta?.reconcileError,
    );
    if (errors.length > 0) {
      const details = errors
        .map(
          (r) =>
            `${r.meta?.name?.kind}/${r.meta?.name?.name}: ${r.meta?.reconcileError}`,
        )
        .join("\n");
      throw new Error(`Reconciliation errors:\n${details}`);
    }
    return;
  }

  // The set never stabilized within the timeout. If the project did fully
  // reconcile at some point, proceed: the churn is background activity (e.g.
  // cloud sync), not an unready project. Only fail if it never reached idle.
  if (reachedIdle) return;

  const pending = dataResourcesOf(lastResources)
    .filter((r) => !isIdle(r))
    .map(
      (r) =>
        `${r.meta?.name?.kind}/${r.meta?.name?.name}: ${r.meta?.reconcileStatus}`,
    )
    .join("\n");
  throw new Error(
    `Project did not become ready on port ${port}. Still pending:\n${pending || "(none)"}`,
  );
}
