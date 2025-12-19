import {
  type InlineContext,
  inlineContextIsWithin,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { inlineContextsAreEqual } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { get, writable } from "svelte/store";
import type { InlineContextPickerSection } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import { ParentPickerTypes } from "@rilldata/web-common/features/chat/core/context/picker/data.ts";

export class PickerOptionsHighlightManager {
  public highlightedContext = writable<InlineContext | null>(null);

  private highlightableContexts: InlineContext[] = [];
  private highlightedIndex = -1;

  public filterOptionsUpdated(
    filteredSections: InlineContextPickerSection[],
    selectedChatContext: InlineContext | null,
  ) {
    // Convert the filtered options to a flat list for ease of navigation using arrow keys.
    this.highlightableContexts = filteredSections
      .flatMap((section) => section.options)
      .flatMap((parentOption) => [
        parentOption.context,
        ...(parentOption.children?.flatMap((c) => c.options) ?? []),
      ]);

    // Prefer selecting already selected context.
    const currentHighlightedContext = get(this.highlightedContext);
    if (currentHighlightedContext) {
      const currentHighlightedIndex = this.highlightableContexts.findIndex(
        (c) => inlineContextsAreEqual(c, currentHighlightedContext),
      );
      if (currentHighlightedIndex !== -1) {
        this.highlightedIndex = currentHighlightedIndex;
        this.updateHighlightedContext();
        return;
      }

      if (currentHighlightedContext.type === InlineContextType.Loading) {
        const parentIndex = this.highlightableContexts.findIndex((c) =>
          inlineContextIsWithin(c, currentHighlightedContext),
        );
        if (parentIndex !== -1) {
          this.highlightedIndex = parentIndex + 1;
          this.updateHighlightedContext();
          return;
        }
      }
    }

    // Next prefer the selected context if available.
    // TODO: only do this for the 1st time perhaps?
    const selectedContextAvailable =
      selectedChatContext &&
      this.highlightableContexts.some((c) =>
        inlineContextsAreEqual(c, selectedChatContext),
      );
    if (selectedContextAvailable) {
      this.highlightContext(selectedChatContext);
      return;
    }

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

  public highlightContext(context: InlineContext | null) {
    if (!context) {
      this.highlightedIndex = -1;
      this.updateHighlightedContext();
      return false;
    }

    const newIndex = this.highlightableContexts.findIndex((hc) =>
      inlineContextsAreEqual(context, hc),
    );
    if (newIndex === -1) return false;

    this.highlightedIndex = newIndex;
    this.updateHighlightedContext();
    return true;
  }

  public mouseOverHandler(ctx: InlineContext) {
    return (node: Node) => {
      const onMouseEnter = () => this.highlightContext(ctx);
      node.addEventListener("mouseenter", onMouseEnter);
      return {
        destroy() {
          node.removeEventListener("mouseenter", onMouseEnter);
        },
      };
    };
  }

  private updateHighlightedContext() {
    this.highlightedContext.set(
      this.highlightableContexts[this.highlightedIndex] ?? null,
    );
  }
}
