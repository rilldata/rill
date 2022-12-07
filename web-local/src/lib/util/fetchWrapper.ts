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
    const paramParts = [];
    for (const p in params) {
      paramParts.push(`${p}=${encodeURIComponent(params[p] as string)}`);
    }
    if (paramParts.length) {
      url = `${url}?${paramParts.join("&")}`;
    }
  }

  const resp = await fetch(url, {
    method,
    ...(data ? { body: serializeBody(data) } : {}),
    headers,
    signal,
  });
  if (!resp.ok) {
    const data = await resp.json();

    // Return runtime errors in the same form as the Axios client had previously
    if (data.code && data.message) {
      return Promise.reject({
        response: {
          status: resp.status,
          data,
        },
      });
    }

    // Fallback error handling
    const err = new Error();
    (err as any).response = await resp.json();
    return Promise.reject(err);
  }
  return resp.json();
}

function serializeBody(body: BodyInit | Record<string, unknown>): BodyInit {
  return body instanceof FormData ? body : JSON.stringify(body);
}
