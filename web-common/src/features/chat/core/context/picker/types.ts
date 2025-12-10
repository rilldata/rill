import type { InlineContext } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

export type InlineContextPickerOption = {
  context: InlineContext;
  recentlyUsed?: boolean;
  currentlyActive?: boolean;
  childContextCategories?: InlineContext[][]; // Only one level of child contexts is supported.
};
