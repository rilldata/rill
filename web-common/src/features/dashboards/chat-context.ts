import {
  type ChatConfig,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";
import { getExploreNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { useStableExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
import { isExpressionEmpty } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { createStableTimeControlStoreFromName } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import type {
  RuntimeServiceCompleteBody,
  V1AnalystAgentContext,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived, type Readable } from "svelte/store";

export function createDashboardChatConfig(client: RuntimeClient): ChatConfig {
  const activeExploreContextStore = getActiveExploreContext(client);
  return {
    agent: ToolName.ANALYST_AGENT,
    additionalContextStoreGetter: () => activeExploreContextStore,
    emptyChatLabel: "Happy to help explore your data",
    placeholder:
      "Type a question, or press @ to insert a metric, dimension, or measure.",
    minChatHeight: "min-h-[4rem]",
  };
}

/**
 * Creates a store that contains the active explore context sent to the Complete API.
 * It returns RuntimeServiceCompleteBody with V1AnalystAgentContext that is passed to the API.
 */
function getActiveExploreContext(
  client: RuntimeClient,
): Readable<Partial<RuntimeServiceCompleteBody>> {
  const exploreNameStore = getExploreNameStore();

  const exploreState = useStableExploreState(exploreNameStore);
  const timeControlsStore = createStableTimeControlStoreFromName(
    client,
    exploreNameStore,
  );

  return derived(
    [exploreNameStore, exploreState, timeControlsStore],
    ([exploreName, exploreState, timeControlsStore]) => {
      const analystAgentContext: V1AnalystAgentContext = {
        explore: exploreName,
      };

      if (
        timeControlsStore.selectedTimeRange?.start &&
        timeControlsStore.selectedTimeRange?.end
      ) {
        analystAgentContext.timeStart =
          timeControlsStore.selectedTimeRange.start.toISOString();
        analystAgentContext.timeEnd =
          timeControlsStore.selectedTimeRange.end.toISOString();
      }

      if (
        exploreState?.visibleDimensions.length &&
        !exploreState?.allDimensionsVisible
      ) {
        analystAgentContext.dimensions = exploreState.visibleDimensions;
      }

      if (
        exploreState?.visibleMeasures.length &&
        !exploreState?.allMeasuresVisible
      ) {
        analystAgentContext.measures = exploreState.visibleMeasures;
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
