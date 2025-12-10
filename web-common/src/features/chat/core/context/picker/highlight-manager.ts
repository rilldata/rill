import { type InlineContext } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { inlineContextsAreEqual } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { writable } from "svelte/store";
import type { InlineContextPickerOption } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import { ParentPickerTypes } from "@rilldata/web-common/features/chat/core/context/picker/data.ts";

export class PickerOptionsHighlightManager {
  public highlightedContext = writable<InlineContext | null>(null);

  private highlightableContexts: InlineContext[] = [];
  private highlightedIndex = -1;

  public filterOptionsUpdated(filteredOptions: InlineContextPickerOption[]) {
    // Convert the filtered options to a flat list for ease of navigation using arrow keys.
    this.highlightableContexts = filteredOptions.flatMap((option) => [
      option.context,
      ...(option.childContextCategories?.flat() ?? []),
    ]);

    // Prefer non-parent context if available for the parent option.
    const nonParentAvailable =
      this.highlightableContexts.length > 1 &&
      ParentPickerTypes.has(this.highlightableContexts[1].type);
    this.highlightedIndex = nonParentAvailable ? 1 : 0;
    this.updateHighlightedContext();
  }

  public highlightNextContext() {
    if (this.highlightableContexts.length === 0) return null;
    this.highlightedIndex =
      (this.highlightedIndex + 1) % this.highlightableContexts.length;
    this.updateHighlightedContext();
  }

  public highlightPreviousContext() {
    if (this.highlightableContexts.length === 0) return null;
    this.highlightedIndex =
      (this.highlightedIndex - 1 + this.highlightableContexts.length) %
      this.highlightableContexts.length;
    this.updateHighlightedContext();
  }

  public highlightContext(context: InlineContext) {
    const newIndex = this.highlightableContexts.findIndex((hc) =>
      inlineContextsAreEqual(context, hc),
    );
    if (newIndex !== -1) {
      this.highlightedIndex = newIndex;
      this.updateHighlightedContext();
    }
  }

  private updateHighlightedContext() {
    this.highlightedContext.set(
      this.highlightableContexts[this.highlightedIndex] ?? null,
    );
  }
}
