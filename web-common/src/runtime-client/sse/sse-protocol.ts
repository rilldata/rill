/**
 * Pure, dependency-free SSE protocol parser.
 *
 * Converts a byte stream in the Server-Sent Events wire format
 * (https://html.spec.whatwg.org/multipage/server-sent-events.html) into an
 * async iterable of SSEMessage values. Consumers decide what the messages
 * mean.
 */

/**
 * Represents a Server-Sent Event message.
 */
export interface SSEMessage {
  /** Event type. undefined means the default "message" type per the SSE spec. */
  type?: string;
  /** Raw event data, with multi-line `data:` fields joined by "\n". */
  data: string;
}

/**
 * Parse a single SSE line and accumulate it into `event`. Mutates `event`
 * for efficiency while streaming. Unknown fields (id:, retry:) are ignored.
 */
export function parseSSELine(line: string, event: Partial<SSEMessage>): void {
  const trimmedLine = line.trim();

  if (!trimmedLine) return;

  if (trimmedLine.startsWith(":")) return;

  if (trimmedLine.startsWith("event:")) {
    event.type = trimmedLine.slice(6).trim();
    return;
  }

  if (trimmedLine.startsWith("data:")) {
    const data = trimmedLine.slice(5).trim();
    event.data = event.data ? event.data + "\n" + data : data;
    return;
  }
}

/**
 * An empty line signals the end of an event in the SSE wire format.
 */
export function isEventComplete(line: string): boolean {
  return line.trim() === "";
}

/**
 * An event is emittable once it has a non-empty `data` field. The SSE spec
 * treats data-less events as protocol noise (e.g. keepalive comments).
 */
export function isValidEvent(
  event: Partial<SSEMessage>,
): event is SSEMessage {
  return event.data !== undefined && event.data !== "";
}

/**
 * Parse an SSE byte stream into an async iterable of decoded events.
 *
 * The parser:
 *   - Joins multi-line `data:` fields with "\n".
 *   - Ignores comments (lines starting with ":") and unknown fields.
 *   - Handles events split across chunk boundaries and both LF and CRLF
 *     line endings.
 *   - Yields events only when a complete blank-line boundary is seen, plus
 *     a final event if the stream ends mid-buffer with valid data.
 */
export async function* readSSEStream(
  body: ReadableStream<Uint8Array>,
): AsyncIterable<SSEMessage> {
  const reader = body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let currentEvent: Partial<SSEMessage> = {};

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      // Normalize CRLF so split("\n") alone handles both endings.
      const normalized = buffer.replace(/\r\n/g, "\n");
      const lines = normalized.split("\n");

      // The last element may be a partial line; hold it for the next chunk.
      buffer = lines.pop() ?? "";

      for (const line of lines) {
        if (isEventComplete(line)) {
          if (isValidEvent(currentEvent)) {
            yield currentEvent;
          }
          currentEvent = {};
        } else {
          parseSSELine(line, currentEvent);
        }
      }
    }

    // Flush any remaining event held in the final buffer.
    if (isValidEvent(currentEvent)) {
      yield currentEvent;
    }
  } finally {
    reader.releaseLock();
  }
}
