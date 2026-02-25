import {
  type ChatConfig,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";
import type {
  RuntimeServiceCompleteBody,
  V1AnalystAgentContext,
} from "@rilldata/web-common/runtime-client";
import { getCanvasNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { derived, type Readable } from "svelte/store";
import { getCanvasStoreUnguarded } from "@rilldata/web-common/features/canvas/state-managers/state-managers.ts";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

export function createCanvasChatConfig(client: RuntimeClient): ChatConfig {
  return {
    agent: ToolName.ANALYST_AGENT,
    additionalContextStoreGetter: () => getActiveCanvasContext(client),
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
function getActiveCanvasContext(
  client: RuntimeClient,
): Readable<Partial<RuntimeServiceCompleteBody>> {
  const instanceId = client.instanceId;
  const canvasNameStore = getCanvasNameStore();

  return derived([canvasNameStore], ([canvasName], set) => {
    const canvasStore = getCanvasStoreUnguarded(canvasName, instanceId);
    if (!canvasStore?.canvasEntity) {
      set({ analystAgentContext: { canvas: canvasName } });
      return;
    }

    const canvasFiltersStore = derived(
      [
        canvasStore.canvasEntity.filterManager.filterMapStore,
        canvasStore.canvasEntity.timeManager.state.interval,
      ],
      ([filtersMap, selectedInterval]) => {
        return {
          filtersMap,
          selectedInterval,
        };
      },
    );

    return canvasFiltersStore.subscribe(({ filtersMap, selectedInterval }) => {
      const analystAgentContext: V1AnalystAgentContext = {
        canvas: canvasName,
      };

      if (selectedInterval?.isValid) {
        analystAgentContext.timeStart = selectedInterval.start.toISO();
        analystAgentContext.timeEnd = selectedInterval.end.toISO();
      }

      if (filtersMap.size) {
        analystAgentContext.wherePerMetricsView = {};
        filtersMap.forEach((expr, mv) => {
          if (expr.cond?.exprs?.length) {
            analystAgentContext.wherePerMetricsView![mv] = expr;
          }
        });
      }

      set({ analystAgentContext });
    });
  });
}
