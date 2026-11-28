import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import {
  AllMeasureFilterOperationOptions,
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter";
import { get } from "svelte/store";

export interface WorkbookHeader {
  title: string;
  timeRangeText: string;
  comparisonText: string;
  timeZone: string;
  filterLines: string[];
}

// How many filter values a line spells out before summarizing the rest.
const MAX_LISTED_VALUES = 20;

// The header of every dashboard sheet: the title, the exact resolved time
// range and comparison (never the relative alias), the time zone and one line
// per active filter. Mirrors what CanvasPdfExportHeader renders for the PDF.
export function buildWorkbookHeader(
  canvasEntity: CanvasEntity,
  fallbackTitle: string,
  includeFilters: boolean,
): WorkbookHeader {
  const { interval, timeGrain, timeZone, showComparison, comparisonInterval } =
    get(canvasEntity.timeFilterManager.getTimeControlStore());

  return {
    title: get(canvasEntity.titleStore) || fallbackTitle,
    timeRangeText: interval ? prettyFormatTimeRange(interval, timeGrain) : "",
    comparisonText:
      showComparison && comparisonInterval
        ? prettyFormatTimeRange(comparisonInterval, timeGrain)
        : "",
    timeZone: timeZone ?? "",
    filterLines: includeFilters
      ? describeFilters(canvasEntity.expressionFilterManager)
      : [],
  };
}

// One line per active filter, e.g. "Country: US, CA" or "Impressions > 1000".
export function describeFilters(manager: ExpressionFilterManager): string[] {
  const lines: string[] = [];
  const { dimensions, measures } = manager.sortedFilterManagers;

  for (const dimension of dimensions) {
    if (!dimension.expr) continue;
    const prefix = dimension.exclude ? "excludes " : "";
    switch (dimension.mode) {
      case DimensionFilterMode.Contains:
        lines.push(
          `${dimension.label}: ${prefix}contains "${dimension.inputText}"`,
        );
        break;
      default:
        lines.push(
          `${dimension.label}: ${prefix}${listValues(dimension.selectedValues)}`,
        );
    }
  }

  for (const measure of measures) {
    if (!measure.expr) continue;
    const operation = AllMeasureFilterOperationOptions.find(
      (option) => option.value === measure.operation,
    )?.shortLabel;
    let criteria: string;
    switch (measure.operation) {
      case MeasureFilterOperation.Between:
        criteria = `between ${measure.value1} and ${measure.value2}`;
        break;
      case MeasureFilterOperation.NotBetween:
        criteria = `not between ${measure.value1} and ${measure.value2}`;
        break;
      default:
        criteria = `${operation ?? measure.operation} ${measure.value1}`;
        if (measure.type === MeasureFilterType.PercentChange) criteria += "%";
    }
    const subject =
      measure.type === MeasureFilterType.AbsoluteChange
        ? `${measure.label} change`
        : measure.type === MeasureFilterType.PercentChange
          ? `${measure.label} % change`
          : measure.type === MeasureFilterType.PercentOfTotal
            ? `${measure.label} % of total`
            : measure.label;
    lines.push(`${subject} ${criteria}`);
  }

  return lines;
}

function listValues(values: string[]): string {
  if (values.length <= MAX_LISTED_VALUES) return values.join(", ");
  const rest = values.length - MAX_LISTED_VALUES;
  return `${values.slice(0, MAX_LISTED_VALUES).join(", ")} and ${rest} more`;
}
