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
import {
  getQueryServiceResolveTemplatedStringQueryKey,
  getQueryServiceResolveTemplatedStringQueryOptions,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived } from "svelte/store";
import type { Readable } from "svelte/store";

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
  component: string;
  applyFormatting: boolean;
  timeRange: V1TimeRange | undefined;
  exprByMetricsView: Record<string, V1Expression>;
  metricsViews: Record<string, V1MetricsView | undefined>;
}): QueryServiceResolveTemplatedStringBody | null {
  const { component, applyFormatting, timeRange, exprByMetricsView } = params;

  return {
    // The runtime resolves the component's saved content rather than a body sent by the browser.
    // Only content declared in the project may use project variables, so sending the content as a body would reject it.
    component,
    useFormatTokens: applyFormatting,
    additionalTimeRange: timeRange,
    ...(Object.keys(exprByMetricsView).length > 0 && {
      additionalWhereByMetricsView: exprByMetricsView,
    }),
  };
}

export function getResolveTemplatedStringQueryOptions(
  component: MarkdownCanvasComponent,
  client: RuntimeClient,
): Readable<
  ReturnType<typeof getQueryServiceResolveTemplatedStringQueryOptions>
> {
  return derived(
    [component.parent.expressionFilterManager.exprByMetricsViewStore],
    ([exprByMetricsView], set) => {
      return derived(
        [
          component.specStore,
          component.resource,
          component.timeAndFilterStore,
          component.parent?.specStore ?? null,
          component.parent.timeManager.hasTimeSeriesStore,
        ],
        ([spec, resource, timeAndFilters, parentSpec, hasTimeSeries]) => {
          const content = spec?.content ?? "";
          const applyFormatting = spec?.apply_formatting === true;
          const needsTemplating = hasTemplatingSyntax(content);

          const metricsViews = parentSpec?.data?.metricsViews ?? {};

          const requestBody = buildRequestBody({
            component: component.id,
            applyFormatting,
            timeRange: timeAndFilters?.timeRange,
            exprByMetricsView,
            metricsViews,
          });

          const enabled =
            !!needsTemplating &&
            !!content &&
            !!client.instanceId &&
            !!requestBody &&
            (!hasTimeSeries || !!timeAndFilters?.timeRange);

          const body: QueryServiceResolveTemplatedStringBody =
            !enabled || !requestBody
              ? { component: component.id, useFormatTokens: applyFormatting }
              : requestBody;

          const queryEnabled = enabled && !!requestBody;

          // Cast needed: QueryServiceResolveTemplatedStringBody uses string
          // timestamps while the v2 proto request expects Timestamp objects;
          // the v2 fromJson layer handles the conversion at runtime.
          // The request names the component instead of carrying its content, and the runtime resolves the content
          // it has reconciled. The resource's state version is added to the query key so the text is refetched once a
          // reconcile completes; keying on the local content instead would cache the previous text under the new key
          // when the request races the reconcile.
          return getQueryServiceResolveTemplatedStringQueryOptions(
            client,
            body as any,
            {
              query: {
                enabled: queryEnabled,
                queryKey: [
                  ...getQueryServiceResolveTemplatedStringQueryKey(
                    client.instanceId,
                    body as any,
                  ),
                  resource?.meta?.stateVersion ?? "",
                ],
              },
            },
          );
        },
      ).subscribe(set);
    },
  );
}
