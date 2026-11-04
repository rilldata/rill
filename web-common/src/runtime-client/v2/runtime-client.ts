import {
  Code,
  ConnectError,
  createClient,
  type Client,
  type Transport,
} from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { QueryService } from "../../proto/gen/rill/runtime/v1/queries_connect";
import { RuntimeService } from "../../proto/gen/rill/runtime/v1/api_connect";
import { ConnectorService } from "../../proto/gen/rill/runtime/v1/connectors_connect";
import { extractErrorMessage, extractErrorStatusCode } from "../../lib/errors";
import {
  RUNTIME_ACCESS_TOKEN_DEFAULT_TTL,
  JWT_EXPIRY_WARNING_WINDOW,
} from "../constants";
import { RequestQueue, createQueueInterceptor } from "./request-queue";

export type AuthContext = "user" | "mock" | "magic" | "embed";

/**
 * Fetches a fresh JWT from whoever issued the current one, e.g. by refetching GetProject in Rill Cloud.
 * Resolves with the new token; rejects when the issuer could not be reached or refused.
 * The signal aborts when the client gives up on the call, so the issuer request can be cancelled.
 */
export type JwtRefresher = (signal: AbortSignal) => Promise<string | undefined>;

// How long a request waits for a fresh JWT before it fails.
const JWT_REFRESH_TIMEOUT = 60_000;
// How long one call to the refresher may take before it is abandoned and retried.
const JWT_REFRESH_ATTEMPT_TIMEOUT = 20_000;
// Delay before retrying a failed refresh: doubles from the base up to the max while the issuer keeps failing.
const JWT_REFRESH_BACKOFF_BASE = 1_000;
const JWT_REFRESH_BACKOFF_MAX = 10_000;

const TimedOut = Symbol("timed out");

export class RuntimeClient {
  readonly host: string;
  readonly instanceId: string;
  readonly transport: Transport;
  readonly requestQueue: RequestQueue;

  // JWT state (mutable; read by the transport interceptor)
  private currentJwt: string | undefined;
  private jwtExpiresAt = 0;
  private _authContext: AuthContext;
  // Resolved and replaced whenever the JWT, the auth context or a refresher changes, or the client
  // is disposed, so that requests waiting for a fresh JWT re-check their situation right away
  private jwtChanged = createDeferred();
  // Refreshers per auth context: RuntimeProvider registers one for its context and View As one for "mock"
  private readonly jwtRefreshers = new Map<AuthContext, JwtRefresher>();
  // In-flight refresh shared by every request waiting on it, and the backoff after failures
  private jwtRefresh: Promise<void> | undefined;
  private jwtRefreshAbort: AbortController | undefined;
  private jwtRefreshBackoff = 0;
  private jwtRefreshNotBefore = 0;
  private disposed = false;

  // Cached service clients (created once per RuntimeClient)
  private _queryService: Client<typeof QueryService> | null = null;
  private _runtimeService: Client<typeof RuntimeService> | null = null;
  private _connectorService: Client<typeof ConnectorService> | null = null;

  constructor(opts: {
    host: string;
    instanceId: string;
    jwt?: string;
    authContext?: AuthContext;
  }) {
    if (typeof opts.host !== "string") {
      throw new Error(
        "RuntimeClient requires host to be a string. " +
          "An empty string is valid (same-origin for local dev).",
      );
    }
    if (!opts.instanceId) {
      throw new Error(
        "RuntimeClient requires a non-empty instanceId. " +
          "The caller should not mount RuntimeProvider until instanceId is available.",
      );
    }
    this.host = opts.host;
    this.instanceId = opts.instanceId;
    this.currentJwt = opts.jwt;
    this.jwtExpiresAt = opts.jwt ? jwtExpiry(opts.jwt, Date.now()) : 0;
    this._authContext = opts.authContext ?? "user";
    this.requestQueue = new RequestQueue();

    this.transport = createConnectTransport({
      baseUrl: opts.host,
      interceptors: [
        // Queue controls when requests fire (outermost: wraps everything)
        createQueueInterceptor(this.requestQueue),
        // JWT interceptor adds auth header just before the network call
        (next) => async (req) => {
          await this.waitForFreshJwt();
          const jwt = this.currentJwt;
          if (jwt) req.header.set("Authorization", `Bearer ${jwt}`);
          try {
            return await next(req);
          } catch (err) {
            // The runtime's clock is the source of truth: when it rejects a token this client still
            // considered valid, refresh once and retry. The runtime rejects expired tokens before the
            // handler runs, so the retry is safe; only unary requests can be re-sent as is.
            if (
              !jwt ||
              req.stream ||
              !this.canRefreshJwt() ||
              !isExpiredJwtError(err)
            ) {
              throw err;
            }
            if (jwt === this.currentJwt) this.jwtExpiresAt = 0;
            await this.waitForFreshJwt();
            if (this.currentJwt) {
              req.header.set("Authorization", `Bearer ${this.currentJwt}`);
            } else {
              req.header.delete("Authorization");
            }
            return next(req);
          }
        },
      ],
    });
  }

