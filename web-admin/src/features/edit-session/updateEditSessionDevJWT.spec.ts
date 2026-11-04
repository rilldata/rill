import { adminServiceGetDeployment } from "@rilldata/web-admin/client";
import { createUpdateEditSessionDevJWT } from "@rilldata/web-admin/features/edit-session/updateEditSessionDevJWT";
import { createQueryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "@rilldata/web-common/runtime-client/constants";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@rilldata/web-admin/client", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@rilldata/web-admin/client")>()),
  adminServiceGetDeployment: vi.fn(),
}));

describe("createUpdateEditSessionDevJWT", () => {
  const fetchMock = vi.fn<typeof fetch>();
  const getDeployment = vi.mocked(adminServiceGetDeployment);

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

  it("reissues an expired mock JWT instead of waiting for a refetch that never comes", async () => {
    getDeployment
      .mockResolvedValueOnce({ accessToken: "mock-token-1" })
      .mockResolvedValueOnce({ accessToken: "mock-token-2" });
    const client = createClient();
    const updateDevJWT = createUpdateEditSessionDevJWT(
      "deployment-1",
      "edit-session-jwt",
    );

    await updateDevJWT(createQueryClient(), client, {
      email: "viewer@example.com",
    });
    expect(client.getJwt()).toBe("mock-token-1");
    expect(client.authContext).toBe("mock");

    // The mock token's TTL passes without a scheduled refetch (e.g. the laptop was asleep)
    vi.advanceTimersByTime(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL);
    try {
      await client.runtimeService.ping({});
    } catch {
      // ConnectRPC may reject the minimal mock response; the request still went out
    }

    expect(getDeployment).toHaveBeenCalledTimes(2);
    expect(getDeployment.mock.calls[1][0]).toBe("deployment-1");
    expect(getDeployment.mock.calls[1][1].attributes).toMatchObject({
      email: "viewer@example.com",
      domain: "example.com",
    });
    expect(client.getJwt()).toBe("mock-token-2");
    const headers = fetchMock.mock.calls[0][1]?.headers as Headers;
    expect(headers.get("Authorization")).toBe("Bearer mock-token-2");
  });

  it("restores the edit session JWT and the user context when View As ends", async () => {
    getDeployment.mockResolvedValueOnce({ accessToken: "mock-token" });
    const client = createClient();
    const queryClient = createQueryClient();
    const updateDevJWT = createUpdateEditSessionDevJWT(
      "deployment-1",
      "edit-session-jwt",
    );

    await updateDevJWT(queryClient, client, { email: "viewer@example.com" });
    await updateDevJWT(queryClient, client, null);

    expect(client.getJwt()).toBe("edit-session-jwt");
    expect(client.authContext).toBe("user");
  });
});
