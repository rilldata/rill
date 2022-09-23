import {
  TestSuiteParameter,
  TestSuiteSetup,
} from "@adityahegde/typescript-test-utils";
import terminate from "terminate/promise";
import { isPortOpen } from "$web-local/common/utils/isPortOpen";
import { CLI_COMMAND } from "./getCliCommand";
import { exec } from "node:child_process";
import type { ChildProcess } from "node:child_process";
import { promisify } from "util";
import { waitUntil } from "$web-local/common/utils/waitUtils";

const execPromise = promisify(exec);

export interface TestServerSetupParameter extends TestSuiteParameter {
  serverPort: number;
  cliFolder: string;
}

export class TestServerSetup extends TestSuiteSetup {
  child: ChildProcess;

  setupTest(): Promise<void> {
    return Promise.resolve();
  }
  teardownTest(): Promise<void> {
    return Promise.resolve();
  }
  async teardownSuite(): Promise<void> {
    await terminate(this.child.pid);
    return undefined;
  }

  public async setupSuite(
    testSuiteParameter: TestServerSetupParameter
  ): Promise<void> {
    // Test to see if server is already running on PORT.
    if (await isPortOpen(testSuiteParameter.serverPort)) {
      console.error(
        `Cannot run tests, server is already running on ${testSuiteParameter.serverPort}`
      );
      process.exit(1);
    }

    await execPromise(`mkdir -p ${testSuiteParameter.cliFolder}`);
    await execPromise(`rm -rf ${testSuiteParameter.cliFolder}/*`);

    await execPromise(
      `${CLI_COMMAND} init --project ${testSuiteParameter.cliFolder}`
    );
    await execPromise(
      `${CLI_COMMAND} import-source --project ${testSuiteParameter.cliFolder} ./test/data/Users.parquet`
    );
    await execPromise(
      `${CLI_COMMAND} import-source --project ${testSuiteParameter.cliFolder} ./test/data/AdImpressions.parquet`
    );
    await execPromise(
      `${CLI_COMMAND} import-source --project ${testSuiteParameter.cliFolder} ./test/data/AdBids.parquet`
    );

    let serverStarted = false;

    // Run Rill Developer in the background, logging to stdout.
    this.child = exec(
      `${CLI_COMMAND} start --project ${testSuiteParameter.cliFolder}`,
      {
        env: {
          ...process.env,
          RILL_SERVER_PORT: testSuiteParameter.serverPort + "",
        },
      }
    );
    this.child.stdout.pipe(process.stdout);
    // Watch for server startup in output.
    this.child.stdout.on("data", (data) => {
      if (data.startsWith("Server started at")) {
        serverStarted = true;
      }
    });
    // Terminate if the process exits.
    process.on("exit", async () => {
      await terminate(this.child.pid);
    });

    return new Promise((resolve) => {
      waitUntil(() => serverStarted).then(() => resolve());
    });
  }
}
