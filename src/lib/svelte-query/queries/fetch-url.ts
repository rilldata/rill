import { config } from "$lib/application-state-stores/application-store";

export async function fetchUrl(path: string, method: string, body?) {
  const resp = await fetch(`${config.server.serverUrl}/api/v1/${path}`, {
    method,
    headers: { "Content-Type": "application/json" },
    ...(body ? { body: JSON.stringify(body) } : {}),
  });
  const json = await resp.json();
  if (!resp.ok) {
    const err = new Error(json.messages[0].message);
    return Promise.reject(err);
  }
  return json;
}
