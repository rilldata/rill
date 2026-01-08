import { type InlineContext } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { inlineContextsAreEqual } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { get, writable } from "svelte/store";
import type { InlineContextPickerParentOption } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import { ParentPickerTypes } from "@rilldata/web-common/features/chat/core/context/picker/data";
import type { ContextPickerUIState } from "@rilldata/web-common/features/chat/core/context/picker/ui-state.ts";

export class PickerOptionsHighlightManager {
  public highlightedContext = writable<InlineContext | null>(null);

  private highlightedIndex = -1;
  private metadata: HighlightManagerMetadata;

  public constructor(private readonly uiState: ContextPickerUIState) {
    this.metadata = new HighlightManagerMetadata(this.uiState, []);
  }

  public filterOptionsUpdated(
    filteredOptions: InlineContextPickerParentOption[],
    selectedChatContext: InlineContext | null,
  ) {
    const newMetadata = new HighlightManagerMetadata(
      this.uiState,
      filteredOptions,
    );

    const currentHighlightedContext = get(this.highlightedContext);
    this.highlightedIndex = newMetadata.getUpdatedIndex(
      currentHighlightedContext,
      selectedChatContext,
      this.metadata,
    );
    this.metadata = newMetadata;
    this.updateHighlightedContext();
  }

  public highlightNextContext() {
    if (this.metadata.highlightableContexts.length === 0) return null;
    this.highlightedIndex = this.metadata.nextValidIndex(this.highlightedIndex);
    this.updateHighlightedContext();
  }

  public highlightPreviousContext() {
    if (this.metadata.highlightableContexts.length === 0) return null;
    this.highlightedIndex = this.metadata.prevValidIndex(this.highlightedIndex);
    this.updateHighlightedContext();
  }

  public collapseToClosestParent() {
    const currentHighlightedContext = get(this.highlightedContext);
    if (!currentHighlightedContext) {
      return;
    }
    const parentIndex = this.metadata.closestParentIndex(
      currentHighlightedContext,
      this.highlightedIndex,
    );
    if (parentIndex === -1) return;
    const parentOption = this.metadata.parentOptionMap.get(parentIndex);
    if (parentOption && this.uiState.isExpanded(parentOption.context.key)) {
      this.uiState.collapse(parentOption.context.key);
    }
    this.highlightedIndex = parentIndex;
    this.updateHighlightedContext();
  }

  public openCurrentParentOption() {
    const parentOption = this.metadata.parentOptionMap.get(
      this.highlightedIndex,
    );
    if (parentOption) this.uiState.expand(parentOption.context.key);
  }

  public highlightContext(context: InlineContext | null) {
    if (!context) {
      this.highlightedIndex = -1;
      this.updateHighlightedContext();
      return;
    }

    // Save on running indexOf if the context is already highlighted.
    if (context === get(this.highlightedContext)) return;

    const newIndex = this.metadata.indexOf(context);
    if (newIndex === -1) return;

    this.highlightedIndex = newIndex;
    this.updateHighlightedContext();
  }

  private updateHighlightedContext() {
    this.highlightedContext.set(
      this.metadata.highlightableContexts[this.highlightedIndex] ?? null,
    );
  }
}

class HighlightManagerMetadata {
  public readonly highlightableContexts: InlineContext[];
  public readonly parentOptionMap = new Map<
    number,
    InlineContextPickerParentOption
  >();
  private readonly closestParentIndices: number[];

  public constructor(
    private readonly uiState: ContextPickerUIState,
    filteredOptions: InlineContextPickerParentOption[],
  ) {
    // Convert the filtered options to a flat list for ease of navigation using arrow keys.
    let cursor = 0;
    this.highlightableContexts = filteredOptions.flatMap((parent) => {
      this.parentOptionMap.set(cursor, parent);
      const children = parent.children ?? [];
      const optionsForParent = [parent.context, ...children];
      cursor += optionsForParent.length;
      return optionsForParent;
    });

    this.closestParentIndices = new Array<number>(
      this.highlightableContexts.length,
    );
    for (let i = 0, pi = 0; i < this.highlightableContexts.length; i++) {
      if (ParentPickerTypes.has(this.highlightableContexts[i].type)) {
        pi = i;
      }
      this.closestParentIndices[i] = pi;
    }
  }

  public getUpdatedIndex(
    highlightedContext: InlineContext | null,
    selectedContext: InlineContext | null,
    prevMetadata: HighlightManagerMetadata,
  ) {
    if (this.highlightableContexts.length === 0) return -1;

    // First, prefer the current highlighted context or its closest parent.
    if (highlightedContext) {
      // Prefer the context itself if it is available.
      const highlightedContextIndex = this.indexOf(highlightedContext);
      if (highlightedContextIndex !== -1) return highlightedContextIndex;

      // Otherwise prefer the closest parent.
      // Find the parent context in the previous metadata.
      const prevParentIndex =
        prevMetadata.closestParentIndex(highlightedContext);
      if (prevParentIndex !== -1) {
        // Find the parent context in the new metadata.
        const parentContext =
          prevMetadata.highlightableContexts[prevParentIndex];
        const parentIndex = parentContext ? this.indexOf(parentContext) : -1;
        if (parentIndex !== -1) return parentIndex;
      }
    }

    // Then prefer the selected context if available.
    if (selectedContext) return this.indexOf(selectedContext);

    // Finally, prefer non-parent context if available for the parent option.
    const nonParentAvailable =
      this.highlightableContexts.length > 1 &&
      ParentPickerTypes.has(this.highlightableContexts[1].type);
    return nonParentAvailable ? 1 : 0;
  }

  public contains(context: InlineContext) {
    return this.highlightableContexts.some((hc) =>
      inlineContextsAreEqual(context, hc),
    );
  }

  public indexOf(context: InlineContext) {
    return this.highlightableContexts.findIndex((hc) =>
      inlineContextsAreEqual(context, hc),
    );
  }

  public closestParentIndex(
    context: InlineContext,
    index = this.indexOf(context),
  ) {
    if (index === -1 || index >= this.closestParentIndices.length) return -1;
    return this.closestParentIndices[index];
  }

  public nextValidIndex(index: number) {
    let nextIndex = (index + 1) % this.highlightableContexts.length;

    const currentContext = this.highlightableContexts[index];
    const currentParentOption = this.parentOptionMap.get(index);
    const currentParentIsClosed =
      ParentPickerTypes.has(currentContext.type) &&
      currentParentOption &&
      !this.uiState.isExpanded(currentParentOption.context.key);

    if (currentParentIsClosed) {
      let validContext = this.highlightableContexts[nextIndex];
      while (!ParentPickerTypes.has(validContext.type)) {
        nextIndex = (nextIndex + 1) % this.highlightableContexts.length;
        validContext = this.highlightableContexts[nextIndex];
      }
    }

    return nextIndex;
  }

  public prevValidIndex(index: number) {
    const prevIndex =
      (index - 1 + this.highlightableContexts.length) %
      this.highlightableContexts.length;

    const prevContext = this.highlightableContexts[prevIndex];
    if (ParentPickerTypes.has(prevContext.type)) return prevIndex;

    const prevParentIndex = this.closestParentIndex(prevContext, prevIndex);
    const prevParentOption = this.parentOptionMap.get(prevParentIndex);
    const prevParentIsClosed =
      prevParentOption &&
      !this.uiState.isExpanded(prevParentOption.context.key);
    if (prevParentIsClosed) {
      return prevParentIndex;
    }

    return prevIndex;
  }
}
