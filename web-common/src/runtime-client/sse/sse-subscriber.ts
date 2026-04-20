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
 * Typed routing layer over SSEConnection. Consumers register one decoder
 * per event type; SSESubscriber runs it, catches parse errors, and emits
 * the decoded payload through a typed event API.
 *
 * Per the SSE spec, frames without an `event:` line default to type
 * "message". The subscriber normalizes `msg.type === undefined` to
 * "message" before decoder lookup, so a consumer streaming untagged
 * frames (e.g. the admin AI endpoint emits success payloads untagged and
 * failures as `event: error`) can register a "message" decoder and route
 * them cleanly. If no "message" decoder is registered, untagged frames
 * fall through to `onUnknown`.
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
    const type = (message.type ?? "message") as keyof TMap;
    const decoder = this.decoders[type];
    if (!decoder) {
      this.options.onUnknown?.(message);
      return;
    }
    let payload: TMap[typeof type];
    try {
      payload = decoder(message.data);
    } catch (err) {
      this.options.onParseError?.(err, message);
      return;
    }
    // Cast narrows the variadic emit signature; the TMap[K] payload is
    // type-safe at the call site above.
    (this.events.emit as (type: keyof TMap, payload: unknown) => void)(
      type,
      payload,
    );
  };
}
