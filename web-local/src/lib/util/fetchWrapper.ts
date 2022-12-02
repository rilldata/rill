import type { RequestQueueEntry } from "@rilldata/web-local/lib/http-request-queue/HttpRequestQueueTypes";

export async function fetchWrapper({
  url,
  method,
  headers,
  data,
  params,
  signal,
}: RequestQueueEntry) {
  if (signal && signal.aborted) return Promise.reject(new Error("Aborted"));

  headers ??= { "Content-Type": "application/json" };

  if (params) {
    const u = new URL(url);
    for (const p in params) {
      u.searchParams.append(p, params[p]);
    }
    url = u.toString();
  }

  const resp = await fetch(url, {
    method,
    ...(data ? { body: serializeBody(data) } : {}),
    headers,
    signal,
  });
  if (!resp.ok) {
    const err = new Error();
    (err as any).response = await resp.json();
    return Promise.reject(err);
  }
  return resp.json();
}

function serializeBody(body: BodyInit | Record<string, unknown>): BodyInit {
  return body instanceof FormData ? body : JSON.stringify(body);
}