  /**
   * Called by RuntimeProvider when the parent passes a new JWT prop.
   * Returns true if the auth context changed (caller should invalidate queries).
   */
  updateJwt(jwt: string | undefined, authContext?: AuthContext): boolean {
    const authContextChanged =
      !!this._authContext && !!authContext && authContext !== this._authContext;

    const jwtChanged = jwt !== this.currentJwt;
    if (jwtChanged) {
      this.currentJwt = jwt;
      this.jwtExpiresAt = jwt ? jwtExpiry(jwt, Date.now()) : 0;
    }
    if (authContext) this._authContext = authContext;
    if (jwtChanged || authContextChanged) this.wakeWaitingRequests();

    return authContextChanged;
  }

  /**
   * Registers the callback that fetches a fresh JWT for the current auth context, replacing any earlier
   * one for that context; undefined removes it. Refreshers are kept per context: RuntimeProvider's stays
   * registered for its context while View As installs a mock token with its own refresher for "mock",
   * and applies again once View As ends and the context switches back.
   */
  setJwtRefresher(refreshJwt: JwtRefresher | undefined): void {
    const authContext = this._authContext;
    if (refreshJwt === this.jwtRefreshers.get(authContext)) return;
    if (refreshJwt) {
      this.jwtRefreshers.set(authContext, refreshJwt);
    } else {
      this.jwtRefreshers.delete(authContext);
    }
    // A new issuer starts without the backoff the previous one earned
    this.jwtRefreshBackoff = 0;
    this.jwtRefreshNotBefore = 0;
    this.wakeWaitingRequests();
  }

  /**
   * Refreshes the JWT if it is expired or about to expire, e.g. when the tab becomes visible again
   * after a laptop sleep skipped the scheduled refresh. Failures are left to the next request,
   * which retries with backoff and surfaces the error.
   */
  async refreshJwtIfStale(): Promise<void> {
    if (this.isJwtFresh() || !this.canRefreshJwt()) return;
    try {
      await this.runJwtRefresh();
    } catch {
      // Reported by the next request that needs the token
    }
  }

  /** Returns the current JWT (used by SSE clients and other non-query consumers). */
  getJwt(): string | undefined {
    return this.currentJwt;
  }

  /** Returns how the client authenticates with the runtime, e.g. "magic" on public URLs. */
  get authContext(): AuthContext {
    return this._authContext;
  }

  get queryService() {
    return (this._queryService ??= createClient(QueryService, this.transport));
  }

  get runtimeService() {
    return (this._runtimeService ??= createClient(
      RuntimeService,
      this.transport,
    ));
  }

  get connectorService() {
    return (this._connectorService ??= createClient(
      ConnectorService,
      this.transport,
    ));
  }

  /**
   * Blocks until the JWT is fresh. With a refresher registered for the current auth context it asks
   * the issuer for a new token right away, retrying with backoff while the issuer is unreachable;
   * without one it waits for an updateJwt() call. Gives up after JWT_REFRESH_TIMEOUT, or at once
   * when the issuer refuses because the session or the project access is gone.
   */
  private async waitForFreshJwt(): Promise<void> {
    // Embeds have a 24h backend-issued TTL and no refresher; skip client-side expiry check
    if (this._authContext === "embed") return;

    const deadline = Date.now() + JWT_REFRESH_TIMEOUT;
    let lastError: unknown;
    while (!this.isJwtFresh()) {
      if (this.disposed) {
        throw new DOMException("Request cancelled", "AbortError");
      }
      const remaining = deadline - Date.now();
      if (remaining <= 0) {
        throw new Error(
          "Timed out waiting for a fresh JWT" +
            (lastError ? `: ${extractErrorMessage(lastError)}` : ""),
        );
      }
      // Whatever happens below, a change to the JWT or the client wakes us to re-check
      const changed = this.jwtChanged.promise;
      if (!this.canRefreshJwt()) {
        // Wait for whoever owns the token (e.g. View As in Rill Developer) to call updateJwt()
        await waitAtMost(changed, remaining);
        continue;
      }
      try {
        await waitAtMost(
          Promise.race([this.runJwtRefresh(), changed]),
          remaining,
        );
      } catch (err) {
        const status = extractErrorStatusCode(err);
        if (status === 401 || status === 403) {
          throw new Error(
            `Failed to refresh the runtime JWT: ${extractErrorMessage(err)}`,
          );
        }
        lastError = err;
      }
    }
  }

  private isJwtFresh(): boolean {
    return (
      !this.currentJwt ||
      Date.now() + JWT_EXPIRY_WARNING_WINDOW <= this.jwtExpiresAt
    );
  }

  private canRefreshJwt(): boolean {
    return this.jwtRefreshers.has(this._authContext);
  }

  private wakeWaitingRequests(): void {
    const { resolve } = this.jwtChanged;
    this.jwtChanged = createDeferred();
    resolve();
  }

  /** Runs one refresh, shared by every caller that joins while it is in flight. */
  private runJwtRefresh(): Promise<void> {
    this.jwtRefresh ??= this.refreshJwtOnce().finally(() => {
      this.jwtRefresh = undefined;
    });
    return this.jwtRefresh;
  }

