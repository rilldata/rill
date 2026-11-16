import { adminServiceGetDeployment } from "@rilldata/web-admin/client";
import {
  createMockUserJwtRefresher,
  createUpdateEditSessionDevJWT,
  editSessionRuntimeAuth,
} from "@rilldata/web-admin/features/edit-session/updateEditSessionDevJWT";
import {
  selectedMockUserJWT,
  selectedMockUserStore,
} from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
import {
  createQueryClient,
  queryClient as globalQueryClient,
} from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL,
  RUNTIME_ACCESS_TOKEN_DEFAULT_TTL,
} from "@rilldata/web-common/runtime-client/constants";
import {
  RuntimeClient,
  type JwtRefresher,
} from "@rilldata/web-common/runtime-client/v2";
import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@rilldata/web-admin/client", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@rilldata/web-admin/client")>()),
  adminServiceGetDeployment: vi.fn(),
}));

describe("edit session View As", () => {
  const fetchMock = vi.fn<typeof fetch>();
  const getDeployment = vi.mocked(adminServiceGetDeployment);
  const mockUser = { email: "viewer@example.com" };

  beforeEach(() => {
    vi.useFakeTimers();
    fetchMock.mockReset().mockImplementation(() =>
      Promise.resolve(
        new Response(new Uint8Array(0), {
          status: 200,
          headers: { "content-type": "application/proto" },
        }),
      ),
    );
    vi.stubGlobal("fetch", fetchMock);
    getDeployment.mockReset();
    // The mock user refresher fetches through the global query client
    globalQueryClient.clear();
    selectedMockUserStore.set(null);
    selectedMockUserJWT.set(null);
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  function createClient() {
    return new RuntimeClient({
      host: "http://localhost:9009",
      instanceId: "test-instance",
      jwt: "edit-session-jwt",
    });
  }

  /** Sends a request; ConnectRPC may reject the minimal mock response, which does not matter here. */
  async function ping(client: RuntimeClient) {
    try {
      await client.runtimeService.ping({});
    } catch {
      // no-op
    }
  }

  /** Lets requests waiting for a fresh JWT see the token that was just applied. */
  function pollForFreshJwt() {
    return vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
  }

  function sentAuthHeader(index: number) {
    const headers = fetchMock.mock.calls[index]?.[1]?.headers as
      | Headers
      | undefined;
    return headers?.get("Authorization") ?? null;
  }

  it("installs the mock user's JWT and restores the user's when View As ends", async () => {
    getDeployment.mockResolvedValueOnce({ accessToken: "mock-token" });
    const client = createClient();
    const queryClient = createQueryClient();
    const updateDevJWT = createUpdateEditSessionDevJWT(
      "deployment-1",
      "edit-session-jwt",
    );

    await updateDevJWT(queryClient, client, mockUser);
    expect(client.getJwt()).toBe("mock-token");
    expect(client.authContext).toBe("mock");
    expect(get(selectedMockUserStore)).toBe(mockUser);
    expect(get(selectedMockUserJWT)).toBe("mock-token");

    await updateDevJWT(queryClient, client, null);
    expect(client.getJwt()).toBe("edit-session-jwt");
    expect(client.authContext).toBe("user");
    expect(get(selectedMockUserStore)).toBeNull();
    expect(get(selectedMockUserJWT)).toBeNull();
  });

  it("reissues the mock user's JWT with the same attributes", async () => {
    getDeployment.mockResolvedValueOnce({ accessToken: "mock-token-2" });
    selectedMockUserStore.set(mockUser);

    const refresh = createMockUserJwtRefresher("deployment-1", mockUser);
    await expect(refresh()).resolves.toBe("mock-token-2");

    expect(getDeployment).toHaveBeenCalledTimes(1);
    const [deploymentId, body, signal] = getDeployment.mock.calls[0];
    expect(deploymentId).toBe("deployment-1");
    expect(body.attributes).toMatchObject({
      email: "viewer@example.com",
      domain: "example.com",
    });
    expect(body.accessTokenTtlSeconds).toBe(
      RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 1000,
    );
    // The query client's signal, so cancelling the query aborts the request
    expect(signal).toBeInstanceOf(AbortSignal);
    expect(get(selectedMockUserJWT)).toBe("mock-token-2");
  });

  it("leaves the View As token alone when View As moved on during the reissue", async () => {
    getDeployment.mockResolvedValueOnce({ accessToken: "stale-viewer-token" });
    selectedMockUserStore.set(mockUser);
    selectedMockUserJWT.set("viewer-token");

    const refresh = createMockUserJwtRefresher("deployment-1", mockUser);
    const reissued = refresh();
    selectedMockUserStore.set({ email: "other@example.com" });
    selectedMockUserJWT.set("other-token");
    await reissued;

    expect(get(selectedMockUserJWT)).toBe("other-token");
  });

  it("keeps the mock credentials when a user JWT refresh lands after View As started", async () => {
    getDeployment
      .mockResolvedValueOnce({ accessToken: "mock-token" })
      .mockResolvedValueOnce({ accessToken: "mock-token-2" });
    const client = createClient();
    const queryClient = createQueryClient();
    let resolveUserRefresh!: (jwt: string) => void;
    const userRefreshJwt: JwtRefresher = vi.fn(
      () =>
        new Promise<string>((resolve) => {
          resolveUserRefresh = resolve;
        }),
    );
    const mockRefreshJwt = createMockUserJwtRefresher("deployment-1", mockUser);
    // Plays the edit layout and RuntimeProvider: the client gets whatever the layout
    // derives from the GetProject JWT and the View As selection
    function forwardToProvider(userJwt: string) {
      const auth = editSessionRuntimeAuth({
        userJwt,
        userRefreshJwt,
        mockUser: get(selectedMockUserStore),
        mockJwt: get(selectedMockUserJWT),
        mockRefreshJwt,
      });
      client.updateJwt(auth.jwt, auth.authContext);
      client.setJwtRefresher(auth.refreshJwt);
    }
    forwardToProvider("edit-session-jwt");

    // The user's JWT expires and a request starts an on-demand refresh
    vi.advanceTimersByTime(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL);
    const request = ping(client);
    await vi.advanceTimersByTimeAsync(1);
    expect(userRefreshJwt).toHaveBeenCalledTimes(1);

    // View As starts while that refresh is still in flight
    const updateDevJWT = createUpdateEditSessionDevJWT(
      "deployment-1",
      "edit-session-jwt",
    );
    await updateDevJWT(queryClient, client, mockUser);
    forwardToProvider("edit-session-jwt");
    await pollForFreshJwt();
    await request;
    expect(sentAuthHeader(0)).toBe("Bearer mock-token");

    // The user refresh completes and the layout forwards the new user JWT; the mock credentials stay
    resolveUserRefresh("edit-session-jwt-2");
    await vi.advanceTimersByTimeAsync(1);
    forwardToProvider("edit-session-jwt-2");
    expect(client.getJwt()).toBe("mock-token");
    expect(client.authContext).toBe("mock");

    // When the mock token expires, it is the mock issuer that is asked, not the user's
    vi.advanceTimersByTime(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL);
    const mockRequest = ping(client);
    await pollForFreshJwt();
    await mockRequest;
    expect(getDeployment).toHaveBeenCalledTimes(2);
    expect(userRefreshJwt).toHaveBeenCalledTimes(1);
    expect(client.getJwt()).toBe("mock-token-2");
    expect(sentAuthHeader(1)).toBe("Bearer mock-token-2");

    // View As ends: the latest user JWT applies
    await updateDevJWT(queryClient, client, null);
    forwardToProvider("edit-session-jwt-2");
    expect(client.getJwt()).toBe("edit-session-jwt-2");
    expect(client.authContext).toBe("user");
  });
});
