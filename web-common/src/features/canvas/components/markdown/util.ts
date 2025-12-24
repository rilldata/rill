import { defaultMarkdownAlignment } from "@rilldata/web-common/features/canvas/components/markdown";
import type { ComponentAlignment } from "@rilldata/web-common/features/canvas/components/types";
import type { MarkdownCanvasComponent } from "@rilldata/web-common/features/canvas/components/markdown";
import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import type {
  MetricsViewSpecMeasure,
  V1Expression,
  V1MetricsView,
  V1TimeRange,
  QueryServiceResolveTemplatedStringBody,
} from "@rilldata/web-common/runtime-client";
import { getQueryServiceResolveTemplatedStringQueryOptions } from "@rilldata/web-common/runtime-client";
import { derived } from "svelte/store";
import type { ParsedFilters } from "../../stores/filter-state";
import type { Readable } from "svelte/motion";

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
      }: {
        metrics_view: string;
        field: string;
        value: number | string | null;
      } = JSON.parse(trimmed) as {
        metrics_view: string;
        field: string;
        value: number | string | null;
      };

      const metricsView = metricsViews[metrics_view];
      const metricsViewSpec = metricsView?.state?.validSpec;
      if (!metricsViewSpec) return fullMatch;

      const measure = metricsViewSpec.measures?.find(
        (m: MetricsViewSpecMeasure) => m.name === field,
      );
      if (!measure) return fullMatch;

      if (value === null) {
        const formatter = createMeasureValueFormatter<null>(measure);
        const formatted = formatter(null);
        return formatted ?? "";
      }

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
  parsedMetricsViewFilters: ParsedFilters[];
  metricsViews: Record<string, V1MetricsView | undefined>;
}): QueryServiceResolveTemplatedStringBody | null {
  const { content, applyFormatting, timeRange, parsedMetricsViewFilters } =
    params;

  const additionalWhereByMetricsView: Record<string, V1Expression> = {};

  parsedMetricsViewFilters.forEach((f) => {
    additionalWhereByMetricsView[f.metricsViewName] = f.where;
  });

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
): Readable<
  ReturnType<typeof getQueryServiceResolveTemplatedStringQueryOptions>
> {
  return derived(
    [component.parent.filterManager.metricsViewFilters],
    ([metricsViewFilters], set) => {
      derived(
        [
          component.specStore,
          component.timeAndFilterStore,
          component.parent?.specStore ?? null,

          component.parent.timeManager.hasTimeSeriesStore,
          ...Array.from(metricsViewFilters.values()).map((f) => f.parsed),
        ],
        ([
          spec,
          timeAndFilters,
          parentSpec,

          hasTimeSeries,
          ...parsedMetricsViewFilters
        ]) => {
          const content = spec?.content ?? "";
          const applyFormatting = spec?.apply_formatting === true;
          const needsTemplating = hasTemplatingSyntax(content);
          const instanceId = component.parent.instanceId;

          const metricsViews = parentSpec?.data?.metricsViews ?? {};

          const requestBody = buildRequestBody({
            content,
            applyFormatting,
            timeRange: timeAndFilters?.timeRange,
            parsedMetricsViewFilters,
            metricsViews,
          });

          const enabled =
            !!needsTemplating &&
            !!content &&
            !!instanceId &&
            !!requestBody &&
            (!hasTimeSeries || !!timeAndFilters?.timeRange);

          const body: QueryServiceResolveTemplatedStringBody =
            !enabled || !requestBody
              ? { body: content, useFormatTokens: applyFormatting }
              : requestBody;

          const queryEnabled = enabled && !!requestBody;

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
      ).subscribe(set);
    },
  );
}
