import {
  type ChatConfig,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";
import type {
  RuntimeServiceCompleteBody,
  V1AIPrompt,
  V1AnalystAgentContext,
} from "@rilldata/web-common/runtime-client";
import { getCanvasNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { derived, type Readable } from "svelte/store";
import { getCanvasStoreUnguarded } from "@rilldata/web-common/features/canvas/state-managers/state-managers.ts";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import {
  createProjectPromptsStore,
  getDashboardFallbackPrompts,
  resolveSuggestedPrompts,
} from "@rilldata/web-common/features/chat/core/suggested-prompts/suggested-prompts.ts";
import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

export function createCanvasChatConfig(client: RuntimeClient): ChatConfig {
  return {
    agent: ToolName.ANALYST_AGENT,
    additionalContextStoreGetter: () => getActiveCanvasContext(client),
    suggestedPromptsStoreGetter: getCanvasSuggestedPrompts,
    emptyChatLabel: m.chat_happy_to_explore(),
    placeholder: m.chat_placeholder_analyst(),
    minChatHeight: "min-h-[4rem]",
  };
}

/**
 * Creates a store with the starter prompts for the active canvas:
 * its configured `ai_prompts`, else the prompts generated during reconciliation, else the project's `ai_prompts`, else a generic fallback.
 */
function getCanvasSuggestedPrompts(
  client: RuntimeClient,
): Readable<V1AIPrompt[]> {
  const canvasNameStore = getCanvasNameStore();
  const projectPromptsStore = createProjectPromptsStore(client);

  return derived(
    [canvasNameStore, projectPromptsStore],
    ([canvasName, projectPrompts], set) => {
      const canvasQuery = useResource(client, canvasName, ResourceKind.Canvas);
      return canvasQuery.subscribe((res) => {
        const canvas = res.data?.canvas;
        set(
          resolveSuggestedPrompts([
            canvas?.state?.validSpec?.aiPrompts ?? canvas?.spec?.aiPrompts,
            canvas?.state?.aiSuggestedPrompts,
            projectPrompts,
            getDashboardFallbackPrompts(),
          ]),
        );
      });
    },
    [] as V1AIPrompt[],
  );
}

/**
 * Creates a store that contains the active canvas context sent to the Complete API.
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
        canvasStore.canvasEntity.expressionFilterManager.exprByMetricsViewStore,
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
        analystAgentContext.timeStart = selectedInterval.start.toUTC().toISO();
        analystAgentContext.timeEnd = selectedInterval.end.toUTC().toISO();
      }

      if (Object.keys(filtersMap).length > 0) {
        analystAgentContext.wherePerMetricsView = {};
        Object.entries(filtersMap).forEach(([mv, expr]) => {
          if (expr.cond?.exprs?.length) {
            analystAgentContext.wherePerMetricsView![mv] = expr;
          }
        });
      }

      set({ analystAgentContext });
    });
  });
}