  private async refreshJwtOnce(): Promise<void> {
    // Pin the refresher and the context it serves: both can change while this attempt waits or runs
    // (e.g. View As switching to a mock token), and a token must only be applied to the context it
    // was issued for
    const authContext = this._authContext;
    const refreshJwt = this.jwtRefreshers.get(authContext);
    if (!refreshJwt) return;
    const delay = this.jwtRefreshNotBefore - Date.now();
    if (delay > 0) await sleep(delay);
    if (!this.servesCurrentContext(refreshJwt, authContext)) return;

    // A call that stalls is abandoned and cancelled so the retry is not stuck behind it
    const abort = new AbortController();
    this.jwtRefreshAbort = abort;
    try {
      const jwt = await waitAtMost(
        refreshJwt(abort.signal),
        JWT_REFRESH_ATTEMPT_TIMEOUT,
      );
      if (jwt === TimedOut) {
        abort.abort();
        throw new Error(
          `the issuer did not respond within ${JWT_REFRESH_ATTEMPT_TIMEOUT / 1000}s`,
        );
      }
      if (!this.servesCurrentContext(refreshJwt, authContext)) return;
      if (!jwt) throw new Error("the issuer returned no JWT");
      this.updateJwt(jwt);
      if (!this.isJwtFresh()) {
        throw new Error("the issuer returned an expired JWT");
      }
      this.jwtRefreshBackoff = 0;
      this.jwtRefreshNotBefore = 0;
    } catch (err) {
      this.jwtRefreshBackoff = Math.min(
        Math.max(this.jwtRefreshBackoff * 2, JWT_REFRESH_BACKOFF_BASE),
        JWT_REFRESH_BACKOFF_MAX,
      );
      this.jwtRefreshNotBefore = Date.now() + this.jwtRefreshBackoff;
      throw err;
    } finally {
      if (this.jwtRefreshAbort === abort) this.jwtRefreshAbort = undefined;
    }
  }

  /** Whether the refresher an attempt pinned is still the one registered for the client's current context. */
  private servesCurrentContext(
    refreshJwt: JwtRefresher,
    authContext: AuthContext,
  ): boolean {
    return (
      !this.disposed &&
      this._authContext === authContext &&
      this.jwtRefreshers.get(authContext) === refreshJwt
    );
  }

  dispose(): void {
    this.disposed = true;
    this.jwtRefreshAbort?.abort();
    this.wakeWaitingRequests();
    this.requestQueue.clear();
  }
}

/**
 * Matches the runtime's rejection of an expired token. It is raised by the auth interceptor
 * before the handler runs, so a request that gets it can be retried with a fresh token.
 */
function isExpiredJwtError(err: unknown): boolean {
  return (
    err instanceof ConnectError &&
    err.code === Code.Unauthenticated &&
    err.rawMessage.includes("jwt is expired")
  );
}

/**
 * Returns when the JWT expires, in milliseconds since the epoch.
 * The exp claim is the source of truth, so a token handed over from a cache is not credited with a
 * full TTL again. A fresh token is also capped at receivedAt plus its own lifetime (exp minus iat),
 * which keeps a client clock that runs behind the issuer's from presenting tokens the runtime
 * already rejects. Tokens without readable claims fall back to the default TTL from receipt.
 */
function jwtExpiry(jwt: string, receivedAt: number): number {
  const claims = decodeJwtClaims(jwt);
  if (typeof claims?.exp !== "number") {
    return receivedAt + RUNTIME_ACCESS_TOKEN_DEFAULT_TTL;
  }
  const expiresAt = claims.exp * 1000;
  if (typeof claims.iat !== "number") return expiresAt;
  return Math.min(expiresAt, receivedAt + (claims.exp - claims.iat) * 1000);
}

/** Decodes the payload of a JWT without verifying it. Returns undefined for anything that is not a JWT. */
function decodeJwtClaims(jwt: string): Record<string, unknown> | undefined {
  const parts = jwt.split(".");
  if (parts.length !== 3) return undefined;
  try {
    const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const bytes = Uint8Array.from(atob(base64), (c) => c.charCodeAt(0));
    const claims: unknown = JSON.parse(new TextDecoder().decode(bytes));
    return typeof claims === "object" && claims !== null
      ? (claims as Record<string, unknown>)
      : undefined;
  } catch {
    return undefined;
  }
}

/** Settles like the promise, or resolves with TimedOut once `ms` elapse first; the timer that lost is cleared. */
function waitAtMost<T>(
  promise: Promise<T>,
  ms: number,
): Promise<T | typeof TimedOut> {
  let timer: ReturnType<typeof setTimeout> | undefined;
  const timeout = new Promise<typeof TimedOut>((resolve) => {
    timer = setTimeout(() => resolve(TimedOut), ms);
  });
  return Promise.race([promise, timeout]).finally(() => clearTimeout(timer));
}

function createDeferred(): { promise: Promise<void>; resolve: () => void } {
  let resolve!: () => void;
  const promise = new Promise<void>((r) => {
    resolve = r;
  });
  return { promise, resolve };
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
