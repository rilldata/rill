import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createRuntimeServiceGetInstance,
  type V1AIPrompt,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived, type Readable } from "svelte/store";

/**
 * Picks the prompts to show from a list of sources in order of precedence:
 * the first source with at least one usable prompt wins and all of its prompts are shown.
 * The parser already caps each list, so no display limit is applied here.
 * For a dashboard the sources are its own `ai_prompts`, then the project's `ai_prompts` from rill.yaml.
 * Returns an empty list when nothing is configured, in which case no prompts are shown.
 */
export function resolveSuggestedPrompts(
  sources: (V1AIPrompt[] | undefined)[],
): V1AIPrompt[] {
  for (const source of sources) {
    const prompts = (source ?? []).filter((p) => p.prompt?.trim());
    if (prompts.length > 0) {
      return prompts;
    }
  }
  return [];
}

/** The project-wide `ai_prompts` from rill.yaml, exposed on the instance. */
export function createProjectPromptsStore(
  client: RuntimeClient,
): Readable<V1AIPrompt[] | undefined> {
  // Pass the query client explicitly so this can be created outside component initialization.
  const instanceQuery = createRuntimeServiceGetInstance(
    client,
    {},
    undefined,
    queryClient,
  );
  return derived(instanceQuery, ($q) => $q.data?.instance?.aiPrompts);
}
