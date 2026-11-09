import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import { QueryClient } from "@tanstack/svelte-query";
import { overwriteGetLocale } from "../../../lib/i18n/gen/runtime";
import { RuntimeClient } from "../../../runtime-client/v2";
import {
  getRuntimeServiceGetResourceQueryKey,
  V1ReconcileStatus,
  type V1Resource,
} from "../../../runtime-client";
import Harness from "./test/BackfillPartitionsHarness.svelte";

const resource: V1Resource = {
  meta: {
    name: { kind: "rill.runtime.v1.Model", name: "hourly" },
    reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_IDLE,
  },
  model: {
    spec: { incremental: true, partitionsResolver: "sql" },
    state: { partitionsConcurrency: 1, partitionsModelId: "catalog-a" },
  },
};

let client: RuntimeClient;
let queryClient: QueryClient;
let failNextPage: boolean;
let latestResource: V1Resource;
let submissions: unknown[];
let partitionRequestCount: number;
const scrollIntoView = Object.getOwnPropertyDescriptor(
  HTMLElement.prototype,
  "scrollIntoView",
);

beforeEach(() => {
  Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
    configurable: true,
    value: vi.fn(),
  });
  overwriteGetLocale(() => "en");
  failNextPage = false;
  latestResource = resource;
  submissions = [];
  partitionRequestCount = 0;
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, networkMode: "always" } },
  });
  vi.stubGlobal(
    "fetch",
    vi.fn<typeof fetch>().mockImplementation(async (input, init) => {
      const url = input instanceof Request ? input.url : input.toString();
      const body =
        typeof init?.body === "string"
          ? init.body
          : ArrayBuffer.isView(init?.body)
            ? new TextDecoder().decode(init.body)
            : "{}";
      if (url.endsWith("/GetModelPartitions")) {
        partitionRequestCount++;
        const request: unknown = JSON.parse(body);
        const nextPage =
          request !== null &&
          typeof request === "object" &&
          "pageToken" in request &&
          request.pageToken === "next";
        if (nextPage && failNextPage)
          return new Response(
            JSON.stringify({
              code: "unavailable",
              message: "Partition page failed",
            }),
            { status: 503, headers: { "content-type": "application/json" } },
          );
        return response(
          nextPage
            ? {
                partitions: [
                  {
                    key: "p2",
                    data: { hour: "2" },
                    executedOn: "2026-03-01T00:00:00Z",
                    skipped: true,
                  },
                ],
              }
            : {
                partitions: [0, 1].map((hour) => ({
                  key: `p${hour}`,
                  data: { hour: String(hour) },
                  executedOn: "2026-03-01T00:00:00Z",
                })),
                nextPageToken: "next",
              },
        );
      }
      if (url.endsWith("/GetResource"))
        return response({ resource: latestResource });
      if (url.endsWith("/CreateTrigger")) {
        submissions.push(JSON.parse(body));
        return response({});
      }
      throw new Error(`Unexpected runtime request: ${url}`);
    }),
  );
  client = new RuntimeClient({
    host: "http://runtime.test",
    instanceId: "test",
  });
});

afterEach(() => {
  cleanup();
  queryClient.clear();
  client.dispose();
  vi.unstubAllGlobals();
  if (scrollIntoView)
    Object.defineProperty(
      HTMLElement.prototype,
      "scrollIntoView",
      scrollIntoView,
    );
  else Reflect.deleteProperty(HTMLElement.prototype, "scrollIntoView");
});

function response(body: unknown) {
  return new Response(JSON.stringify(body), {
    headers: { "content-type": "application/json" },
  });
}

async function preview() {
  const button = screen.getByRole("button", { name: "Preview partitions" });
  await waitFor(() => expect(button).toBeEnabled());
  const field = screen.getByRole("button", { name: "Partition field" });
  field.focus();
  await fireEvent.keyDown(field, { key: "ArrowDown" });
  await screen.findByRole("option", { name: "hour" });
  await fireEvent.keyDown(field, { key: "Enter" });
  await fireEvent.input(screen.getByLabelText("Start (inclusive)"), {
    target: { value: "0" },
  });
  await fireEvent.input(screen.getByLabelText("End (inclusive)"), {
    target: { value: "2" },
  });
  await fireEvent.click(button);
  return screen.findByRole("button", { name: "Backfill 3 partitions" });
}

