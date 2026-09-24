import {
  createRuntimeServiceGetInstance,
  type V1AIPrompt,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived, type Readable } from "svelte/store";

/** The chat shows at most this many starter prompts. */
export const MAX_SUGGESTED_PROMPTS = 4;

/**
 * Picks the prompts to show from a list of sources in order of precedence:
 * the first source with at least one usable prompt wins, capped at MAX_SUGGESTED_PROMPTS.
 * For a dashboard the sources are its own `ai_prompts`, then the project's `ai_prompts` from rill.yaml.
 * Returns an empty list when nothing is configured, in which case no prompts are shown.
 */
export function resolveSuggestedPrompts(
  sources: (V1AIPrompt[] | undefined)[],
): V1AIPrompt[] {
  for (const source of sources) {
    const prompts = (source ?? []).filter((p) => p.prompt?.trim());
    if (prompts.length > 0) {
      return prompts.slice(0, MAX_SUGGESTED_PROMPTS);
    }
  }
  return [];
}

/** The project-wide `ai_prompts` from rill.yaml, exposed on the instance. */
export function createProjectPromptsStore(
  client: RuntimeClient,
): Readable<V1AIPrompt[] | undefined> {
  const instanceQuery = createRuntimeServiceGetInstance(client, {});
  return derived(instanceQuery, ($q) => $q.data?.instance?.aiPrompts);
}
