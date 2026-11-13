import {
  adminServiceGetProject,
  getAdminServiceGetProjectQueryKey,
  V1DeploymentStatus,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import {
  createRuntimeJwtRefresher,
  getProjectQueryOptions,
} from "@rilldata/web-admin/features/projects/project-query-options";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { JWT_REFRESH_TIMEOUT } from "@rilldata/web-common/runtime-client/constants";
import type { JwtRefresher } from "@rilldata/web-common/runtime-client/v2";
import { QueryObserver } from "@tanstack/svelte-query";
import { CanceledError } from "axios";
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
  const compute = getProjectQueryOptions(vi.fn()).refetchInterval;
  if (typeof compute !== "function") {
    throw new Error("expected refetchInterval to be a function");
  }
  return compute({ state } as unknown as Parameters<typeof compute>[0]);
}

describe("getProjectQueryOptions refetchInterval", () => {
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

    /** Makes the first GetProject attempts hang until they are aborted, like axios requests on a dead connection. */
    function stallAttempts(count: number): AbortSignal[] {
      vi.useFakeTimers();
      const signals: AbortSignal[] = [];
      getProject
        .mockReset()
        .mockImplementation((_org, _project, _params, signal) => {
          if (!signal) throw new Error("expected the query client's signal");
          signals.push(signal);
          if (signals.length > count) return Promise.resolve({ jwt: "fresh" });
          // Axios rejects with a CanceledError once the signal aborts
          return new Promise((_, reject) =>
            signal.addEventListener("abort", () => reject(new CanceledError())),
          );
        });
      return signals;
    }

    it("retries network and server errors with backoff", async () => {
      vi.useFakeTimers();
      getProject
        .mockReset()
        .mockRejectedValueOnce(networkError)
        .mockRejectedValueOnce(httpError(503))
        .mockResolvedValueOnce({ jwt: "fresh" });

      const jwt = userRefresher()();
      await vi.advanceTimersByTimeAsync(10_000);

      await expect(jwt).resolves.toBe("fresh");
      expect(getProject).toHaveBeenCalledTimes(3);
    });

    it("fails at once when the session or the project access is gone", async () => {
      getProject.mockReset().mockRejectedValue(httpError(403));

      await expect(userRefresher()()).rejects.toEqual(httpError(403));
      expect(getProject).toHaveBeenCalledTimes(1);
    });

    it("abandons an attempt that stalls and retries it", async () => {
      const signals = stallAttempts(1);

      const jwt = userRefresher()();
      await vi.advanceTimersByTimeAsync(11_000);
      expect(getProject).toHaveBeenCalledTimes(1);

      await vi.advanceTimersByTimeAsync(5_000);
      await expect(jwt).resolves.toBe("fresh");
      expect(signals[0].aborted).toBe(true);
      expect(getProject).toHaveBeenCalledTimes(2);
    });

    it("aborts the request when its query is cancelled", async () => {
      const signals = stallAttempts(1);

      const settled = userRefresher()().catch(() => {});
      await queryClient.cancelQueries({
        queryKey: getAdminServiceGetProjectQueryKey("org", "proj"),
      });
      await settled;

      expect(signals[0].aborted).toBe(true);
      expect(getProject).toHaveBeenCalledTimes(1);
    });

    it("gives up on a refresh that keeps stalling before the requests waiting for it do", async () => {
      stallAttempts(Infinity);

      let failure: unknown;
      userRefresher()().catch((err: unknown) => {
        failure = err;
      });
      await vi.advanceTimersByTimeAsync(JWT_REFRESH_TIMEOUT - 1);

      expect(failure).toBeInstanceOf(CanceledError);
      expect(getProject).toHaveBeenCalledTimes(4);
    });

    it("is not held up by a stalled poll that it joins", async () => {
      const signals = stallAttempts(1);
      // The layout's query, e.g. a scheduled poll left hanging by a laptop sleep
      const poll = new QueryObserver(queryClient, {
        queryKey: getAdminServiceGetProjectQueryKey("org", "proj"),
        ...getProjectQueryOptions((signal) =>
          adminServiceGetProject("org", "proj", undefined, signal),
        ),
      });
      const unsubscribe = poll.subscribe(() => {});
      await vi.advanceTimersByTimeAsync(5_000);
      expect(getProject).toHaveBeenCalledTimes(1);

      // The refresh joins the poll's request instead of sending its own
      const jwt = userRefresher()();
      await vi.advanceTimersByTimeAsync(6_000);
      expect(getProject).toHaveBeenCalledTimes(1);

      await vi.advanceTimersByTimeAsync(5_000);
      await expect(jwt).resolves.toBe("fresh");
      expect(signals[0].aborted).toBe(true);
      expect(getProject).toHaveBeenCalledTimes(2);
      unsubscribe();
    });

    it("refetches even when the cached token is seconds old", async () => {
      // The runtime may have just rejected that token, so handing it back would not help
      queryClient.setQueryData(["/v1/orgs/org/projects/proj"], {
        jwt: "rejected",
      });
      getProject.mockReset().mockResolvedValue({ jwt: "fresh" });

      await expect(userRefresher()()).resolves.toBe("fresh");
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
      const jwts = Promise.all([refresh(), refresh()]);
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
