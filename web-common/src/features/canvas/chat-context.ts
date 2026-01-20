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

export const canvasChatConfig = {
  agent: ToolName.ANALYST_AGENT,
  additionalContextStoreGetter: () => getActiveCanvasContext(),
  emptyChatLabel: "Happy to help explore your data",
  placeholder:
    "Type a question, or press @ to insert a metric, dimension, or measure.",
  minChatHeight: "min-h-[4rem]",
} satisfies ChatConfig;

/**
 * Creates a store that contains the active explore context sent to the Complete API.
 * It returns RuntimeServiceCompleteBody with V1AnalystAgentContext that is passed to the API.
 */
function getActiveCanvasContext(): Readable<
  Partial<RuntimeServiceCompleteBody>
> {
  const canvasNameStore = getCanvasNameStore();

  return derived([canvasNameStore], ([canvasName]) => {
    const analystAgentContext: V1AnalystAgentContext = {
      canvas: canvasName,
    };

    // TODO: canvas state

    return {
      analystAgentContext,
    } satisfies Partial<RuntimeServiceCompleteBody>;
  });
}
