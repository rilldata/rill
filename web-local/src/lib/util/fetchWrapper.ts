export async function fetchWrapperDirect(
  url: string,
  method: string,
  body?: BodyInit | Record<string, unknown>,
  headers: HeadersInit = { "Content-Type": "application/json" }
) {
  const resp = await fetch(url, {
    method,
    ...(body ? { body: serializeBody(body) } : {}),
    headers,
  });
  if (!resp.ok) {
    const err = new Error();
    (err as any).response = await resp.json();
    return Promise.reject(err);
  }
  return resp.json();
}
export async function fetchWrapper(
  path: string,
  method: string,
  body?: BodyInit | Record<string, unknown>,
  headers: HeadersInit = { "Content-Type": "application/json" }
) {
  const resp = await fetch(`${RILL_RUNTIME_URL}/api/${path}`, {
    method,
    ...(body ? { body: serializeBody(body) } : {}),
    headers,
  });
  if (!resp.ok) {
    const err = new Error();
    (err as any).response = await resp.json();
    return Promise.reject(err);
  }
  return (await resp.json())?.data;
}

export async function* streamingFetchWrapper<T>(
  path: string,
  method: string,
  body?: Record<string, unknown>
): AsyncGenerator<T> {
  let response: Response;
  try {
    response = await fetch(`${RILL_RUNTIME_URL}/api/${path}`, {
      method,
      ...(body ? { body: JSON.stringify(body) } : {}),
      headers: { "Content-Type": "application/json" },
    });
  } catch (err) {
    console.error(err);
    return;
  }
  const reader = response.body.getReader();
  const decoder = new TextDecoder();

  let readResult = await reader.read();
  while (!readResult.done) {
    const parts = decoder.decode(readResult.value).split("\n");
    for (const part of parts) {
      if (part === "") continue;
      try {
        const json = JSON.parse(part);
        yield json;
      } catch (err) {
        console.error(err);
      }
    }
    readResult = await reader.read();
  }
}

function serializeBody(body: BodyInit | Record<string, unknown>): BodyInit {
  return body instanceof FormData ? body : JSON.stringify(body);
}
