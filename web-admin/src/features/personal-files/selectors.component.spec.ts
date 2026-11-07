// @vitest-environment jsdom
import { getAdminServiceListPersonalFilesQueryKey } from "@rilldata/web-admin/client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { cleanup, render, screen } from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";
import PersonalFilteredResourcesTest from "./test/PersonalFilteredResourcesTest.svelte";

const mocks = vi.hoisted(() => ({ httpClient: vi.fn() }));
vi.mock("@rilldata/web-admin/client/http-client", () => ({
  httpClient: mocks.httpClient,
}));

function canvas(name: string) {
  return { meta: { name: { kind: ResourceKind.Canvas, name } } };
}

afterEach(() => {
  cleanup();
  queryClient.clear();
  vi.clearAllMocks();
});

describe("getPersonalFilteredResources", () => {
  it("refetches both lists when mounted again, so a dashboard saved elsewhere shows up", async () => {
    // The lists cached on a first visit, before the AI (or an MCP client)
    // saved "new" away from this page.
    queryClient.setQueryData(
      getAdminServiceListPersonalFilesQueryKey("acme", "ecom"),
      { files: ["old"] },
    );
    queryClient.setQueryData(getRuntimeServiceListResourcesQueryKey("i1", {}), {
      resources: [canvas("old")],
    });
    mocks.httpClient.mockResolvedValue({ files: ["old", "new"] });
    const listResources = vi.fn().mockResolvedValue({
      toJson: () => ({ resources: [canvas("old"), canvas("new")] }),
    });
    const client = {
      instanceId: "i1",
      runtimeService: { listResources },
    } as unknown as RuntimeClient;

    render(PersonalFilteredResourcesTest, { client });

    expect(await screen.findByText("new")).toBeTruthy();
    expect(mocks.httpClient).toHaveBeenCalledOnce();
    expect(listResources).toHaveBeenCalledOnce();
  });
});
