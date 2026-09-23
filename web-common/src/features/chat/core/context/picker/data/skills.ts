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

// Value in a skill's `agents` for the analyst agent, which answers the chat.
const ANALYST_SKILL_AGENT = "analyst";

/**
 * Creates a store that contains a flat list of the project's skills for the analyst agent.
 */
export function getSkillsPickerOptions(
  client: RuntimeClient,
): Readable<PickerItem[]> {
  const skillResourcesQuery = createQuery(
    getClientFilteredResourcesQueryOptions(client, ResourceKind.Skill),
    queryClient,
  );

  return derived(skillResourcesQuery, (skillResourcesResp) =>
    getSkillPickerItems(skillResourcesResp.data ?? []),
  );
}

/**
 * Returns picker items for the skills the analyst agent can load:
 * the ones that apply to the analyst and reconciled without errors.
 */
export function getSkillPickerItems(resources: V1Resource[]): PickerItem[] {
  return resources
    .filter(
      (res) =>
        res.skill?.spec?.agents?.includes(ANALYST_SKILL_AGENT) &&
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
