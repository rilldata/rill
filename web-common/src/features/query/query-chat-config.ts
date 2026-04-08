import {
  type ChatConfig,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types";
import type { RuntimeServiceCompleteBody } from "@rilldata/web-common/runtime-client";
import { derived, get, readable, writable, type Readable } from "svelte/store";
import type { NotebookStore } from "./query-store";

const emptyContext: Readable<Partial<RuntimeServiceCompleteBody>> = readable(
  {},
);

/** Set by QueryWorkspace when the notebook is created; read by the chat config. */
export const activeNotebook = writable<NotebookStore | null>(null);

function getFocusedCell() {
  const nb = get(activeNotebook);
  if (!nb) return null;
  const state = get(nb);
  return state.cells.find((c) => c.id === state.focusedCellId) ?? null;
}

export function createQueryChatConfig(): ChatConfig {
  return {
    agent: ToolName.ANALYST_AGENT,
    additionalContextStoreGetter: () => emptyContext,
    promptTransformer: (prompt) => {
      const cell = getFocusedCell();
      if (!cell?.sql?.trim()) return prompt;
      const limitNote =
        cell.limit !== undefined ? `\nLimit: ${cell.limit} rows` : "";
      return `[Current SQL query (${cell.connector} connector)${limitNote}]\n\`\`\`sql\n${cell.sql}\n\`\`\`\n\n${prompt}`;
    },
    contextIndicator: () =>
      derived(activeNotebook, ($nb) => {
        if (!$nb) return null;
        const state = get($nb);
        const cell = state.cells.find((c) => c.id === state.focusedCellId);
        if (!cell?.sql?.trim()) return null;
        return `SQL query attached (${cell.connector})`;
      }),
    emptyChatLabel: "Ask for help with your SQL",
    placeholder: "Ask about your SQL query...",
    minChatHeight: "min-h-[4rem]",
  };
}
