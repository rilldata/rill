import { rmdirSync, rmSync } from "fs";
import { execSync } from "node:child_process";
import { getTestConfig } from "./getTestConfig";
import { InlineTestServer } from "./InlineTestServer";
import type { TestServer } from "./TestServer";

/**
 * Creates a server with 'port' in the same process.
 * Make sure to use unique port for each suite.
 *
 * Automatically starts and stops the server.
 * Returns config and the server reference.
 * Check {@link TestServer} and {@link InlineTestServer} for various methods.
 *
 * TODO: auto assign port
 */
export function useInlineTestServer(port: number, folder = "temp/test") {
  const config = getTestConfig(folder, {
    socketPort: port,
    serveStaticFile: true,
  });

  const inlineServer = new InlineTestServer(config);

  beforeAll(async () => {
    rmSync(folder, {
      force: true,
      recursive: true,
    });
    await inlineServer.init();
  });

  afterAll(async () => {
    await inlineServer.destroy();
  });

  return {
    config,
    inlineServer,
  };
}

/**
 * Call this at the top level of suite to load test tables.
 */
export function useTestTables(server: TestServer) {
  beforeAll(async () => {
    await server.loadTestTables();
  });
}

/**
 * Call this at the top level of suite to load a model with given query and name.
 * Make sure to call {@link useTestTables} before this.
 */
export function useTestModel(server: TestServer, query: string, name: string) {
  beforeAll(async () => {
    await server.dataModelerService.dispatch("addModel", [{ query, name }]);
    await server.waitForModels();
  });
}
