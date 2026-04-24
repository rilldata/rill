import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import {
  type V1WatchFilesResponse,
  type V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client";
import {
  createFileInvalidatorState,
  handleFileEvent,
  type FileInvalidatorState,
} from "@rilldata/web-common/runtime-client/invalidation/file-invalidators";
import {
  createResourceInvalidatorState,
  handleResourceEvent,
  type ResourceInvalidatorState,
} from "@rilldata/web-common/runtime-client/invalidation/resource-invalidators";
import {
  ConnectionStatus,
  createSSEStream,
  type SSEStream,
} from "@rilldata/web-common/runtime-client/sse";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { QueryClient } from "@tanstack/svelte-query";

const MAX_RETRIES = 3;

/**
 * Idle-timeout presets.
 *
 * - `aggressive`: pause the stream quickly when the tab idles. Used by Rill
 *   Developer where the browser's 6-connection per-host limit bites because
 *   SSE, queries, and dev assets all share `localhost:<port>`.
 * - `none`: don't attach a lifecycle at all. Used by the cloud editor and
 *   other consumers that need a persistent connection.
 */
const LIFECYCLE_PRESETS = {
  aggressive: { short: 20_000, normal: 120_000 },
  none: undefined,
} as const satisfies Record<
  string,
  { short: number; normal: number } | undefined
>;

export type LifecyclePreset = keyof typeof LIFECYCLE_PRESETS;

/** Server-sent error payload. The unified /sse endpoint emits an
 * `event: error` frame on gRPC errors; the payload shape is JSON-decoded
 * best-effort. */
interface V1WatchErrorPayload {
  code?: string;
  message?: string;
}

type WatcherEventMap = {
  file: V1WatchFilesResponse;
  resource: V1WatchResourcesResponse;
  error: V1WatchErrorPayload;
};

export interface FileAndResourceWatcherOptions {
  runtimeClient: RuntimeClient;
  queryClient: QueryClient;
  /** Lifecycle preset. "none" skips attaching an SSEConnectionLifecycle entirely. */
  lifecycle: LifecyclePreset;
  /** Hook fired before each reconnect attempt. Cloud editor passes a JWT
   * refresh here; local Rill Developer does not. */
  onBeforeReconnect?: () => Promise<void>;
}

/**
 * Thin watcher that wires SSE transport → typed subscriber → pure invalidators.
 *
 * One instance per mount: Rill Cloud's editor switches between projects and
 * branches, each backed by a distinct runtime, so the old singleton no
 * longer matches the semantics of the frontend.
 */
export class FileAndResourceWatcher {
  public readonly status: SSEStream<WatcherEventMap>["status"];

  private readonly stream: SSEStream<WatcherEventMap>;

  private readonly instanceId: string;
  private readonly runtimeClient: RuntimeClient;
  private readonly queryClient: QueryClient;

  private readonly fileState: FileInvalidatorState =
    createFileInvalidatorState();
  private readonly resourceState: ResourceInvalidatorState =
    createResourceInvalidatorState();

  private currentUrl: string | undefined;

  constructor(options: FileAndResourceWatcherOptions) {
    this.runtimeClient = options.runtimeClient;
    this.instanceId = options.runtimeClient.instanceId;
    this.queryClient = options.queryClient;

    const preset = LIFECYCLE_PRESETS[options.lifecycle];
    this.stream = createSSEStream<WatcherEventMap>({
      connection: {
        maxRetryAttempts: MAX_RETRIES,
        retryOnError: true,
        retryOnClose: true,
        onBeforeReconnect: options.onBeforeReconnect,
      },
      decoders: {
        file: (data) => JSON.parse(data) as V1WatchFilesResponse,
        resource: (data) => JSON.parse(data) as V1WatchResourcesResponse,
        error: (data) => {
          try {
            return JSON.parse(data) as V1WatchErrorPayload;
          } catch {
            return { message: data };
          }
        },
      },
      lifecycle: preset ? { idleTimeouts: preset } : undefined,
    });
    this.status = this.stream.status;

    this.stream.on("file", (event) => {
      void handleFileEvent(
        event,
        this.queryClient,
        this.runtimeClient,
        this.fileState,
      );
    });

    this.stream.on("resource", (event) => {
      // web-local e2e tests parse this log line to wait for specific resources
      // to finish reconciling.
      if (import.meta.env.VITE_PLAYWRIGHT_TEST) {
        console.log(
          `[${event.resource?.meta?.reconcileStatus}] ${event.name?.kind}/${event.name?.name}`,
        );
      }
      void handleResourceEvent(
        event,
        this.queryClient,
        this.runtimeClient,
        this.resourceState,
      );
    });

    this.stream.on("error", (payload) => {
      console.warn("SSE watch error:", payload);
    });

    // On reconnect, re-run the post-connect bootstrap: events emitted while
    // disconnected may have been dropped, so we force a full re-fetch of
    // runtime-scoped queries and refresh the file-artifacts index.
    this.stream.onConnection("reconnect", () => {
      void this.invalidateAll().then(() =>
        fileArtifacts.init(this.runtimeClient, this.queryClient),
      );
    });
  }

  /**
   * Begin watching the given URL. Calling `start` with the same URL is a
   * no-op, so the component can re-run on reactive updates without
   * thrashing the transport.
   */
  public start(url: string): void {
    if (url === this.currentUrl) return;
    this.currentUrl = url;
    this.stream.start(url, {
      getJwt: () => this.runtimeClient.getJwt(),
    });
  }

  public close(cleanup = false): void {
    this.currentUrl = undefined;
    this.stream.close(cleanup);
  }

  private invalidateAll() {
    return this.queryClient.invalidateQueries({
      predicate: (query) => {
        const key = query.queryKey;
        if (key.length >= 3 && key[2] === this.instanceId) {
          const svc = key[0];
          return (
            svc === "QueryService" ||
            svc === "RuntimeService" ||
            svc === "ConnectorService"
          );
        }
        return false;
      },
    });
  }
}

export { ConnectionStatus };
