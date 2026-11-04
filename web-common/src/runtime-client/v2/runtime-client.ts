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
  CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL,
} from "../constants";
import { RequestQueue, createQueueInterceptor } from "./request-queue";

export type AuthContext = "user" | "mock" | "magic" | "embed";

/**
 * Fetches a fresh JWT from whoever issued the current one, e.g. by refetching GetProject in Rill Cloud.
 * Resolves with the new token; rejects when the issuer could not be reached or refused.
 */
export type JwtRefresher = () => Promise<string | undefined>;

// How long a request waits for a fresh JWT before it fails.
const JWT_REFRESH_TIMEOUT = 60_000;
// Delay before retrying a failed refresh: doubles from the base up to the max while the issuer keeps failing.
const JWT_REFRESH_BACKOFF_BASE = 1_000;
const JWT_REFRESH_BACKOFF_MAX = 10_000;

export class RuntimeClient {
  readonly host: string;
  readonly instanceId: string;
  readonly transport: Transport;
  readonly requestQueue: RequestQueue;

  // JWT state (mutable; read by the transport interceptor)
  private currentJwt: string | undefined;
  private jwtExpiresAt = 0;
  private _authContext: AuthContext;
  // Refresher registered by RuntimeProvider, bound to the auth context it was registered under
  private refreshJwt: JwtRefresher | undefined;
  private refreshJwtAuthContext: AuthContext | undefined;
  // In-flight refresh shared by every request waiting on it, and the backoff after failures
  private jwtRefresh: Promise<void> | undefined;
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

    if (jwt !== this.currentJwt) {
      this.currentJwt = jwt;
      this.jwtExpiresAt = jwt ? jwtExpiry(jwt, Date.now()) : 0;
    }
    if (authContext) this._authContext = authContext;

    return authContextChanged;
  }

  /**
   * Registers the callback that fetches a fresh JWT when the current one is about to expire.
   * It stays bound to the auth context it was registered under: a mock JWT that View As sets directly
   * on the client must not be replaced by the provider's token, so the refresher is ignored until the
   * client is back in that context.
   */
  setJwtRefresher(refreshJwt: JwtRefresher | undefined): void {
    this.refreshJwt = refreshJwt;
    this.refreshJwtAuthContext = this._authContext;
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
   * Blocks until the JWT is fresh. With a refresher registered it asks the issuer for a new token
   * right away, retrying with backoff while the issuer is unreachable; without one it waits for an
   * updateJwt() call. Gives up after JWT_REFRESH_TIMEOUT, or at once when the issuer refuses because
   * the session or the project access is gone.
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
      if (Date.now() > deadline) {
        throw new Error(
          "Timed out waiting for a fresh JWT" +
            (lastError ? `: ${extractErrorMessage(lastError)}` : ""),
        );
      }
      if (!this.canRefreshJwt()) {
        // Wait for whoever owns the token (e.g. View As) to call updateJwt()
        await sleep(CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL);
        continue;
      }
      try {
        await this.runJwtRefresh();
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
    return (
      !!this.refreshJwt && this.refreshJwtAuthContext === this._authContext
    );
  }

  /** Runs one refresh, shared by every caller that joins while it is in flight. */
  private runJwtRefresh(): Promise<void> {
    this.jwtRefresh ??= this.refreshJwtOnce().finally(() => {
      this.jwtRefresh = undefined;
    });
    return this.jwtRefresh;
  }

  private async refreshJwtOnce(): Promise<void> {
    const delay = this.jwtRefreshNotBefore - Date.now();
    if (delay > 0) await sleep(delay);
    const refreshJwt = this.refreshJwt;
    const authContext = this._authContext;
    if (this.disposed || !refreshJwt) return;
    try {
      const jwt = await refreshJwt();
      // The provider may have switched auth context (View As) meanwhile; its token wins
      if (
        this.disposed ||
        this.refreshJwt !== refreshJwt ||
        this._authContext !== authContext
      ) {
        return;
      }
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
    }
  }

  dispose(): void {
    this.disposed = true;
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

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
