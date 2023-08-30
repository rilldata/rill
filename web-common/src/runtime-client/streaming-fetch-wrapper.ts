export async function* streamingFetchWrapper<T>(
  url: string,
  method: string,
  body?: Record<string, unknown>,
  signal?: AbortSignal
): AsyncGenerator<T> {
  let response: Response;
  try {
    response = await fetch(url, {
      method,
      ...(body ? { body: JSON.stringify(body) } : {}),
      headers: { "Content-Type": "application/json" },
      signal,
    });
  } catch (err) {
    return;
  }
  if (!response.body) {
    return;
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
