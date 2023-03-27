import { afterAll, afterEach, beforeAll, beforeEach } from "@jest/globals";
import { isPortOpen } from "@rilldata/web-local/lib/util/isPortOpen";
import { asyncWaitUntil } from "@rilldata/web-local/lib/util/waitUtils";
import { rmSync } from "fs";
import type { ChildProcess } from "node:child_process";
import { spawn } from "node:child_process";
import path from "node:path";
import { Browser, chromium, Page } from "playwright";
import treeKill from "tree-kill";

export function useTestServer(port: number, dir: string) {
  let childProcess: ChildProcess;

  beforeEach(async () => {
    rmSync(dir, {
      force: true,
      recursive: true,
    });

    childProcess = spawn(
      path.join(__dirname, "../../../..", "rill"),
      [
        "start",
        "--no-open",
        "--port",
        port + "",
        "--port-grpc",
        port + 1000 + "",
        "--project",
        dir,
      ],
      {
        stdio: "inherit",
        shell: true,
      }
    );
    childProcess.on("error", console.log);
    await asyncWaitUntil(() => isPortOpen(port));
  });

  afterEach(async () => {
    if (childProcess.pid) treeKill(childProcess.pid);
    await asyncWaitUntil(async () => !(await isPortOpen(port)));
  });
}

export function useTestBrowser(port: number) {
  let browser: Browser;
  const testBrowser: {
    page: Page;
  } = {
    page: undefined,
  };

  beforeAll(async () => {
    browser = await chromium.launch({
      headless: false,
      // devtools: true,
    });
  });

  beforeEach(async () => {
    testBrowser.page = await browser.newPage();
    await testBrowser.page.goto(`http://localhost:${port}`);
  });

  afterEach(() => {
    return testBrowser.page.close();
  });

  afterAll(() => {
    return browser?.close();
  });

  return testBrowser;
}
