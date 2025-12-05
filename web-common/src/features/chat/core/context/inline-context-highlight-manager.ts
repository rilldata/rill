import {
  InlineContextType,
  type InlineContext,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import type { MetricsViewContextOption } from "@rilldata/web-common/features/chat/core/context/inline-context-data.ts";
import { inlineContextsAreEqual } from "web-common/src/features/chat/core/context/inline-context.ts";
import { writable } from "svelte/store";

export class InlineContextHighlightManager {
  public highlightedContext = writable<InlineContext | null>(null);

  private highlightableContexts: InlineContext[] = [];
  private highlightedIndex = -1;

  public filterOptionsUpdated(filteredOptions: MetricsViewContextOption[]) {
    // Convert the filtered options to a flat list for ease of navigation using arrow keys.
    this.highlightableContexts = filteredOptions.flatMap((option) => [
      option.metricsViewContext,
      ...option.measures,
      ...option.dimensions,
    ]);

    // Prefer non-metrics context if available for the 1st metrics view.
    const nonMetricsViewAvailable =
      this.highlightableContexts.length > 1 &&
      this.highlightableContexts[1].type !== InlineContextType.MetricsView;
    this.highlightedIndex = nonMetricsViewAvailable ? 1 : 0;
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
