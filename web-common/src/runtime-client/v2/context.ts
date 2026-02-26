import { getContext } from "svelte";
import { writable } from "svelte/store";
import type { RuntimeClient } from "./runtime-client";

export const RUNTIME_CONTEXT_KEY = Symbol("runtime-client");

/**
 * Module-level store that mirrors the active RuntimeClient.
 * Set by RuntimeProvider on mount, cleared on destroy.
 *
 * Used by components that render OUTSIDE RuntimeProvider's subtree
 * (e.g. TopNavigationBar in the root layout) but need reactive
 * access to the current RuntimeClient.
 */
export const runtimeClientStore = writable<RuntimeClient | null>(null);

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
