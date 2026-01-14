import {
  type ChatConfig,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";
import type {
  RuntimeServiceCompleteBody,
  V1AnalystAgentContext,
} from "@rilldata/web-common/runtime-client";
import { getExploreNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { useStableExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
import { createStableTimeControlStoreFromName } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { isExpressionEmpty } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { derived, type Readable } from "svelte/store";

export const dashboardChatConfig = {
  agent: ToolName.ANALYST_AGENT,
  additionalContextStoreGetter: () => getActiveExploreContext(), // TODO: add canvas context as well
  emptyChatLabel: "Happy to help explore your data",
  placeholder:
    "Type a question, or press @ to insert a metric, dimension, or measure.",
  minChatHeight: "min-h-[4rem]",
} satisfies ChatConfig;

/**
 * Creates a store that contains the active explore context sent to the Complete API.
 * It returns RuntimeServiceCompleteBody with V1AnalystAgentContext that is passed to the API.
 */
function getActiveExploreContext(): Readable<
  Partial<RuntimeServiceCompleteBody>
> {
  const exploreNameStore = getExploreNameStore();

  const exploreState = useStableExploreState(exploreNameStore);
  const timeControlsStore =
    createStableTimeControlStoreFromName(exploreNameStore);

  return derived(
    [exploreNameStore, exploreState, timeControlsStore],
    ([exploreName, exploreState, timeControlsStore]) => {
      const analystAgentContext: V1AnalystAgentContext = {
        explore: exploreName,
      };

      if (timeControlsStore?.timeStart && timeControlsStore?.timeEnd) {
        analystAgentContext.timeStart = timeControlsStore.timeStart;
        analystAgentContext.timeEnd = timeControlsStore.timeEnd;
      }

      const filterIsAvailable = !isExpressionEmpty(exploreState?.whereFilter);
      if (filterIsAvailable) {
        analystAgentContext.where = exploreState?.whereFilter;
      }

      return {
        analystAgentContext,
      } satisfies Partial<RuntimeServiceCompleteBody>;
    },
  );
}
