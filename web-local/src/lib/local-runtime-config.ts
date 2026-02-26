import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

// Detect dev mode without depending on SvelteKit's $app/environment
// (which is unavailable to tsc outside Vite/SvelteKit context)
const isDev =
  typeof import.meta.env !== "undefined" && import.meta.env.DEV === true;

export const LOCAL_HOST = isDev ? "http://localhost:9009" : "";
export const LOCAL_INSTANCE_ID = "default";

// Singleton RuntimeClient for use in load functions (outside Svelte context).
// Web-local has no JWT auth, so this is a simple, long-lived instance.
let _localClient: RuntimeClient | null = null;
export function getLocalRuntimeClient(): RuntimeClient {
  if (!_localClient) {
    _localClient = new RuntimeClient({
      host: LOCAL_HOST,
      instanceId: LOCAL_INSTANCE_ID,
    });
  }
  return _localClient;
}
