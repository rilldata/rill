import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  getRuntimeServiceGetInstanceQueryKey,
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceGetFile,
  runtimeServiceAnalyzeConnectors,
  runtimeServiceGetInstance,
  runtimeServiceGetResource,
  runtimeServiceListResources,
  runtimeServicePutFile,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  ResourceKind,
  SingletonProjectParserName,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { RuntimeFileIO } from "@rilldata/web-common/features/entity-management/file-io";
import { makeTestEnvEditSession } from "@rilldata/web-common/features/env-management/test/test-env-store";
import { clickhouseSchema } from "@rilldata/web-common/features/templates/schemas/clickhouse";
import { createConnector } from "./connector";

vi.mock("@rilldata/web-common/runtime-client", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@rilldata/web-common/runtime-client")
  >()),
  runtimeServiceGetFile: vi.fn(),
  runtimeServiceAnalyzeConnectors: vi.fn(),
  runtimeServiceGetInstance: vi.fn(),
  runtimeServiceGetResource: vi.fn(),
  runtimeServiceListResources: vi.fn(),
  runtimeServicePutFile: vi.fn(),
}));

vi.mock("@rilldata/web-common/features/welcome/is-project-initialized", () => ({
  isProjectInitialized: vi.fn().mockResolvedValue(true),
}));
vi.mock("$app/navigation", () => ({ invalidate: vi.fn() }));

const client = { instanceId: "clickhouse-readiness" } as RuntimeClient;
const connectorName = "clickhouse";
const connectorPath = "/connectors/clickhouse.yaml";
const parserKey = getRuntimeServiceGetResourceQueryKey(client.instanceId, {
  name: { name: SingletonProjectParserName, kind: ResourceKind.ProjectParser },
});
const connectorKey = getRuntimeServiceGetResourceQueryKey(client.instanceId, {
  name: { name: connectorName, kind: ResourceKind.Connector },
});
const explorerKey = getRuntimeServiceAnalyzeConnectorsQueryKey(
  client.instanceId,
);

function parser(version: number): V1Resource {
  return {
    meta: { version: String(version) },
    projectParser: { state: { watching: true } },
  };
}

function connector(reconcileError?: string, stateVersion = "2"): V1Resource {
  return {
    meta: {
      name: { name: connectorName, kind: ResourceKind.Connector },
      filePaths: [connectorPath],
      stateVersion,
      reconcileStatus: "RECONCILE_STATUS_IDLE",
      reconcileError,
    },
    connector: { spec: { driver: "clickhouse" } },
  };
}

function observe<T>(promise: Promise<T>) {
  const outcome: { status: string; value?: T; error?: unknown } = {
    status: "pending",
  };
  void promise.then(
    (value) => Object.assign(outcome, { status: "fulfilled", value }),
    (error: unknown) => Object.assign(outcome, { status: "rejected", error }),
  );
  return outcome;
}

async function submit(validate: boolean) {
  const { envEditSession } = await makeTestEnvEditSession(
    "clickhouse",
    clickhouseSchema,
  );
  return observe(
    createConnector({
      runtimeClient: client,
      queryClient,
      connectorName,
      connectorDriver: { name: "clickhouse" },
      formValues: { host: "localhost", username: "default", password: "test" },
      validate,
      envEditSession,
    }),
  );
}

