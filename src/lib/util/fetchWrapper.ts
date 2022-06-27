import { config } from "$lib/application-state-stores/application-store";

export async function fetchWrapper(
  path: string,
  method: string,
  body?: Record<string, unknown>
) {
  const resp = await fetch(`${config.server.serverUrl}/api/${path}`, {
    method,
    ...(body ? { body: JSON.stringify(body) } : {}),
    headers: { "Content-Type": "application/json" },
  });
  return (await resp.json())?.data;
}

export async function* streamingFetchWrapper<T>(
  path: string,
  method: string,
  body?: Record<string, unknown>
): AsyncGenerator<T> {
  const response = await fetch(`${config.server.serverUrl}/api/${path}`, {
    method,
    ...(body ? { body: JSON.stringify(body) } : {}),
    headers: { "Content-Type": "application/json" },
  });
  const reader = response.body.getReader();
  const decoder = new TextDecoder();

  let readResult = await reader.read();
  while (!readResult.done) {
    const parts = decoder.decode(readResult.value).split("\n");
    for (const part of parts) {
      if (part === "") continue;
      try {
        yield JSON.parse(part);
      } catch (err) {
        console.error(err);
      }
    }
    readResult = await reader.read();
  }
}
