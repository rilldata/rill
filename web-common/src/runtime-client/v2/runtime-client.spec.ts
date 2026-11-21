import { Code, ConnectError } from "@connectrpc/connect";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { RuntimeClient, type AuthContext } from "./runtime-client";
import {
  RUNTIME_ACCESS_TOKEN_DEFAULT_TTL,
  JWT_EXPIRY_WARNING_WINDOW,
  CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL,
} from "../constants";

/** Time offset that places Date.now() just inside the expiry warning window. */
const PAST_WARNING_WINDOW =
  RUNTIME_ACCESS_TOKEN_DEFAULT_TTL - JWT_EXPIRY_WARNING_WINDOW + 1;

/**
 * Returns a mock fetch that returns a valid Connect-protocol unary response.
 * The response is an empty protobuf message (valid for any proto3 response type).
 */
function createFetchMock() {
  return vi.fn<typeof fetch>().mockImplementation(() =>
    Promise.resolve(
      new Response(new Uint8Array(0), {
        status: 200,
        headers: { "content-type": "application/proto" },
      }),
    ),
  );
}

/** Extract the Authorization header from the given fetch call. */
function getAuthHeader(
  fetchMock: ReturnType<typeof vi.fn>,
  index: number,
): string | null {
  const call = fetchMock.mock.calls[index] as
    | [RequestInfo, RequestInit | undefined]
    | undefined;
  const headers = call?.[1]?.headers as Headers | undefined;
  return headers?.get?.("Authorization") ?? null;
}

/** Extract the Authorization header from the most recent fetch call. */
function getLastAuthHeader(fetchMock: ReturnType<typeof vi.fn>): string | null {
  return getAuthHeader(fetchMock, fetchMock.mock.calls.length - 1);
}

/** Builds an unsigned JWT whose payload carries the given claims (times in seconds since the epoch). */
function makeJwt(claims: Record<string, unknown>): string {
  const payload = btoa(JSON.stringify(claims))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
  return `eyJhbGciOiJub25lIn0.${payload}.signature`;
}

/** A unary Connect error response as the runtime returns it for an expired token. */
function expiredJwtResponse() {
  return new Response(
    JSON.stringify({
      code: "unauthenticated",
      message: "rpc error: code = Unauthenticated desc = jwt is expired",
    }),
    { status: 401, headers: { "content-type": "application/json" } },
  );
}