describe("ClickHouse connector readiness", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    queryClient.clear();
    fileArtifacts.removeFile(connectorPath);
    fileArtifacts.setClient(client, new RuntimeFileIO());
    queryClient.setQueryData(parserKey, { resource: parser(1) });
    queryClient.setQueryDefaults(explorerKey, { staleTime: Infinity });
    queryClient.setQueryData(explorerKey, { connectors: [{ name: "duckdb" }] });
    vi.mocked(runtimeServiceAnalyzeConnectors).mockResolvedValue({
      connectors: [
        {
          name: connectorName,
          driver: { name: "clickhouse", implementsOlap: true },
        },
      ],
    });
    vi.mocked(runtimeServiceGetFile).mockResolvedValue({
      blob: "olap_connector: duckdb\n",
    });
    vi.mocked(runtimeServiceGetInstance).mockResolvedValue({
      instance: { olapConnector: connectorName },
    });
    vi.mocked(runtimeServiceGetResource).mockImplementation(
      async (_, params) => ({
        resource:
          params.name?.kind === ResourceKind.ProjectParser
            ? parser(2)
            : connector(),
      }),
    );
    vi.mocked(runtimeServiceListResources).mockResolvedValue({
      resources: [connector()],
    });
    vi.mocked(runtimeServicePutFile).mockResolvedValue({});
  });

  afterEach(() => {
    queryClient.clear();
    fileArtifacts.removeFile(connectorPath);
    vi.clearAllTimers();
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  // Replay both event schedules: the RPC succeeds in either case, but the
  // parser notification can be lost while the .env change restarts the stream.
  it.each([
    { deliverEvent: true, parseDelay: 0 },
    { deliverEvent: false, parseDelay: 0 },
    { deliverEvent: true, parseDelay: 300 },
    { deliverEvent: false, parseDelay: 300 },
    { deliverEvent: true, parseDelay: 900 },
    { deliverEvent: false, parseDelay: 900 },
  ])(
    "surfaces invalid credentials with event=$deliverEvent and parse delay=$parseDelay ms",
    async ({ deliverEvent, parseDelay }) => {
      let parsed = false;
      vi.mocked(runtimeServicePutFile).mockImplementation(async () => {
        setTimeout(() => {
          parsed = true;
          if (deliverEvent) {
            queryClient.setQueryData(parserKey, { resource: parser(2) });
          }
        }, parseDelay);
        return {};
      });
      vi.mocked(runtimeServiceGetResource).mockImplementation(
        async (_, params) => {
          if (params.name?.kind === ResourceKind.ProjectParser) {
            return { resource: parser(parsed ? 2 : 1) };
          }
          if (!parsed) throw { code: 5, message: "resource not found" };
          return { resource: connector("Authentication failed") };
        },
      );

      const outcome = await submit(true);
      await vi.advanceTimersByTimeAsync(4_000);

      expect(outcome).toMatchObject({
        status: "rejected",
        error: { details: "Authentication failed" },
      });
    },
  );

  it("surfaces new validation errors when the state version rolls from 9 to 10", async () => {
    queryClient.setQueryData(connectorKey, {
      resource: connector(undefined, "9"),
    });
    vi.mocked(runtimeServicePutFile).mockImplementation(async () => {
      queryClient.setQueryData(parserKey, { resource: parser(2) });
      return {};
    });
    vi.mocked(runtimeServiceGetResource).mockImplementation(
      async (_, params) => ({
        resource:
          params.name?.kind === ResourceKind.ProjectParser
            ? parser(2)
            : connector("Authentication failed", "10"),
      }),
    );

    const outcome = await submit(true);
    await vi.advanceTimersByTimeAsync(4_000);

    expect(outcome).toMatchObject({
      status: "rejected",
      error: { details: "Authentication failed" },
    });
  });

  it("validates and refreshes a successful connection without parser events", async () => {
    const outcome = await submit(true);
    await vi.advanceTimersByTimeAsync(1_000);

    expect(outcome).toMatchObject({
      status: "fulfilled",
      value: connectorPath,
    });
    expect(fileArtifacts.getNamesForKind(ResourceKind.Connector)).toContain(
      connectorName,
    );
    expect(queryClient.getQueryData(explorerKey)).toMatchObject({
      connectors: [{ name: connectorName }],
    });
  });

  it("recovers from a closed controller while validating credentials", async () => {
    const readyAt = Date.now() + 900;
    vi.mocked(runtimeServiceGetResource).mockImplementation(
      async (_, params) => {
        if (Date.now() < readyAt)
          throw { code: 14, message: "controller is closed" };
        return {
          resource:
            params.name?.kind === ResourceKind.ProjectParser
              ? parser(2)
              : connector("Authentication failed"),
        };
      },
    );

    const outcome = await submit(true);
    await vi.advanceTimersByTimeAsync(4_000);
    expect(outcome).toMatchObject({
      status: "rejected",
      error: { details: "Authentication failed" },
    });
  });

  it("reports a timeout if the parser never processes the file", async () => {
    vi.mocked(runtimeServiceGetResource).mockResolvedValue({
      resource: parser(1),
    });
    const outcome = await submit(true);
    await vi.advanceTimersByTimeAsync(21_000);

    expect(outcome).toMatchObject({
      status: "rejected",
      error: { message: "Timed out waiting for project parser version 2" },
    });
  });

  it("reports a timeout if the connector never finishes reconciling", async () => {
    const reconciling = connector();
    reconciling.meta!.reconcileStatus = "RECONCILE_STATUS_RUNNING";
    vi.mocked(runtimeServiceGetResource).mockImplementation(
      async (_, params) => ({
        resource:
          params.name?.kind === ResourceKind.ProjectParser
            ? parser(2)
            : reconciling,
      }),
    );
    const outcome = await submit(true);
    await vi.advanceTimersByTimeAsync(61_000);

    expect(outcome).toMatchObject({
      status: "rejected",
      error: {
        message: "Timed out waiting for resource clickhouse to reconcile",
      },
    });
  });

  it("reports a timeout if the OLAP setting is never applied", async () => {
    vi.mocked(runtimeServiceGetInstance).mockResolvedValue({
      instance: { olapConnector: "duckdb" },
    });
    const outcome = await submit(false);
    await vi.advanceTimersByTimeAsync(21_000);

    expect(outcome).toMatchObject({
      status: "rejected",
      error: { message: "Timed out waiting for OLAP connector clickhouse" },
    });
  });

  it("surfaces permission errors without retrying them", async () => {
    const error = { code: 7, message: "Permission denied" };
    vi.mocked(runtimeServiceGetInstance).mockRejectedValue(error);
    const outcome = await submit(false);
    await vi.advanceTimersByTimeAsync(1_000);

    expect(outcome).toMatchObject({ status: "rejected", error });
    expect(runtimeServiceGetInstance).toHaveBeenCalledTimes(1);
  });

  it("accepts a completed parser when repository watching is disabled", async () => {
    const completedParser = parser(2);
    completedParser.projectParser!.state!.watching = false;
    completedParser.meta!.reconcileStatus = "RECONCILE_STATUS_IDLE";
    vi.mocked(runtimeServiceGetResource).mockImplementation(
      async (_, params) => ({
        resource:
          params.name?.kind === ResourceKind.ProjectParser
            ? completedParser
            : connector(),
      }),
    );
    const outcome = await submit(false);
    await vi.advanceTimersByTimeAsync(1_000);

    expect(outcome).toMatchObject({
      status: "fulfilled",
      value: connectorPath,
    });
  });

  it.each([undefined, "Authentication failed"])(
    "waits for the OLAP change and refreshes the explorer when saving without testing (error=%s)",
    async (reconcileError) => {
      let configAppliedAt = Infinity;
      const savedConnector = connector(reconcileError);
      queryClient.setQueryDefaults(
        getRuntimeServiceListResourcesQueryKey(client.instanceId, {}),
        { staleTime: Infinity },
      );
      queryClient.setQueryData(
        getRuntimeServiceListResourcesQueryKey(client.instanceId, {}),
        { resources: [] },
      );
      queryClient.setQueryData(
        getRuntimeServiceGetInstanceQueryKey(client.instanceId, {
          sensitive: true,
        }),
        { instance: { olapConnector: "duckdb" } },
      );
      vi.mocked(runtimeServicePutFile).mockImplementation(async (_, body) => {
        if (body.path === "rill.yaml") configAppliedAt = Date.now() + 1_000;
        return {};
      });
      vi.mocked(runtimeServiceGetInstance).mockImplementation(async () => ({
        instance: {
          olapConnector:
            Date.now() >= configAppliedAt ? connectorName : "duckdb",
        },
      }));
      vi.mocked(runtimeServiceGetResource).mockImplementation(
        async (_, params) => {
          // Updating the instance precedes the end of the controller restart.
          if (
            Date.now() >= configAppliedAt &&
            Date.now() < configAppliedAt + 400
          ) {
            throw { code: 14, message: "controller is closed" };
          }
          return {
            resource:
              params.name?.kind === ResourceKind.ProjectParser
                ? parser(2)
                : savedConnector,
          };
        },
      );
      vi.mocked(runtimeServiceListResources).mockResolvedValue({
        resources: [savedConnector],
      });

      const outcome = await submit(false);
      await vi.advanceTimersByTimeAsync(100);
      expect(outcome.status).toBe("pending");

      await vi.advanceTimersByTimeAsync(3_000);
      expect(outcome).toMatchObject({
        status: "fulfilled",
        value: connectorPath,
      });
      expect(fileArtifacts.getNamesForKind(ResourceKind.Connector)).toContain(
        connectorName,
      );
      expect(queryClient.getQueryData(connectorKey)).toEqual({
        resource: savedConnector,
      });
      expect(queryClient.getQueryData(explorerKey)).toMatchObject({
        connectors: [{ name: connectorName }],
      });
    },
  );
});
