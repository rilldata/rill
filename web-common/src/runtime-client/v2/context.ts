import { getContext } from "svelte";
import { RuntimeClient, type AuthContext } from "./runtime-client";

export const RUNTIME_CONTEXT_KEY = Symbol("runtime-client");

// ── Cached factory ──────────────────────────────────────────────────

const clientCache = new Map<string, RuntimeClient>();

function cacheKey(host: string, instanceId: string): string {
  return `${host}::${instanceId}`;
}

/**
 * Returns a cached RuntimeClient for the given config, creating one if needed.
 * If a client already exists for the same host+instanceId, its JWT is updated
 * and the existing instance is returned.
 *
 * This is THE way to obtain a RuntimeClient outside of Svelte component context
 * (e.g. in SvelteKit load functions). Inside components, use `useRuntimeClient()`.
 */
export function getRuntimeClient(config: {
  host: string;
  instanceId: string;
  jwt?: string;
  authContext?: string;
}): RuntimeClient {
  const authContext = config.authContext as AuthContext | undefined;
  const key = cacheKey(config.host, config.instanceId);
  let client = clientCache.get(key);
  if (client) {
    client.updateJwt(config.jwt, authContext);
    return client;
  }
  client = new RuntimeClient({
    host: config.host,
    instanceId: config.instanceId,
    jwt: config.jwt,
    authContext,
  });
  clientCache.set(key, client);
  return client;
}

/**
 * Removes a client from the cache. Called by RuntimeProvider on destroy
 * so the next caller gets a fresh instance.
 */
export function evictRuntimeClient(client: RuntimeClient): void {
  const key = cacheKey(client.host, client.instanceId);
  if (clientCache.get(key) === client) {
    clientCache.delete(key);
  }
}

/**
 * Returns the RuntimeClient set by the nearest ancestor RuntimeProvider.
 * Must be called during component initialization (top-level `<script>`).
 */
export function useRuntimeClient(): RuntimeClient {
  const client = getContext<RuntimeClient | undefined>(RUNTIME_CONTEXT_KEY);
  if (!client) {
    throw new Error(
      "useRuntimeClient() was called outside of a <RuntimeProvider>. " +
        "Ensure a RuntimeProvider is an ancestor of this component.",
    );
  }
  return client;
}

/**
 * Like useRuntimeClient(), but returns null instead of throwing
 * when no RuntimeProvider ancestor exists. Useful for components
 * that render both inside and outside a runtime context (e.g. navigation bars).
 */
export function tryUseRuntimeClient(): RuntimeClient | null {
  return getContext<RuntimeClient | undefined>(RUNTIME_CONTEXT_KEY) ?? null;
}