describe("RuntimeClient JWT refresh", () => {
  let fetchMock: ReturnType<typeof createFetchMock>;

  beforeEach(() => {
    vi.useFakeTimers();
    fetchMock = createFetchMock();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  function createClient(
    opts: {
      jwt?: string;
      authContext?: AuthContext;
    } = {},
  ) {
    return new RuntimeClient({
      host: "http://localhost:9009",
      instanceId: "test-instance",
      jwt: opts.jwt,
      authContext: opts.authContext ?? "user",
    });
  }

  /**
   * Make a ping request through the RuntimeClient's transport.
   * The response may fail to parse (our mock is minimal), so we
   * catch errors and just inspect the fetch mock for the auth header.
   */
  async function makeRequest(client: RuntimeClient) {
    try {
      await client.runtimeService.ping({});
    } catch {
      // ConnectRPC may throw on our minimal mock response; that's fine
    }
  }

  /** Makes a request and lets the client's poll for a fresh JWT pick up the refreshed token. */
  async function makeRequestThroughRefresh(client: RuntimeClient) {
    const promise = makeRequest(client);
    await vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
    await promise;
  }

  // ── Basic auth header ──────────────────────────────────────────────

  describe("Authorization header", () => {
    it("sends Bearer token when JWT is set", async () => {
      const client = createClient({ jwt: "my-token" });
      await makeRequest(client);
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer my-token");
    });

    it("omits Authorization header when no JWT is provided", async () => {
      const client = createClient();
      await makeRequest(client);
      expect(fetchMock).toHaveBeenCalled();
      expect(getLastAuthHeader(fetchMock)).toBeNull();
    });
  });

  // ── Refresh behavior when JWT is near expiry ───────────────────────

  describe("waitForFreshJwt", () => {
    it("blocks request until updateJwt provides a fresh token", async () => {
      const client = createClient({ jwt: "old-token" });

      // Advance time so the JWT is within the expiry warning window
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      // Start request; it should block in the polling loop
      let resolved = false;
      const promise = makeRequest(client).then(() => {
        resolved = true;
      });

      // Let the polling loop run a couple iterations; fetch should NOT fire
      await vi.advanceTimersByTimeAsync(
        CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL * 2,
      );
      expect(resolved).toBe(false);
      expect(fetchMock).not.toHaveBeenCalled();

      // Provide a fresh JWT
      client.updateJwt("fresh-token");

      // Advance past the next poll so the loop picks up the new token
      await vi.advanceTimersByTimeAsync(
        CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL * 2,
      );
      await promise;

      expect(resolved).toBe(true);
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer fresh-token");
    });

    it("getFreshJwt waits for a refresh when the token is about to expire", async () => {
      const client = createClient({ jwt: "old-token" });
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      let jwt: string | undefined;
      const promise = client.getFreshJwt().then((value) => {
        jwt = value;
      });
      await vi.advanceTimersByTimeAsync(
        CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL * 2,
      );
      expect(jwt).toBeUndefined();

      client.updateJwt("fresh-token");
      await vi.advanceTimersByTimeAsync(
        CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL * 2,
      );
      await promise;
      expect(jwt).toBe("fresh-token");
    });

    it("sends request immediately when JWT is not near expiry", async () => {
      const client = createClient({ jwt: "valid-token" });

      // Advance time but stay well before the warning window
      vi.advanceTimersByTime(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 2);

      await makeRequest(client);
      expect(fetchMock).toHaveBeenCalled();
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer valid-token");
    });

    it("throws after 60s timeout if no refresh arrives", async () => {
      const client = createClient({ jwt: "expiring-token" });

      // Advance to within the warning window
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      // Start the request (will block). Attach .catch immediately so the
      // rejection is handled before ConnectRPC's internal abort fires.
      const promise = client.runtimeService.ping({}).catch((e: Error) => e);

      // Advance past the 60s deadline
      await vi.advanceTimersByTimeAsync(61_000);

      const error = await promise;
      expect(error).toBeInstanceOf(Error);
      expect((error as Error).message).toContain(
        "Timed out waiting for a fresh JWT",
      );
      expect(fetchMock).not.toHaveBeenCalled();
    });

    it("skips expiry check for embed auth context", async () => {
      const client = createClient({
        jwt: "embed-token",
        authContext: "embed",
      });

      // Advance well past the TTL
      vi.advanceTimersByTime(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL * 2);

      await makeRequest(client);
      expect(fetchMock).toHaveBeenCalled();
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer embed-token");
    });
  });

  // ── Dispose behavior ───────────────────────────────────────────────

  describe("dispose during waitForFreshJwt", () => {
    it("cancels with AbortError when disposed instead of sending stale JWT", async () => {
      const client = createClient({ jwt: "stale-token" });

      // Advance to within the warning window
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      // Start request (blocks in polling loop). Catch immediately to
      // handle ConnectRPC's internal abort rejection.
      const promise = client.runtimeService.ping({}).catch((e: Error) => e);

      // Let the polling loop start
      await vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);

      // Dispose the client while the request is waiting for a fresh JWT
      client.dispose();

      // Advance so the loop picks up the disposed flag
      await vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
      const error = await promise;

      // The request should be cancelled, NOT sent with a stale JWT.
      // waitForFreshJwt throws an AbortError (matching requestQueue.clear()),
      // which ConnectRPC wraps in a ConnectError before it reaches TanStack Query.
      expect(error).toBeInstanceOf(Error);
      expect(fetchMock).not.toHaveBeenCalled();
    });
  });

  // ── Laptop sleep / long idle recovery ──────────────────────────────

  describe("recovery after JWT expiry", () => {
    it("blocks requests when JWT has expired and waits for refresh", async () => {
      const client = createClient({ jwt: "original-token" });

      // Simulate laptop sleep: jump well past the TTL
      vi.advanceTimersByTime(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL + 60_000);

      // Start request; it should block (JWT is expired)
      let resolved = false;
      const promise = makeRequest(client).then(() => {
        resolved = true;
      });

      await vi.advanceTimersByTimeAsync(
        CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL * 2,
      );
      expect(resolved).toBe(false);
      expect(fetchMock).not.toHaveBeenCalled();

      // Simulate the project query refetch delivering a fresh JWT
      client.updateJwt("refreshed-token");

      await vi.advanceTimersByTimeAsync(
        CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL * 2,
      );
      await promise;

      expect(resolved).toBe(true);
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer refreshed-token");
    });

    it("does not block requests that still have time before expiry", async () => {
      const client = createClient({ jwt: "valid-token" });

      // Advance to 2 minutes before expiry (outside the 1s warning window)
      vi.advanceTimersByTime(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL - 2 * 60 * 1000);

      // Request should proceed immediately; 2 minutes of remaining
      // lifetime is well outside the 1-second network-latency buffer
      await makeRequest(client);
      expect(fetchMock).toHaveBeenCalled();
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer valid-token");
    });
  });

  // ── updateJwt ──────────────────────────────────────────────────────

  describe("updateJwt", () => {
    it("updates jwtReceivedAt when the token string changes", async () => {
      const client = createClient({ jwt: "token-v1" });

      // Advance to near expiry
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      // Update with a new token (simulates the 15-min refetch)
      client.updateJwt("token-v2");

      // The request should succeed immediately (fresh jwtReceivedAt)
      await makeRequest(client);
      expect(fetchMock).toHaveBeenCalled();
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer token-v2");
    });

    it("does NOT update jwtReceivedAt when the same token is passed", async () => {
      const client = createClient({ jwt: "same-token" });

      // Advance to near expiry
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      // Pass the same token again (no-op for receivedAt)
      client.updateJwt("same-token");

      // Request should still block because jwtReceivedAt wasn't refreshed
      let resolved = false;
      const promise = makeRequest(client).then(() => {
        resolved = true;
      });

      await vi.advanceTimersByTimeAsync(
        CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL * 2,
      );
      expect(resolved).toBe(false);
      expect(fetchMock).not.toHaveBeenCalled();

      // Now provide a genuinely new token
      client.updateJwt("actually-new-token");
      await vi.advanceTimersByTimeAsync(
        CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL * 2,
      );
      await promise;

      expect(getLastAuthHeader(fetchMock)).toBe("Bearer actually-new-token");
    });

    it("returns true when auth context changes", () => {
      const client = createClient({ jwt: "token", authContext: "user" });
      const changed = client.updateJwt("token", "mock");
      expect(changed).toBe(true);
    });

    it("returns false when auth context stays the same", () => {
      const client = createClient({ jwt: "token", authContext: "user" });
      const changed = client.updateJwt("new-token", "user");
      expect(changed).toBe(false);
    });

    it("exposes the auth context, including updates", () => {
      const client = createClient({ jwt: "token", authContext: "magic" });
      expect(client.authContext).toBe("magic");
      client.updateJwt("token", "user");
      expect(client.authContext).toBe("user");
    });
  });

  // ── Active refresh through the registered refresher ────────────────

  describe("refresher", () => {
    it("asks the refresher for a new JWT instead of waiting for updateJwt", async () => {
      const client = createClient({ jwt: "expiring-token" });
      const refreshJwt = vi.fn().mockResolvedValue("fresh-token");
      client.setJwtRefresher(refreshJwt);
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      await makeRequestThroughRefresh(client);

      expect(refreshJwt).toHaveBeenCalledTimes(1);
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer fresh-token");
    });

    it("releases every waiting request once the refreshed JWT lands", async () => {
      const client = createClient({ jwt: "expiring-token" });
      let resolveRefresh!: (jwt: string) => void;
      const refreshJwt = vi.fn(
        () =>
          new Promise<string>((resolve) => {
            resolveRefresh = resolve;
          }),
      );
      client.setJwtRefresher(refreshJwt);
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      const requests = Promise.all([
        makeRequest(client),
        makeRequest(client),
        makeRequest(client),
      ]);
      await vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
      expect(fetchMock).not.toHaveBeenCalled();

      resolveRefresh("fresh-token");
      await vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
      await requests;

      expect(fetchMock).toHaveBeenCalledTimes(3);
      for (let i = 0; i < 3; i++) {
        expect(getAuthHeader(fetchMock, i)).toBe("Bearer fresh-token");
      }
    });

    it("fails the request when the refresher fails, leaving retries to the refresher", async () => {
      const client = createClient({ jwt: "expiring-token" });
      const refreshJwt = vi.fn().mockRejectedValue({
        response: {
          status: 403,
          data: { message: "does not have permission to read project" },
        },
      });
      client.setJwtRefresher(refreshJwt);
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      const promise = client.runtimeService.ping({}).catch((e: Error) => e);
      await vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
      const error = await promise;

      expect(error).toBeInstanceOf(Error);
      expect((error as Error).message).toContain(
        "Failed to refresh the runtime JWT: does not have permission to read project",
      );
      expect(refreshJwt).toHaveBeenCalledTimes(1);
      expect(fetchMock).not.toHaveBeenCalled();
    });

    it("fails when the issuer hands back an already expired JWT", async () => {
      const client = createClient({ jwt: "expiring-token" });
      const now = Date.now() / 1000;
      const refreshJwt = vi
        .fn()
        .mockResolvedValue(makeJwt({ iat: now - 3600, exp: now - 60 }));
      client.setJwtRefresher(refreshJwt);
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      const promise = client.runtimeService.ping({}).catch((e: Error) => e);
      await vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
      const error = await promise;

      expect((error as Error).message).toContain(
        "Failed to refresh the runtime JWT: the issuer returned an expired JWT",
      );
      expect(fetchMock).not.toHaveBeenCalled();
    });

    it("gives up on a refresh that stalls after 60s", async () => {
      const client = createClient({ jwt: "expiring-token" });
      const refreshJwt = vi.fn(() => new Promise<string>(() => {}));
      client.setJwtRefresher(refreshJwt);
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      let settled = false;
      const promise = client.runtimeService
        .ping({})
        .catch((e: Error) => e)
        .finally(() => {
          settled = true;
        });
      await vi.advanceTimersByTimeAsync(59_000);
      expect(settled).toBe(false);

      await vi.advanceTimersByTimeAsync(2_000);
      const error = await promise;
      expect((error as Error).message).toContain(
        "Timed out waiting for a fresh JWT",
      );
      expect(refreshJwt).toHaveBeenCalledTimes(1);
      expect(fetchMock).not.toHaveBeenCalled();
    });

    it("leaves a refresh that one request gave up on to the requests still waiting for it", async () => {
      const client = createClient({ jwt: "expiring-token" });
      // Plays the query client: every caller joins the request already in flight
      let resolveRefresh!: (jwt: string) => void;
      const inFlight = new Promise<string>((resolve) => {
        resolveRefresh = resolve;
      });
      const refreshJwt = vi.fn(() => inFlight);
      client.setJwtRefresher(refreshJwt);
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      const first = client.runtimeService.ping({}).catch((e: Error) => e);
      await vi.advanceTimersByTimeAsync(50_000);
      let secondSettled = false;
      const second = makeRequest(client).then(() => {
        secondSettled = true;
      });

      // The first request runs out of time; the second keeps waiting on the same refresh
      await vi.advanceTimersByTimeAsync(11_000);
      expect(((await first) as Error).message).toContain(
        "Timed out waiting for a fresh JWT",
      );
      expect(secondSettled).toBe(false);

      resolveRefresh("fresh-token");
      await vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
      await second;
      expect(refreshJwt).toHaveBeenCalledTimes(2);
      expect(fetchMock).toHaveBeenCalledTimes(1);
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer fresh-token");
    });

    it("stops waiting as soon as the request is cancelled", async () => {
      // e.g. a query cancelled on navigation: it must not hold its queue slot until the refresh completes
      const client = createClient({ jwt: "expiring-token" });
      client.setJwtRefresher(vi.fn(() => new Promise<string>(() => {})));
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      const abort = new AbortController();
      let settled = false;
      const promise = client.runtimeService
        .ping({}, { signal: abort.signal })
        .catch((e: Error) => e)
        .finally(() => {
          settled = true;
        });
      await vi.advanceTimersByTimeAsync(5_000);
      expect(settled).toBe(false);

      abort.abort();
      await vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
      expect(settled).toBe(true);
      const error = await promise;
      expect(error).toBeInstanceOf(ConnectError);
      expect((error as ConnectError).code).toBe(Code.Canceled);
      expect(fetchMock).not.toHaveBeenCalled();
    });

    it("releases requests waiting on a stalled refresh as soon as a JWT arrives from elsewhere", async () => {
      // e.g. the scheduled GetProject refetch completes while the refresh hangs
      const client = createClient({ jwt: "expiring-token" });
      client.setJwtRefresher(vi.fn(() => new Promise<string>(() => {})));
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      let resolved = false;
      const promise = makeRequest(client).then(() => {
        resolved = true;
      });
      await vi.advanceTimersByTimeAsync(5_000);
      expect(resolved).toBe(false);

      client.updateJwt("fresh-token");
      await vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
      await promise;

      expect(resolved).toBe(true);
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer fresh-token");
    });

    it("does not apply a user token that arrives after View As switched the context", async () => {
      const client = createClient({ jwt: "user-token", authContext: "user" });
      let resolveRefresh!: (jwt: string) => void;
      const refreshJwt = vi.fn(
        () =>
          new Promise<string>((resolve) => {
            resolveRefresh = resolve;
          }),
      );
      client.setJwtRefresher(refreshJwt);
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      const promise = makeRequest(client);
      await vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
      expect(refreshJwt).toHaveBeenCalledTimes(1);

      // View As installs a mock token while the user's refresh is in flight
      client.updateJwt("mock-token", "mock");
      resolveRefresh("fresh-user-token");
      await vi.advanceTimersByTimeAsync(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
      await promise;

      expect(client.getJwt()).toBe("mock-token");
      expect(client.authContext).toBe("mock");
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer mock-token");
    });

    it("uses the refresher the provider registered with the current context", async () => {
      const client = createClient({ jwt: "user-token", authContext: "user" });
      const refreshUserJwt = vi.fn().mockResolvedValue("fresh-user-token");
      client.setJwtRefresher(refreshUserJwt);

      // View As in the editor: the provider passes the mock token and the refresher that reissues it
      client.updateJwt("mock-token", "mock");
      const refreshMockJwt = vi.fn().mockResolvedValue("fresh-mock-token");
      client.setJwtRefresher(refreshMockJwt);
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      await makeRequestThroughRefresh(client);
      expect(refreshMockJwt).toHaveBeenCalledTimes(1);
      expect(refreshUserJwt).not.toHaveBeenCalled();
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer fresh-mock-token");

      // View As ends: the provider passes the user's token and refresher again
      client.updateJwt("user-token", "user");
      client.setJwtRefresher(refreshUserJwt);
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);
      await makeRequestThroughRefresh(client);
      expect(refreshUserJwt).toHaveBeenCalledTimes(1);
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer fresh-user-token");
    });

    it("removes the refresher when given undefined", async () => {
      const client = createClient({ jwt: "token" });
      const refreshJwt = vi.fn().mockResolvedValue("fresh-token");
      client.setJwtRefresher(refreshJwt);
      client.setJwtRefresher(undefined);
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      const promise = client.runtimeService.ping({}).catch((e: Error) => e);
      await vi.advanceTimersByTimeAsync(61_000);
      const error = await promise;

      expect(refreshJwt).not.toHaveBeenCalled();
      expect((error as Error).message).toContain(
        "Timed out waiting for a fresh JWT",
      );
    });

    it("ignores a refresher registered for another auth context", async () => {
      // View As in Rill Developer sets a mock JWT directly on the client; the provider's
      // refresher for the "user" context must not replace it with the user's token
      const client = createClient({ jwt: "user-token", authContext: "user" });
      const refreshJwt = vi.fn().mockResolvedValue("fresh-user-token");
      client.setJwtRefresher(refreshJwt);
      client.updateJwt("mock-token", "mock");
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      const promise = client.runtimeService.ping({}).catch((e: Error) => e);
      await vi.advanceTimersByTimeAsync(61_000);
      const error = await promise;

      expect(refreshJwt).not.toHaveBeenCalled();
      expect((error as Error).message).toContain(
        "Timed out waiting for a fresh JWT",
      );

      // Back in the provider's context, the refresher applies again
      client.updateJwt("user-token", "user");
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);
      await makeRequestThroughRefresh(client);

      expect(refreshJwt).toHaveBeenCalledTimes(1);
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer fresh-user-token");
    });

    it("refreshes once and retries when the runtime rejects the JWT as expired", async () => {
      // The client's clock can disagree with the runtime's; the runtime wins
      const client = createClient({ jwt: "token-runtime-rejects" });
      const refreshJwt = vi.fn().mockResolvedValue("fresh-token");
      client.setJwtRefresher(refreshJwt);
      // The retry reuses the request's Headers object, so capture the header as each call is made
      const sentAuthHeaders: (string | null)[] = [];
      fetchMock.mockImplementation((_, init) => {
        sentAuthHeaders.push(
          (init?.headers as Headers | undefined)?.get("Authorization") ?? null,
        );
        return Promise.resolve(
          sentAuthHeaders.length === 1
            ? expiredJwtResponse()
            : new Response(new Uint8Array(0), {
                status: 200,
                headers: { "content-type": "application/proto" },
              }),
        );
      });

      await makeRequestThroughRefresh(client);

      expect(refreshJwt).toHaveBeenCalledTimes(1);
      expect(sentAuthHeaders).toEqual([
        "Bearer token-runtime-rejects",
        "Bearer fresh-token",
      ]);
    });

    it("does not retry on other authentication errors", async () => {
      const client = createClient({ jwt: "bad-token" });
      const refreshJwt = vi.fn().mockResolvedValue("fresh-token");
      client.setJwtRefresher(refreshJwt);
      fetchMock.mockImplementationOnce(() =>
        Promise.resolve(
          new Response(
            JSON.stringify({ code: "unauthenticated", message: "bad token" }),
            { status: 401, headers: { "content-type": "application/json" } },
          ),
        ),
      );

      const error = await client.runtimeService.ping({}).catch((e: Error) => e);

      expect(error).toBeInstanceOf(Error);
      expect(refreshJwt).not.toHaveBeenCalled();
      expect(fetchMock).toHaveBeenCalledTimes(1);
    });

    it("does not retry without a refresher, e.g. for embeds", async () => {
      const client = createClient({ jwt: "embed-token", authContext: "embed" });
      fetchMock.mockImplementationOnce(() =>
        Promise.resolve(expiredJwtResponse()),
      );

      const error = await client.runtimeService.ping({}).catch((e: Error) => e);

      expect(error).toBeInstanceOf(Error);
      expect(fetchMock).toHaveBeenCalledTimes(1);
    });
  });

  // ── Expiry derived from the token's claims ─────────────────────────

  describe("expiry from the token's claims", () => {
    it("treats a token handed over from a cache as expired once its exp claim has passed", async () => {
      // The GetProject cache can hand a new RuntimeProvider a token issued long ago. Crediting it
      // with a full TTL from receipt made the runtime reject requests with "jwt is expired".
      const now = Date.now() / 1000;
      const client = createClient({
        jwt: makeJwt({ iat: now - 40 * 60, exp: now - 10 * 60 }),
      });
      const refreshJwt = vi.fn().mockResolvedValue("fresh-token");
      client.setJwtRefresher(refreshJwt);

      await makeRequestThroughRefresh(client);

      expect(refreshJwt).toHaveBeenCalledTimes(1);
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer fresh-token");
    });

    it("trusts an exp claim beyond the default TTL", async () => {
      const now = Date.now() / 1000;
      const oneHourToken = makeJwt({ iat: now, exp: now + 60 * 60 });
      const client = createClient({ jwt: oneHourToken });
      vi.advanceTimersByTime(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL + 60_000);

      await makeRequest(client);

      expect(fetchMock).toHaveBeenCalledTimes(1);
      expect(getLastAuthHeader(fetchMock)).toBe(`Bearer ${oneHourToken}`);
    });

    it("caps a fresh token at its lifetime from receipt when the client clock runs behind the issuer", async () => {
      // The issuer's clock is 5 minutes ahead: exp looks 35 minutes away, but the token only lives 30
      const now = Date.now() / 1000;
      const skew = 5 * 60;
      const client = createClient({
        jwt: makeJwt({ iat: now + skew, exp: now + skew + 30 * 60 }),
      });
      const refreshJwt = vi.fn().mockResolvedValue("fresh-token");
      client.setJwtRefresher(refreshJwt);
      vi.advanceTimersByTime(PAST_WARNING_WINDOW);

      await makeRequestThroughRefresh(client);

      expect(refreshJwt).toHaveBeenCalledTimes(1);
      expect(getLastAuthHeader(fetchMock)).toBe("Bearer fresh-token");
    });

    it("falls back to the default TTL for tokens without claims", async () => {
      const client = createClient({ jwt: makeJwt({ sub: "dev-user" }) });
      vi.advanceTimersByTime(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 2);

      await makeRequest(client);

      expect(fetchMock).toHaveBeenCalledTimes(1);
    });
  });
});
