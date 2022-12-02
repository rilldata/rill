import { isPortOpen } from "@rilldata/web-local/lib/util/isPortOpen";
import { asyncWaitUntil } from "@rilldata/web-local/lib/util/waitUtils";
import { rmSync } from "fs";
import { spawn } from "node:child_process";
import type { ChildProcess } from "node:child_process";
import path from "node:path";
import treeKill from "tree-kill";

export function useTestServer(port: number, dir: string) {
  let childProcess: ChildProcess;

  beforeAll(async () => {
    rmSync(dir, {
      force: true,
      recursive: true,
    });

    childProcess = spawn(
      "go",
      [
        "run",
        path.join(__dirname, "../../..", "cli/main.go"),
        "start",
        "--no-open",
        "--port",
        port + "",
        "--port-grpc",
        port + 1000 + "",
        "--dir",
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

  afterAll(() => {
    if (childProcess.pid) treeKill(childProcess.pid);
  });
}
