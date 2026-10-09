import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived, type Readable } from "svelte/store";
import { createQuery } from "@tanstack/svelte-query";
import {
  getClientFilteredResourcesQueryOptions,
  ResourceKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import {
  getIdForContext,
  type InlineContext,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import type { PickerItem } from "@rilldata/web-common/features/chat/core/context/picker/picker-tree.ts";
import { ToolName } from "@rilldata/web-common/features/chat/core/types.ts";

/**
 * The value in a skill's `agents` for the agent that answers a chat, by the chat's agent tool.
 * Matches the agents the runtime parser accepts in a skill (see `runtime/parser/parse_skill.go`).
 */
const SKILL_AGENT_BY_CHAT_AGENT: Record<string, string> = {
  [ToolName.ANALYST_AGENT]: "analyst",
  [ToolName.DEVELOPER_AGENT]: "developer",
};

/**
 * Returns the value in a skill's `agents` for the agent that answers a chat, or undefined if that agent has no skills.
 */
export function getSkillAgent(chatAgent: string): string | undefined {
  return SKILL_AGENT_BY_CHAT_AGENT[chatAgent];
}

/**
 * Creates a store that contains a flat list of the project's skills for the given skill agent.
 */
export function getSkillsPickerOptions(
  client: RuntimeClient,
  skillAgent: string,
): Readable<PickerItem[]> {
  const skillResourcesQuery = createQuery(
    getClientFilteredResourcesQueryOptions(client, ResourceKind.Skill),
    queryClient,
  );

  return derived(skillResourcesQuery, (skillResourcesResp) =>
    getSkillPickerItems(skillResourcesResp.data ?? [], skillAgent),
  );
}

/**
 * Returns picker items for the skills the given agent can load:
 * the ones that apply to the agent and reconciled without errors.
 */
export function getSkillPickerItems(
  resources: V1Resource[],
  skillAgent: string,
): PickerItem[] {
  return resources
    .filter(
      (res) =>
        res.skill?.spec?.agents?.includes(skillAgent) &&
        !res.meta?.reconcileError,
    )
    .map((res) => {
      const name = res.meta?.name?.name ?? "";
      const context = {
        type: InlineContextType.Skill,
        skill: name,
        value: name,
        label: name,
      } satisfies InlineContext;

      return {
        id: getIdForContext(context),
        context,
        description: res.skill?.spec?.description,
      } satisfies PickerItem;
    });
}
