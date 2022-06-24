import { config } from "$lib/application-state-stores/application-store";

export async function fetchWrapper(path: string, method: string, body?: any) {
  const resp = await fetch(`${config.server.serverUrl}/api/${path}`, {
    method,
    ...(body ? { body: JSON.stringify(body) } : {}),
    headers: { "Content-Type": "application/json" },
  });
  return (await resp.json())?.data;
}
