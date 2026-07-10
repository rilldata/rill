import {
  SSEConnection,
  type SSEConnectionOptions,
  type SSEStartOptions,
} from "./sse-connection";
import {
  SSEConnectionLifecycle,
  type SSEConnectionLifecycleOptions,
} from "./sse-connection-lifecycle";
import {
  SSESubscriber,
  type Decoder,
  type SSESubscriberOptions,
} from "./sse-subscriber";

type DecoderMap<TMap extends Record<string, unknown>> = Partial<{
  [K in keyof TMap]: Decoder<TMap[K]>;
}>;

export interface SSEStreamLifecycleConfig {
  idleTimeouts: { short: number; normal: number };
  options?: SSEConnectionLifecycleOptions;
}

export interface CreateSSEStreamOptions<TMap extends Record<string, unknown>> {
  connection?: SSEConnectionOptions;
  decoders: DecoderMap<TMap>;
  subscriber?: SSESubscriberOptions;
  lifecycle?: SSEStreamLifecycleConfig;
}

export interface SSEStream<TMap extends Record<string, unknown>> {
  readonly status: SSEConnection["status"];
  readonly on: SSESubscriber<TMap>["on"];
  readonly once: SSESubscriber<TMap>["once"];
  readonly onConnection: SSEConnection["on"];
  readonly onceConnection: SSEConnection["once"];
  start(url: string, options?: SSEStartOptions): void;
  pause(): void;
  resumeIfPaused(): Promise<void>;
  close(cleanup?: boolean): void;
  cleanup(): void;
}

/**
 * Convenience composition for consumers that want one object without
 * collapsing transport/retry concerns into decoding concerns.
 */
export function createSSEStream<TMap extends Record<string, unknown>>(
  options: CreateSSEStreamOptions<TMap>,
): SSEStream<TMap> {
  const connection = new SSEConnection(options.connection);
  const subscriber = new SSESubscriber<TMap>(
    connection,
    options.decoders,
    options.subscriber,
  );
  const lifecycle = options.lifecycle
    ? new SSEConnectionLifecycle(
        connection,
        options.lifecycle.idleTimeouts,
        options.lifecycle.options,
      )
    : undefined;

  const closeStream = (cleanup = false) => {
    lifecycle?.stop();
    connection.close(cleanup);
    if (cleanup) {
      subscriber.cleanup();
    }
  };

  return {
    status: connection.status,
    on: subscriber.on,
    once: subscriber.once,
    onConnection: connection.on,
    onceConnection: connection.once,
    start(url, startOptions = {}) {
      connection.start(url, startOptions);
      lifecycle?.start();
    },
    pause() {
      connection.pause();
    },
    resumeIfPaused() {
      return connection.resumeIfPaused();
    },
    close: closeStream,
    cleanup() {
      closeStream(true);
    },
  };
}
