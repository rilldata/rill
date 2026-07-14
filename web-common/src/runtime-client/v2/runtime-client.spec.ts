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

/** Extract the Authorization header from the most recent fetch call. */
function getLastAuthHeader(fetchMock: ReturnType<typeof vi.fn>): string | null {
  if (fetchMock.mock.calls.length === 0) return null;
  const lastCall = fetchMock.mock.calls[fetchMock.mock.calls.length - 1];
  const headers: Headers | undefined = lastCall[1]?.headers;
  return headers?.get?.("Authorization") ?? null;
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
  });
});
