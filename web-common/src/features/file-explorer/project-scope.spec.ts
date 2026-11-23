import { featureFlags } from "@rilldata/web-common/features/feature-flags";
import { renderInRuntimeContext } from "@rilldata/web-common/features/metrics-views/providers/test/metrics-views-test-utils.svelte.ts";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useProjectScopeId } from "./project-scope";

const GET_METADATA_PATH = "/rill.local.v1.LocalService/GetMetadata";
const PROJECT_PATH = "/Users/jane/my-project";

/** Answers GetMetadata the way the Rill Developer server does: a Connect unary response in JSON. */
function serveGetMetadata(status: number, body: object) {
  const fetchMock = vi.fn((url: string | URL) => {
    if (!url.toString().endsWith(GET_METADATA_PATH)) {
      return Promise.reject(new Error(`unexpected request: ${url.toString()}`));
    }
    return Promise.resolve(
      new Response(JSON.stringify(body), {
        status,
        headers: { "content-type": "application/json" },
      }),
    );
  });
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

/** Gives a query that was (wrongly) started time to reach the network. */
async function settle() {
  for (let i = 0; i < 3; i++) {
    await new Promise((resolve) => setTimeout(resolve, 0));
  }
}

describe("useProjectScopeId", () => {
  let destroy: (() => void) | undefined;

  beforeEach(() => {
    featureFlags.adminServer.resetToDefault();
  });

  afterEach(() => {
    destroy?.();
    destroy = undefined;
    featureFlags.adminServer.resetToDefault();
    vi.unstubAllGlobals();
  });

  // Stays subscribed for the rest of the test, as the file explorer does, so that the query is mounted.
  function renderScopeId() {
    const rendered = renderInRuntimeContext(({ runtimeClient }) =>
      useProjectScopeId(runtimeClient.instanceId),
    );
    const values: (string | undefined)[] = [];
    const unsubscribe = rendered.value.subscribe((value) => values.push(value));
    destroy = () => {
      unsubscribe();
      rendered.destroy();
    };
    return { values, instanceId: rendered.runtimeClient.instanceId };
  }

  it("scopes by the instanceId in Rill Cloud without calling GetMetadata", async () => {
    featureFlags.adminServer.set(true);
    const fetchMock = serveGetMetadata(501, {
      code: "unimplemented",
      message: "not implemented",
    });

    const { values, instanceId } = renderScopeId();
    await settle();

    expect(values).toEqual([instanceId]);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("scopes by the project path from GetMetadata in Rill Developer", async () => {
    const fetchMock = serveGetMetadata(200, { projectPath: PROJECT_PATH });

    const { values } = renderScopeId();
    await vi.waitFor(() => expect(values.at(-1)).toBe(PROJECT_PATH));

    // Unknown while loading, so the explorer is never scoped to the instanceId in between.
    expect(values).toEqual([undefined, PROJECT_PATH]);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("falls back to the instanceId when GetMetadata fails in Rill Developer", async () => {
    serveGetMetadata(500, {
      code: "internal",
      message: "metadata unavailable",
    });

    const { values, instanceId } = renderScopeId();
    await vi.waitFor(() => expect(values.at(-1)).toBe(instanceId));

    expect(values).toEqual([undefined, instanceId]);
  });
});
