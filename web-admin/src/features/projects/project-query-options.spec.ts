import { V1DeploymentStatus, type RpcStatus } from "@rilldata/web-admin/client";
import {
  baseGetProjectQueryOptions,
  createRuntimeJwtRefresher,
} from "@rilldata/web-admin/features/projects/project-query-options";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { afterEach, describe, expect, it, vi } from "vitest";

const FifteenMinutes = 15 * 60 * 1000;
const runningDeployment = {
  deployment: { status: V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING },
};

function httpError(status: number): RpcStatus {
  return { response: { status, data: { message: "nope" } } } as RpcStatus;
}

function refetchInterval(state: { data?: unknown; error?: unknown }) {
  const compute = baseGetProjectQueryOptions.refetchInterval;
  if (typeof compute !== "function") {
    throw new Error("expected refetchInterval to be a function");
  }
  return compute({ state } as unknown as Parameters<typeof compute>[0]);
}

describe("baseGetProjectQueryOptions.refetchInterval", () => {
  it("polls a running deployment at half the JWT TTL", () => {
    expect(refetchInterval({ data: runningDeployment })).toBe(FifteenMinutes);
  });

  it("keeps polling after a transient failure so the JWT still gets refreshed", () => {
    expect(
      refetchInterval({ data: runningDeployment, error: httpError(503) }),
    ).toBe(FifteenMinutes);
  });

  it("stops polling once the session or the project access is gone", () => {
    for (const status of [401, 403]) {
      expect(
        refetchInterval({ data: runningDeployment, error: httpError(status) }),
      ).toBe(false);
    }
  });
});

describe("createRuntimeJwtRefresher", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  function mockFetchQuery(data: unknown) {
    return vi
      .spyOn(queryClient, "fetchQuery")
      .mockImplementation(() => Promise.resolve(data) as never);
  }

  it("refetches GetProject for the signed-in user and returns its JWT", async () => {
    const fetchQuery = mockFetchQuery({ jwt: "fresh" });
    const refresh = createRuntimeJwtRefresher("user", {
      organization: "org",
      project: "proj",
      branch: "dev",
    });

    await expect(refresh?.()).resolves.toBe("fresh");
    expect(fetchQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        queryKey: ["/v1/orgs/org/projects/proj", { branch: "dev" }],
      }),
    );
  });

  it("refetches GetProject with the magic token on public URLs", async () => {
    const fetchQuery = mockFetchQuery({ jwt: "fresh" });
    const refresh = createRuntimeJwtRefresher("magic", {
      organization: "org",
      project: "proj",
      token: "magic",
    });

    await expect(refresh?.()).resolves.toBe("fresh");
    expect(fetchQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        queryKey: ["/v1/orgs/org/projects/proj", "token/magic"],
      }),
    );
  });

  it("refetches the impersonated user's deployment credentials for View As", async () => {
    const fetchQuery = mockFetchQuery({ accessToken: "fresh" });
    const refresh = createRuntimeJwtRefresher("mock", {
      organization: "org",
      project: "proj",
      branch: "dev",
      mockedUserId: "u1",
    });

    await expect(refresh?.()).resolves.toBe("fresh");
    expect(fetchQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        queryKey: [
          "/v1/orgs/org/projects/proj/credentials",
          { userId: "u1", branch: "dev" },
        ],
      }),
    );
  });

  it("has nothing to refetch for embeds", () => {
    expect(
      createRuntimeJwtRefresher("embed", {
        organization: "org",
        project: "proj",
      }),
    ).toBeUndefined();
  });
});
