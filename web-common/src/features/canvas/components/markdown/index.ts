import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import {
  type CanvasComponentType,
  type ComponentAlignment,
  type ComponentCommonProperties,
  type ComponentFilterProperties,
} from "../types";
import { getFilterOptions } from "../util";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import MarkdownCanvas from "./MarkdownCanvas.svelte";
import { derived } from "svelte/store";
import { getFiltersFromText } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-search-text-utils";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";

export { default as Markdown } from "./Markdown.svelte";

export const defaultMarkdownAlignment: ComponentAlignment = {
  vertical: "middle",
  horizontal: "left",
};

export interface MarkdownSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  content: string;
  alignment?: ComponentAlignment;
}

export class MarkdownCanvasComponent extends BaseCanvasComponent<MarkdownSpec> {
  minSize = { width: 1, height: 1 };
  defaultSize = { width: 3, height: 2 };
  resetParams = [];
  type: CanvasComponentType = "markdown";
  component = MarkdownCanvas;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: MarkdownSpec = {
      title: "",
      description: "",
      content: "Your text",
      alignment: defaultMarkdownAlignment,
    };
    super(resource, parent, path, defaultSpec);
  }

  // Override timeAndFilterStore to bypass dimension filter validation
  // Markdown can query multiple metrics views, so we can't validate against a single set of dimensions
  get timeAndFilterStore() {
    return derived(
      [
        this.parent.timeControls.timeRangeStateStore,
        this.localTimeControls.timeRangeStateStore,
        this.parent.timeControls.comparisonRangeStateStore,
        this.localTimeControls.comparisonRangeStateStore,
        this.parent.timeControls.selectedTimezone,
        this.parent.filters.whereFilter,
        this.parent.filters.dimensionThresholdFilters,
        this.parent.specStore,
        this.parent.timeControls.hasTimeSeries,
        this.specStore,
      ],
      ([
        globalTimeRangeState,
        localTimeRangeState,
        globalComparisonRangeState,
        localComparisonRangeState,
        timeZone,
        whereFilter,
        dtf,
        canvasData,
        hasTimeSeries,
        componentSpec,
      ]) => {
        // Time filters
        let timeRange = {
          start: globalTimeRangeState?.timeStart,
          end: globalTimeRangeState?.timeEnd,
          timeZone,
        };
        let showTimeComparison = !!globalComparisonRangeState?.comparisonTimeStart;
        let timeGrain = globalTimeRangeState?.selectedTimeRange?.interval;
        let comparisonTimeRange = {
          start: globalComparisonRangeState?.comparisonTimeStart,
          end: globalComparisonRangeState?.comparisonTimeEnd,
          timeZone,
        };
        let timeRangeState = globalTimeRangeState;
        let comparisonTimeRangeState = globalComparisonRangeState;

        if (componentSpec?.["time_filters"]) {
          timeRange = {
            start: localTimeRangeState?.timeStart,
            end: localTimeRangeState?.timeEnd,
            timeZone,
          };
          comparisonTimeRange = {
            start: localComparisonRangeState?.comparisonTimeStart,
            end: localComparisonRangeState?.comparisonTimeEnd,
            timeZone,
          };
          showTimeComparison = !!localComparisonRangeState?.comparisonTimeStart;
          timeGrain = localTimeRangeState?.selectedTimeRange?.interval;
          timeRangeState = localTimeRangeState;
          comparisonTimeRangeState = localComparisonRangeState;
        }

        // Dimension Filters - SKIP VALIDATION, pass through as-is
        const globalWhere = whereFilter ?? createAndExpression([]);
        let where = globalWhere;

        if (componentSpec?.["dimension_filters"]) {
          const { expr: componentWhere } = getFiltersFromText(
            componentSpec?.["dimension_filters"],
          );
          where = mergeFilters(globalWhere, componentWhere);
        }

        return {
          timeRange,
          showTimeComparison,
          comparisonTimeRange,
          where,
          timeGrain,
          timeRangeState,
          comparisonTimeRangeState,
          hasTimeSeries,
        };
      },
    );
  }

  isValid(spec: MarkdownSpec): boolean {
    return typeof spec.content === "string" && spec.content.trim().length > 0;
  }

  inputParams(): InputParams<MarkdownSpec> {
    return {
      options: {
        content: {
          type: "textArea",
          label: "Markdown",
          description:
            'Write markdown with Go templates. Use {{ metrics_sql "select..." }} to query metrics. Supports Sprig functions for data transformation.',
        },
        alignment: {
          type: "alignment",
          label: "Alignment",
          meta: {
            defaultAlignment: defaultMarkdownAlignment,
          },
        },
      },
      filter: getFilterOptions(false, false),
    };
  }

  static newComponentSpec(): MarkdownSpec {
    const defaultContent = String.raw`# 📊 My Dashboard

Write **markdown** with _formatting_ and access to data from metrics.

Use Go templates with the \`metrics_sql\` function to query your data.`;

    return {
      content: defaultContent,
      alignment: defaultMarkdownAlignment,
    };
  }
}
