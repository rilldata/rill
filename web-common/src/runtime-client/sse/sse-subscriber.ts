import { createEventBinding } from "@rilldata/web-common/lib/event-emitter.ts";
import type { SSEConnection } from "./sse-connection";
import type { SSEMessage } from "./sse-protocol";

export type Decoder<T> = (data: string) => T;

export interface SSESubscriberOptions {
  /** Called for messages whose type has no registered decoder. */
  onUnknown?: (message: SSEMessage) => void;
  /** Called when a decoder throws. The typed event is not emitted. */
  onParseError?: (err: unknown, message: SSEMessage) => void;
}

/**
 * Typed routing layer over SSEConnection.
 *
 * Consumers register one decoder per event type. For each incoming message,
 * SSESubscriber runs the matching decoder and emits the decoded payload
 * through a typed event API.
 *
 * If a decoder throws, the subscriber calls `onParseError` and skips emit.
 *
 * Per the SSE spec, frames without an `event:` line default to `"message"`.
 * This class normalizes `message.type === undefined` to `"message"` before
 * decoder lookup. Consumers that stream untagged frames can register a
 * `"message"` decoder to handle them. If no `"message"` decoder is
 * registered, untagged frames fall through to `onUnknown`.
 */
export class SSESubscriber<TMap extends Record<string, unknown>> {
  private readonly events = createEventBinding<TMap>();
  public readonly on = this.events.on;
  public readonly once = this.events.once;

  private readonly unsubscribeFromConnection: () => void;

  constructor(
    connection: SSEConnection,
    private readonly decoders: Partial<{
      [K in keyof TMap]: Decoder<TMap[K]>;
    }>,
    private readonly options: SSESubscriberOptions = {},
  ) {
    this.unsubscribeFromConnection = connection.on(
      "message",
      this.handleMessage,
    );
  }

  /**
   * Detach the message listener from the underlying connection. Call when
   * the subscriber is no longer needed to prevent leaks.
   */
  public cleanup(): void {
    this.unsubscribeFromConnection();
    this.events.clearListeners();
  }

  private handleMessage = (message: SSEMessage) => {
    const eventType = (message.type ?? "message") as keyof TMap;
    const decoder = this.decoders[eventType];
    if (!decoder) {
      this.options.onUnknown?.(message);
      return;
    }
    try {
      this.emitDecoded(eventType, decoder(message.data));
    } catch (err) {
      this.options.onParseError?.(err, message);
    }
  };

  // Centralize the emit cast so message handling stays linear/readable.
  private emitDecoded(type: keyof TMap, payload: unknown): void {
    // Cast narrows the variadic emit signature; the TMap[K] payload is
    // type-safe at the call site that invokes this helper.
    (this.events.emit as (eventType: keyof TMap, payload: unknown) => void)(
      type,
      payload,
    );
  }
}
