import {
  adminServiceGetProject,
  V1DeploymentStatus,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import {
  baseGetProjectQueryOptions,
  createRuntimeJwtRefresher,
} from "@rilldata/web-admin/features/projects/project-query-options";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { JwtRefresher } from "@rilldata/web-common/runtime-client/v2";
import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("@rilldata/web-admin/client", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@rilldata/web-admin/client")>()),
  adminServiceGetProject: vi.fn(),
}));

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
  const neverAborted = new AbortController().signal;

  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
    queryClient.clear();
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

    await expect(refresh?.(neverAborted)).resolves.toBe("fresh");
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

    await expect(refresh?.(neverAborted)).resolves.toBe("fresh");
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

    await expect(refresh?.(neverAborted)).resolves.toBe("fresh");
    expect(fetchQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        queryKey: [
          "/v1/orgs/org/projects/proj/credentials",
          { userId: "u1", branch: "dev" },
        ],
      }),
    );
  });

  it("cancels the admin request when the client abandons the refresh", () => {
    vi.spyOn(queryClient, "fetchQuery").mockImplementation(
      () => new Promise(() => {}) as never,
    );
    const cancelQueries = vi
      .spyOn(queryClient, "cancelQueries")
      .mockResolvedValue();
    const abort = new AbortController();
    const refresh = createRuntimeJwtRefresher("user", {
      organization: "org",
      project: "proj",
    });

    void refresh?.(abort.signal);
    abort.abort();

    expect(cancelQueries).toHaveBeenCalledWith({
      queryKey: ["/v1/orgs/org/projects/proj"],
    });
  });

  describe("through the query client", () => {
    const getProject = vi.mocked(adminServiceGetProject);
    const networkError = new Error("Network Error");

    function userRefresher(): JwtRefresher {
      const refresh = createRuntimeJwtRefresher("user", {
        organization: "org",
        project: "proj",
      });
      if (!refresh) throw new Error("expected a refresher for the user");
      return refresh;
    }

    it("retries network and server errors with backoff", async () => {
      vi.useFakeTimers();
      getProject
        .mockReset()
        .mockRejectedValueOnce(networkError)
        .mockRejectedValueOnce(httpError(503))
        .mockResolvedValueOnce({ jwt: "fresh" });

      const jwt = userRefresher()(neverAborted);
      await vi.advanceTimersByTimeAsync(10_000);

      await expect(jwt).resolves.toBe("fresh");
      expect(getProject).toHaveBeenCalledTimes(3);
    });

    it("fails at once when the session or the project access is gone", async () => {
      getProject.mockReset().mockRejectedValue(httpError(403));

      await expect(userRefresher()(neverAborted)).rejects.toEqual(
        httpError(403),
      );
      expect(getProject).toHaveBeenCalledTimes(1);
    });

    it("shares one request between concurrent refreshes", async () => {
      let resolveProject!: (data: { jwt: string }) => void;
      getProject.mockReset().mockImplementation(
        () =>
          new Promise((resolve) => {
            resolveProject = resolve;
          }),
      );

      const refresh = userRefresher();
      const jwts = Promise.all([refresh(neverAborted), refresh(neverAborted)]);
      resolveProject({ jwt: "fresh" });

      await expect(jwts).resolves.toEqual(["fresh", "fresh"]);
      expect(getProject).toHaveBeenCalledTimes(1);
    });
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
