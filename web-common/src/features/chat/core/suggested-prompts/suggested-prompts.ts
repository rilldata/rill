import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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
 * For a dashboard the sources are: configured `ai_prompts`, AI-generated prompts, project `ai_prompts`, hardcoded fallback.
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

/** Generic prompts for a dashboard chat, used when nothing is configured or generated. */
export function getDashboardFallbackPrompts(): V1AIPrompt[] {
  return [
    { label: m.chat_prompt_overview_label(), prompt: m.chat_prompt_overview() },
    { label: m.chat_prompt_movers_label(), prompt: m.chat_prompt_movers() },
    {
      label: m.chat_prompt_top_contributors_label(),
      prompt: m.chat_prompt_top_contributors(),
    },
    {
      label: m.chat_prompt_anomalies_label(),
      prompt: m.chat_prompt_anomalies(),
    },
  ];
}

/** Generic prompts for the project-wide chat, used when rill.yaml configures none. */
export function getProjectFallbackPrompts(): V1AIPrompt[] {
  return [
    {
      label: m.chat_prompt_project_data_label(),
      prompt: m.chat_prompt_project_data(),
    },
    {
      label: m.chat_prompt_project_dashboards_label(),
      prompt: m.chat_prompt_project_dashboards(),
    },
    {
      label: m.chat_prompt_project_metrics_label(),
      prompt: m.chat_prompt_project_metrics(),
    },
    {
      label: m.chat_prompt_project_recent_label(),
      prompt: m.chat_prompt_project_recent(),
    },
  ];
}

/** The project-wide `ai_prompts` from rill.yaml, exposed on the instance. */
export function createProjectPromptsStore(
  client: RuntimeClient,
): Readable<V1AIPrompt[] | undefined> {
  const instanceQuery = createRuntimeServiceGetInstance(client, {});
  return derived(instanceQuery, ($q) => $q.data?.instance?.aiPrompts);
}
