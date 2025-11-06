import { getCanvasQueryOptions } from "@rilldata/web-common/features/canvas/selector.ts";
import {
  getValidCanvasSpecsQueryOptions,
  getValidExploreSpecsQueryOptions,
  getValidMetricsViewSpecsQueryOptions,
} from "@rilldata/web-common/features/dashboards/selectors.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export type ReportSource = {
  label: string;
  kind: ResourceKind;
  metricsViewName: string;
  exploreName: string;
  canvasName: string;
};

export function getPrimaryExploreSourceOptions(): Readable<ReportSource[]> {
  const exploreSpecsQuery = createQuery(getValidExploreSpecsQueryOptions());
  const metricsViewSpecsQuery = createQuery(
    getValidMetricsViewSpecsQueryOptions(),
  );

  return derived(
    [exploreSpecsQuery, metricsViewSpecsQuery],
    ([validExploreSpecsResp, validMetricsViewSpecsResp]) => {
      const exploreNames = validExploreSpecsResp.data ?? [];
      const metricsViewsMap = new Map(
        validMetricsViewSpecsResp.data?.map((mv) => [
          mv.meta?.name?.name ?? "",
          mv.metricsView?.state?.validSpec ?? {},
        ]) ?? [],
      );

      const exploreSources = exploreNames
        .map((res) => {
          const exploreName = res.meta?.name?.name ?? "";
          const exploreSpec = res.explore?.state?.validSpec;
          const metricsViewName = exploreSpec?.metricsView ?? "";
          const metricsViewSpec = metricsViewsMap.get(metricsViewName);
          if (!metricsViewSpec?.timeDimension) return undefined;

          return {
            label: exploreSpec?.displayName ?? exploreName,
            kind: ResourceKind.Explore,
            metricsViewName,
            exploreName,
            canvasName: "",
          };
        })
        .filter(Boolean) as ReportSource[];

      return exploreSources;
    },
  );
}

export function getPrimaryCanvasSourceOptions(): Readable<ReportSource[]> {
  const canvasSpecsQuery = createQuery(getValidCanvasSpecsQueryOptions());

  return derived(canvasSpecsQuery, (validCanvasSpecs) => {
    const canvasNames = validCanvasSpecs.data ?? [];
    const canvasSources = canvasNames.map((res) => {
      const canvasName = res.meta?.name?.name ?? "";
      const canvasSpec = res.canvas?.state?.validSpec;
      return {
        label: canvasSpec?.displayName ?? canvasName,
        kind: ResourceKind.Canvas,
        metricsViewName: "", // There can be multiple of these. Will be added as sub options
        exploreName: "",
        canvasName,
      };
    });

    return canvasSources;
  });
}

export function getSecondarySourceForCanvasOptions(
  canvasNameStore: Readable<string>,
): Readable<ReportSource[]> {
  const canvasResolvedSpecQuery = createQuery(
    getCanvasQueryOptions(canvasNameStore),
  );

  return derived(
    [canvasNameStore, canvasResolvedSpecQuery],
    ([canvasName, canvasResolvedSpecResp]) => {
      return Object.entries(canvasResolvedSpecResp.data?.metricsViews ?? [])
        .map(([metricsViewName, metricsViewResource]) => {
          if (!metricsViewResource?.state?.validSpec?.timeDimension) return;
          return {
            label:
              metricsViewResource?.state?.validSpec?.displayName ??
              metricsViewName,
            kind: ResourceKind.MetricsView,
            metricsViewName,
            exploreName: "",
            canvasName,
          };
        })
        .filter(Boolean) as ReportSource[];
    },
  );
}
