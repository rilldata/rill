import { type ChatConfig } from "@rilldata/web-common/features/chat/core/types.ts";
import type { RuntimeServiceCompleteBody } from "@rilldata/web-common/runtime-client";
import { derived, type Readable } from "svelte/store";
import { page } from "$app/stores";

export const developerChatConfig = {
  additionalContextStoreGetter: () => getActiveFileContext(),
  emptyChatLabel: "Ask me about your data or make changes to the project",
  placeholder: "Ask a question or request a change...",
  minChatHeight: "min-h-[2.5rem]",
} satisfies ChatConfig;

/**
 * Creates a store that contains the active file context sent to the Complete API.
 * It returns the RuntimeServiceCompleteBody with V1DeveloperAgentContext that is passed to the API.
 */
function getActiveFileContext(): Readable<Partial<RuntimeServiceCompleteBody>> {
  return derived(page, (pageState) => {
    const filePath = pageState.params?.file;
    if (!filePath) return {} satisfies Partial<RuntimeServiceCompleteBody>;

    return {
      developerAgentContext: {
        currentFilePath: filePath,
      },
    } satisfies Partial<RuntimeServiceCompleteBody>;
  });
}
