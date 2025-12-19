import { describe, it, expect } from "vitest";
import { PickerOptionsHighlightManager } from "@rilldata/web-common/features/chat/core/context/picker/highlight-manager.ts";
import type {
  InlineContextPickerParentOption,
  InlineContextPickerSection,
} from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import {
  createLoadingContext,
  type InlineContext,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import {
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_THREE_DIMENSIONS,
  AD_BIDS_THREE_MEASURES,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import { get, writable } from "svelte/store";

describe("PickerOptionsHighlightManager", () => {
  it("should highlight the child after loading ends", () => {
    const manager = new PickerOptionsHighlightManager();
    manager.filterOptionsUpdated(
      [
        createSection(AD_BIDS_MV_OPTIONS),
        createSection(AD_BIDS_LOADING_OPTIONS),
      ],
      null,
    );
    const loadingContext = AD_BIDS_LOADING_OPTIONS.children[0].options[0];
    manager.highlightContext(loadingContext);
    expect(get(manager.highlightedContext)).toEqual(loadingContext);

    manager.filterOptionsUpdated(
      [createSection(AD_BIDS_MV_OPTIONS), createSection(AD_BIDS_MODEL_OPTIONS)],
      null,
    );
    const firstColumnContext = AD_BIDS_MODEL_OPTIONS.children[0].options[0];
    expect(get(manager.highlightedContext)).toEqual(firstColumnContext);
  });
});

const AD_BIDS_MV_OPTIONS = {
  context: {
    type: InlineContextType.MetricsView,
    metricsView: AD_BIDS_METRICS_NAME,
    value: AD_BIDS_METRICS_NAME,
  },
  openStore: writable(false),
  children: [
    {
      type: InlineContextType.Measure,
      options: AD_BIDS_THREE_MEASURES.map((m) => ({
        type: InlineContextType.Measure,
        value: m.name!,
        measure: m.name,
        metricsView: AD_BIDS_METRICS_NAME,
      })),
    },
    {
      type: InlineContextType.Dimension,
      options: AD_BIDS_THREE_DIMENSIONS.map((d) => ({
        type: InlineContextType.Dimension,
        value: d.name!,
        dimension: d.name,
        metricsView: AD_BIDS_METRICS_NAME,
      })),
    },
  ],
} satisfies InlineContextPickerParentOption;

const AD_BIDS_CONTEXT = {
  type: InlineContextType.Model,
  value: AD_BIDS_NAME,
  model: AD_BIDS_NAME,
} satisfies InlineContext;
const AD_BIDS_LOADING_OPTIONS = {
  context: AD_BIDS_CONTEXT,
  openStore: writable(false),
  children: [
    {
      type: InlineContextType.Loading,
      options: [createLoadingContext(AD_BIDS_CONTEXT)],
    },
  ],
} satisfies InlineContextPickerParentOption;
const AD_BIDS_MODEL_OPTIONS = {
  context: AD_BIDS_CONTEXT,
  openStore: writable(false),
  children: [
    {
      type: InlineContextType.Column,
      options: [
        {
          type: InlineContextType.Column,
          value: AD_BIDS_PUBLISHER_DIMENSION,
          column: AD_BIDS_PUBLISHER_DIMENSION,
        },
        {
          type: InlineContextType.Column,
          value: AD_BIDS_DOMAIN_DIMENSION,
          column: AD_BIDS_DOMAIN_DIMENSION,
        },
      ],
    },
  ],
} satisfies InlineContextPickerParentOption;

function createSection(parentOption: InlineContextPickerParentOption) {
  return {
    type: parentOption.context.type,
    options: [parentOption],
  } satisfies InlineContextPickerSection;
}
