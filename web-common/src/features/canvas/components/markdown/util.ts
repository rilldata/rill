import { defaultMarkdownAlignment } from "@rilldata/web-common/features/canvas/components/markdown";
import type { ComponentAlignment } from "@rilldata/web-common/features/canvas/components/types";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  buildValidMetricsViewFilter,
  createAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
  V1Expression,
  V1MetricsView,
  V1TimeRange,
  QueryServiceResolveTemplatedStringBody,
} from "@rilldata/web-common/runtime-client";

export function getPositionClasses(alignment: ComponentAlignment | undefined) {
  if (!alignment) alignment = defaultMarkdownAlignment;
  let classString = "";

  switch (alignment.horizontal) {
    case "left":
      classString = "items-start";
      break;
    case "center":
      classString = "items-center";
      break;
    case "right":
      classString = "items-end";
  }

  switch (alignment.vertical) {
    case "top":
      classString += " justify-start";
      break;
    case "middle":
      classString += " justify-center";
      break;
    case "bottom":
      classString += " justify-end";
  }

  return classString;
}

export function hasTemplatingSyntax(content: string): boolean {
  return /\{\{[\s\S]*?\}\}/.test(content);
}

export function formatResolvedContent(
  text: string,
  metricsViews: Record<string, V1MetricsView | undefined>,
): string {
  if (!text) return text;

  const formatPattern = /__RILL__FORMAT__\((\{[^}]+\})\)/g;

  return text.replace(formatPattern, (fullMatch, jsonStr: string) => {
    try {
      const {
        metrics_view,
        field,
        value,
      }: { metrics_view: string; field: string; value: number | string } =
        JSON.parse(jsonStr) as {
          metrics_view: string;
          field: string;
          value: number | string;
        };

      const metricsView = metricsViews[metrics_view];
      const metricsViewSpec = metricsView?.state?.validSpec;
      if (!metricsViewSpec) return fullMatch;

      const measure = metricsViewSpec.measures?.find(
        (m: MetricsViewSpecMeasure) => m.name === field,
      );
      if (!measure) return fullMatch;

      const numericValue =
        typeof value === "number" ? value : parseFloat(String(value));
      if (isNaN(numericValue)) return fullMatch;

      const formatter = createMeasureValueFormatter(measure);
      return formatter(numericValue);
    } catch {
      return fullMatch;
    }
  });
}

export function buildRequestBody(params: {
  content: string;
  applyFormatting: boolean;
  timeRange: V1TimeRange | undefined;
  globalWhereFilter: V1Expression | undefined;
  globalDimensionThresholdFilters: DimensionThresholdFilter[];
  metricsViews: Record<string, V1MetricsView | undefined>;
}): QueryServiceResolveTemplatedStringBody | null {
  const {
    content,
    applyFormatting,
    timeRange,
    globalWhereFilter,
    globalDimensionThresholdFilters,
    metricsViews,
  } = params;

  if (!timeRange?.start || !timeRange?.end) return null;

  const additionalWhereByMetricsView: Record<string, V1Expression> = {};
  const metricsViewNames = Object.keys(metricsViews);

  if (
    metricsViewNames.length > 0 &&
    (globalWhereFilter || globalDimensionThresholdFilters.length > 0)
  ) {
    for (const metricsViewName of metricsViewNames) {
      const metricsView = metricsViews[metricsViewName];
      const metricsViewSpec = metricsView?.state?.validSpec;
      const dimensions: MetricsViewSpecDimension[] =
        metricsViewSpec?.dimensions ?? [];
      const measures: MetricsViewSpecMeasure[] =
        metricsViewSpec?.measures ?? [];

      const whereFilter = buildValidMetricsViewFilter(
        globalWhereFilter || createAndExpression([]),
        globalDimensionThresholdFilters,
        dimensions,
        measures,
      );

      if (
        whereFilter &&
        whereFilter.cond?.exprs &&
        whereFilter.cond.exprs.length > 0
      ) {
        additionalWhereByMetricsView[metricsViewName] = whereFilter;
      } else if (
        !whereFilter &&
        globalWhereFilter &&
        globalWhereFilter.cond?.exprs &&
        globalWhereFilter.cond.exprs.length > 0
      ) {
        additionalWhereByMetricsView[metricsViewName] = globalWhereFilter;
      }
    }
  }

  return {
    body: content,
    useFormatTokens: applyFormatting,
    additionalTimeRange: timeRange,
    ...(Object.keys(additionalWhereByMetricsView).length > 0 && {
      additionalWhereByMetricsView,
    }),
  };
}
