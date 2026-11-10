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
import { extractErrorMessage } from "../../lib/errors";
import { waitUntil } from "../../lib/waitUtils";
import {
  RUNTIME_ACCESS_TOKEN_DEFAULT_TTL,
  JWT_EXPIRY_WARNING_WINDOW,
  CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL,
} from "../constants";
import { createRequestAttributionInterceptor } from "./request-attribution";
import { RequestQueue, createQueueInterceptor } from "./request-queue";

export type AuthContext = "user" | "mock" | "magic" | "embed";

/**
 * Fetches a fresh JWT from whoever issued the current one, e.g. by refetching GetProject in Rill Cloud.
 * Resolves with the new token; rejects when the issuer could not be reached or refused.
 * Retrying transient failures and sharing one request between concurrent callers is up to the refresher;
 * in Rill Cloud the query client does both. The signal aborts when the client gives up waiting,
 * so a stalled request can be cancelled instead of being joined by the next caller.
 */
export type JwtRefresher = (signal: AbortSignal) => Promise<string | undefined>;

// How long a request waits for a fresh JWT before it fails.
const JWT_REFRESH_TIMEOUT = 60_000;

export class RuntimeClient {
  readonly host: string;
  readonly instanceId: string;
  readonly transport: Transport;
  readonly requestQueue: RequestQueue;

  // JWT state (mutable; read by the transport interceptor)
  private currentJwt: string | undefined;
  private jwtExpiresAt = 0;
  private _authContext: AuthContext;
  // The refresher RuntimeProvider registered, and the auth context it issues tokens for
  private jwtRefresher:
    | { refreshJwt: JwtRefresher; authContext: AuthContext }
    | undefined;
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
        // Names the RPC on every error or abort reason that leaves it, so error telemetry can attribute
        // unhandled rejections to a request (outermost: it must see the signal before the queue rejects)
        createRequestAttributionInterceptor(),
        // Queue controls when requests fire
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
   * Registers the callback that fetches a fresh JWT for the current auth context; undefined removes it.
   * RuntimeProvider calls it right after updateJwt() with the refresher matching its JWT. The refresher
   * only applies while the client stays in that context: when View As in Rill Developer sets a mock JWT
   * directly on the client, requests wait for the next updateJwt() instead of fetching the user's token.
   */
  setJwtRefresher(refreshJwt: JwtRefresher | undefined): void {
    this.jwtRefresher = refreshJwt
      ? { refreshJwt, authContext: this._authContext }
      : undefined;
  }

  /** Returns the current JWT (used by SSE clients and other non-query consumers). */
  getJwt(): string | undefined {
    return this.currentJwt;
  }

  /**
   * Returns the JWT once it is not about to expire, waiting for a refresh like
   * queries do. For one-off requests sent outside the query client.
   */
  async getFreshJwt(): Promise<string | undefined> {
    if (this.currentJwt) await this.waitForFreshJwt();
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
   * Blocks until the JWT is fresh. With a refresher registered for the current auth context it asks the issuer
   * for a new token right away; without one it waits for an updateJwt() call. Either way a token that arrives
   * through updateJwt() in the meantime, e.g. from the scheduled GetProject refetch, releases the request.
   * Gives up after JWT_REFRESH_TIMEOUT, or as soon as the refresher fails.
   */
  private async waitForFreshJwt(): Promise<void> {
    // Embeds have a 24h backend-issued TTL and no refresher; skip client-side expiry check
    if (this._authContext === "embed" || this.isJwtFresh()) return;

    const refresher = this.jwtRefresher;
    const abort = new AbortController();
    let refreshError: unknown;
    if (refresher?.authContext === this._authContext) {
      this.refreshJwt(refresher, abort.signal).catch((err) => {
        refreshError = err ?? new Error("unknown error");
      });
    }
    await waitUntil(
      () => this.isJwtFresh() || this.disposed || refreshError !== undefined,
      JWT_REFRESH_TIMEOUT,
      CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL,
    );
    if (this.isJwtFresh()) return;

    // Cancel a refresh that is still running, so the next request does not join a stalled one
    abort.abort();
    if (this.disposed) {
      throw new DOMException(
        "Request cancelled: runtime client disposed while waiting for a fresh JWT",
        "AbortError",
      );
    }
    if (refreshError !== undefined) {
      throw new Error(
        `Failed to refresh the runtime JWT: ${extractErrorMessage(refreshError)}`,
      );
    }
    throw new Error("Timed out waiting for a fresh JWT");
  }

  /**
   * Fetches a JWT with the refresher and applies it, unless the client moved on while the request was in flight,
   * e.g. View As switched to a mock token: a token must only be applied to the context it was issued for.
   */
  private async refreshJwt(
    refresher: { refreshJwt: JwtRefresher; authContext: AuthContext },
    signal: AbortSignal,
  ): Promise<void> {
    const jwt = await refresher.refreshJwt(signal);
    if (!jwt) throw new Error("the issuer returned no JWT");
    if (
      this.disposed ||
      this._authContext !== refresher.authContext ||
      this.jwtRefresher?.refreshJwt !== refresher.refreshJwt
    ) {
      return;
    }
    this.updateJwt(jwt);
    if (!this.isJwtFresh()) {
      throw new Error("the issuer returned an expired JWT");
    }
  }

  private isJwtFresh(): boolean {
    return (
      !this.currentJwt ||
      Date.now() + JWT_EXPIRY_WARNING_WINDOW <= this.jwtExpiresAt
    );
  }

  private canRefreshJwt(): boolean {
    return this.jwtRefresher?.authContext === this._authContext;
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
