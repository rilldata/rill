/**
 * Wrapper around native fetch method for streaming requests.
 * Pass in {@link AbortSignal} to control cancellation.
 */
export async function* streamingFetchWrapper<T>(
  url: string,
  method: string,
  body?: Record<string, unknown>,
  headers: HeadersInit = { "Content-Type": "application/json" },
  signal?: AbortSignal,
): AsyncGenerator<T> {
  const response = await fetch(url, {
    method,
    ...(body ? { body: JSON.stringify(body) } : {}),
    headers,
    signal,
  });
  if (!response.body) {
    throw new Error("No response");
  }
  const reader = response.body.getReader();
  const decoder = new TextDecoder();

  let readResult = await reader.read();
  let prevPart = "";
  while (!readResult.done && !signal?.aborted) {
    const str = decoder.decode(readResult.value);
    const parts = str.split("\n");
    for (const part of parts) {
      if (part === "") continue;
      if (!part.endsWith("}")) {
        prevPart += part;
        continue;
      }
      try {
        const json = JSON.parse(prevPart + part);
        prevPart = "";
        yield json;
      } catch (err) {
        prevPart = part;
        // nothing
      }
    }
    readResult = await reader.read();
  }
}
