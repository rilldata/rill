import { getContext } from "svelte";
import type { RuntimeClient } from "./runtime-client";

export const RUNTIME_CONTEXT_KEY = Symbol("runtime-client");

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
