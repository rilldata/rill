import { getRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

// Detect dev mode without depending on SvelteKit's $app/environment
// (which is unavailable to tsc outside Vite/SvelteKit context)
const isDev =
  typeof import.meta.env !== "undefined" && import.meta.env.DEV === true;

export const LOCAL_HOST = isDev ? "http://localhost:9009" : "";
export const LOCAL_INSTANCE_ID = "default";

/** For load functions and tests. In components, use {@link useRuntimeClient} instead. */
export function getLocalRuntimeClient(): RuntimeClient {
  return getRuntimeClient({ host: LOCAL_HOST, instanceId: LOCAL_INSTANCE_ID });
}
