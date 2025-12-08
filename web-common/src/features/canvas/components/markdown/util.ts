import { defaultMarkdownAlignment } from "@rilldata/web-common/features/canvas/components/markdown";
import type { ComponentAlignment } from "@rilldata/web-common/features/canvas/components/types";
import type { MarkdownCanvasComponent } from "@rilldata/web-common/features/canvas/components/markdown";
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
import { getQueryServiceResolveTemplatedStringQueryOptions } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived } from "svelte/store";

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

  const formatPattern = /__RILL__FORMAT__\(([^)]+)\)/g;

  return text.replace(formatPattern, (fullMatch, tokenContent: string) => {
    try {
      const trimmed = tokenContent.trim();
      const {
        metrics_view,
        field,
        value,
      }: { metrics_view: string; field: string; value: number | string } =
        JSON.parse(trimmed) as {
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

function buildRequestBody(params: {
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

export function getResolveTemplatedStringQueryOptions(
  component: MarkdownCanvasComponent,
) {
  return derived(
    [
      component.specStore,
      component.timeAndFilterStore,
      component.parent?.filters?.whereFilter ?? null,
      component.parent?.filters?.dimensionThresholdFilters ?? null,
      component.parent?.specStore ?? null,
      runtime,
    ],
    ([
      spec,
      timeAndFilters,
      parentWhereFilter,
      parentDimensionThresholdFilters,
      parentSpec,
      runtimeState,
    ]) => {
      const content = spec?.content ?? "";
      const applyFormatting = spec?.apply_formatting === true;
      const needsTemplating = hasTemplatingSyntax(content);
      const instanceId = runtimeState?.instanceId ?? "";

      const globalWhereFilter = parentWhereFilter ?? undefined;
      const globalDimensionThresholdFilters =
        parentDimensionThresholdFilters ?? [];
      const metricsViews =
        parentSpec?.data?.metricsViews ?? {};

      const requestBody = buildRequestBody({
        content,
        applyFormatting,
        timeRange: timeAndFilters?.timeRange,
        globalWhereFilter,
        globalDimensionThresholdFilters,
        metricsViews,
      });

      const enabled =
        !!needsTemplating &&
        !!content &&
        !!instanceId &&
        !!requestBody &&
        !!requestBody.additionalTimeRange;

      // Always return query options, but use enabled to control execution
      // When disabled, the query won't execute, so we use a minimal body (won't be used)
      // When enabled, we use the actual requestBody
      const body: QueryServiceResolveTemplatedStringBody = (!enabled || !requestBody)
        ? { body: content, useFormatTokens: applyFormatting }
        : requestBody;
      
      const queryEnabled = enabled && !!requestBody;
      
      // TypeScript can't fully infer generic return type in derived callback,
      // but the runtime type is correct. The type assertion helps TypeScript
      // understand the return type matches the function signature.
      return getQueryServiceResolveTemplatedStringQueryOptions(
        instanceId,
        body,
        {
          query: {
            enabled: queryEnabled,
          },
        },
      );
    },
  );
}
