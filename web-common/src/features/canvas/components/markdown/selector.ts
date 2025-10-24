import type { MarkdownSpec } from "@rilldata/web-common/features/canvas/components/markdown";
import { derived, type Readable } from "svelte/store";

/**
 * Validate Markdown schema
 */
export function validateMarkdownSchema(
  markdownSpec: MarkdownSpec,
): Readable<{
  isValid: boolean;
  error?: string;
  isLoading?: boolean;
}> {
  return derived([], () => {
    if (!markdownSpec.content || markdownSpec.content.trim() === "") {
      return {
        isValid: false,
        error: "Markdown content cannot be empty",
      };
    }

    return {
      isValid: true,
      error: undefined,
    };
  });
}