async function refetchLiveResource() {
  await queryClient.invalidateQueries({
    queryKey: getRuntimeServiceGetResourceQueryKey("test", {
      name: { kind: "rill.runtime.v1.Model", name: "hourly" },
    }),
    exact: true,
  });
}

describe("partition backfill preview", () => {
  it("blocks a partial preview if a later page fails", async () => {
    failNextPage = true;
    render(Harness, { client, queryClient, resource, onTriggered: vi.fn() });
    await screen.findByText(/Partition page failed/);
    expect(
      screen.getByRole("button", { name: "Preview partitions" }),
    ).toBeDisabled();
    expect(submissions).toEqual([]);
  });

  it("submits the remaining selection across pages even when search hides rows", async () => {
    const onTriggered = vi.fn();
    render(Harness, { client, queryClient, resource, onTriggered });
    await preview();
    await fireEvent.click(screen.getAllByRole("button", { name: "Remove" })[0]);
    await fireEvent.input(
      screen.getByPlaceholderText("Search selected partitions"),
      { target: { value: "p1" } },
    );
    await fireEvent.click(
      screen.getByRole("button", { name: "Backfill 2 partitions" }),
    );
    await waitFor(() => expect(onTriggered).toHaveBeenCalledOnce());
    expect(submissions).toEqual([
      {
        instanceId: "test",
        models: [{ model: "hourly", partitions: ["p1", "p2"] }],
      },
    ]);
  });

  it("uses the live model status instead of the dialog's idle snapshot", async () => {
    render(Harness, { client, queryClient, resource, onTriggered: vi.fn() });
    const button = await preview();
    expect(button).toBeEnabled();
    latestResource = {
      ...resource,
      meta: {
        ...resource.meta,
        reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
      },
    };
    await queryClient.invalidateQueries();
    await waitFor(() => expect(button).toBeDisabled());
    expect(submissions).toEqual([]);
  });

  it("blocks submission while the live model is pending", async () => {
    render(Harness, { client, queryClient, resource, onTriggered: vi.fn() });
    const button = await preview();
    expect(button).toBeEnabled();
    latestResource = {
      ...resource,
      meta: {
        ...resource.meta,
        reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_PENDING,
      },
    };
    await refetchLiveResource();
    await waitFor(() => expect(button).toBeDisabled());
    expect(
      screen.queryByRole("button", { name: "Reset selection" }),
    ).not.toBeInTheDocument();
    expect(submissions).toEqual([]);
  });

  it("blocks submission if the live model is no longer partitioned", async () => {
    render(Harness, { client, queryClient, resource, onTriggered: vi.fn() });
    const button = await preview();
    expect(button).toBeEnabled();
    latestResource = {
      ...resource,
      model: { spec: { incremental: true }, state: {} },
    };
    await refetchLiveResource();
    await waitFor(() => expect(button).toBeDisabled());
    expect(submissions).toEqual([]);
  });

  it("refetches a changed partition catalog only after reconciliation", async () => {
    render(Harness, { client, queryClient, resource, onTriggered: vi.fn() });
    await preview();
    const requestsBeforeChange = partitionRequestCount;
    latestResource = {
      ...resource,
      meta: {
        ...resource.meta,
        reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
      },
      model: {
        ...resource.model,
        state: {
          ...resource.model?.state,
          partitionsModelId: "catalog-b",
        },
      },
    };
    await refetchLiveResource();
    await waitFor(() =>
      expect(
        screen.getByRole("button", { name: "Backfill 0 partitions" }),
      ).toBeDisabled(),
    );
    expect(partitionRequestCount).toBe(requestsBeforeChange);

    latestResource = {
      ...latestResource,
      meta: {
        ...latestResource.meta,
        reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_IDLE,
      },
      model: {
        ...latestResource.model,
        state: {
          ...latestResource.model?.state,
          refreshedOn: "2026-03-02T00:00:00Z",
        },
      },
    };
    await refetchLiveResource();
    await waitFor(() =>
      expect(partitionRequestCount).toBeGreaterThan(requestsBeforeChange),
    );
    expect(submissions).toEqual([]);
  });

  it("invalidates and refetches a preview after a same-catalog reconciliation", async () => {
    render(Harness, { client, queryClient, resource, onTriggered: vi.fn() });
    await preview();
    const requestsBeforeReconcile = partitionRequestCount;
    latestResource = {
      ...resource,
      meta: {
        ...resource.meta,
        reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
      },
    };
    await refetchLiveResource();
    await waitFor(() =>
      expect(
        screen.getByRole("button", { name: "Backfill 0 partitions" }),
      ).toBeDisabled(),
    );
    expect(partitionRequestCount).toBe(requestsBeforeReconcile);

    latestResource = {
      ...resource,
      model: {
        ...resource.model,
        state: {
          ...resource.model?.state,
          refreshedOn: "2026-03-02T00:00:00Z",
        },
      },
    };
    await refetchLiveResource();
    await waitFor(() =>
      expect(partitionRequestCount).toBeGreaterThan(requestsBeforeReconcile),
    );
    expect(submissions).toEqual([]);
  });

  it("blocks submission if partition refetch fails after preview", async () => {
    render(Harness, { client, queryClient, resource, onTriggered: vi.fn() });
    const button = await preview();
    failNextPage = true;
    await queryClient.invalidateQueries();
    await screen.findByText(/Partition page failed/);
    expect(button).toBeDisabled();
    expect(submissions).toEqual([]);
  });

  it("allows removing every partition and resetting the selection", async () => {
    render(Harness, { client, queryClient, resource, onTriggered: vi.fn() });
    await preview();
    while (screen.queryAllByRole("button", { name: "Remove" }).length) {
      await fireEvent.click(
        screen.getAllByRole("button", { name: "Remove" })[0],
      );
    }
    expect(
      screen.getByRole("button", { name: "Backfill 0 partitions" }),
    ).toBeDisabled();
    await fireEvent.click(
      screen.getByRole("button", { name: "Reset selection" }),
    );
    expect(
      screen.getByRole("button", { name: "Backfill 3 partitions" }),
    ).toBeEnabled();
  });

  it("invalidates the selection when a range bound changes", async () => {
    render(Harness, { client, queryClient, resource, onTriggered: vi.fn() });
    await preview();
    await fireEvent.input(screen.getByLabelText("Start (inclusive)"), {
      target: { value: "1" },
    });
    expect(
      screen.getByRole("button", { name: "Backfill 0 partitions" }),
    ).toBeDisabled();
    expect(submissions).toEqual([]);
  });

  it("localizes the preview table", async () => {
    overwriteGetLocale(() => "es");
    render(Harness, { client, queryClient, resource, onTriggered: vi.fn() });
    const previewButton = screen.getByRole("button", {
      name: "Previsualizar particiones",
    });
    await waitFor(() => expect(previewButton).toBeEnabled());
    const field = screen.getByRole("button", { name: "Campo de partición" });
    field.focus();
    await fireEvent.keyDown(field, { key: "ArrowDown" });
    await screen.findByRole("option", { name: "hour" });
    await fireEvent.keyDown(field, { key: "Enter" });
    await fireEvent.input(screen.getByLabelText("Inicio (inclusive)"), {
      target: { value: "0" },
    });
    await fireEvent.input(screen.getByLabelText("Fin (inclusive)"), {
      target: { value: "2" },
    });
    await fireEvent.click(previewButton);
    await screen.findByRole("button", { name: "Rellenar 3 particiones" });
    expect(
      screen.getByRole("columnheader", { name: "Datos" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("columnheader", { name: "Ejecutado el" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Omitida")).toBeInTheDocument();
  });
});
