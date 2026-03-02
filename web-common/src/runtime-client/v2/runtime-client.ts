import { createClient, type Client, type Transport } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { QueryService } from "../../proto/gen/rill/runtime/v1/queries_connect";
import { RuntimeService } from "../../proto/gen/rill/runtime/v1/api_connect";
import { ConnectorService } from "../../proto/gen/rill/runtime/v1/connectors_connect";
import {
  RUNTIME_ACCESS_TOKEN_DEFAULT_TTL,
  JWT_EXPIRY_WARNING_WINDOW,
  CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL,
} from "../constants";
import { RequestQueue, createQueueInterceptor } from "./request-queue";

export type AuthContext = "user" | "mock" | "magic" | "embed";

export class RuntimeClient {
  readonly host: string;
  readonly instanceId: string;
  readonly transport: Transport;
  readonly requestQueue: RequestQueue;

  // JWT state (mutable; read by the transport interceptor)
  private currentJwt: string | undefined;
  private jwtReceivedAt: number;
  private authContext: AuthContext;
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
    this.jwtReceivedAt = opts.jwt ? Date.now() : 0;
    this.authContext = opts.authContext ?? "user";
    this.requestQueue = new RequestQueue();

    this.transport = createConnectTransport({
      baseUrl: opts.host,
      interceptors: [
        // Queue controls when requests fire (outermost: wraps everything)
        createQueueInterceptor(this.requestQueue),
        // JWT interceptor adds auth header just before the network call
        (next) => async (req) => {
          if (this.currentJwt) {
            await this.waitForFreshJwt();
            req.header.set("Authorization", `Bearer ${this.currentJwt}`);
          }
          return next(req);
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
      !!this.authContext && !!authContext && authContext !== this.authContext;

    if (jwt !== this.currentJwt) {
      this.currentJwt = jwt;
      this.jwtReceivedAt = Date.now();
    }
    if (authContext) this.authContext = authContext;

    return authContextChanged;
  }

  /** Returns the current JWT (used by SSE clients and other non-query consumers). */
  getJwt(): string | undefined {
    return this.currentJwt;
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
   * If the JWT is close to expiring, wait in a polling loop for
   * RuntimeProvider to call updateJwt() with a fresh token.
   * Ported from http-client.ts:maybeWaitForFreshJWT.
   */
  private async waitForFreshJwt(): Promise<void> {
    // Embeds have a 24h backend-issued TTL; skip client-side expiry check
    if (this.authContext === "embed") return;

    const deadline = Date.now() + 60_000;
    let expiresAt = this.jwtReceivedAt + RUNTIME_ACCESS_TOKEN_DEFAULT_TTL;
    while (Date.now() + JWT_EXPIRY_WARNING_WINDOW > expiresAt) {
      if (Date.now() > deadline || this.disposed) {
        throw new Error("Timed out waiting for a fresh JWT");
      }
      await new Promise((r) =>
        setTimeout(r, CHECK_RUNTIME_STORE_FOR_JWT_INTERVAL),
      );
      // Re-check: provider may have called updateJwt() while we waited
      expiresAt = this.jwtReceivedAt + RUNTIME_ACCESS_TOKEN_DEFAULT_TTL;
    }
  }

  dispose(): void {
    this.disposed = true;
    this.requestQueue.clear();
  }
}
