import { getCanvasQueryOptions } from "@rilldata/web-common/features/canvas/selector.ts";
import {
  getValidCanvasSpecsQueryOptions,
  getValidExploreSpecsQueryOptions,
} from "@rilldata/web-common/features/dashboards/selectors.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export type ReportSource = {
  label: string;
  metricsViewName: string;
  exploreName: string;
  canvasName: string;
};

export function getPrimaryExploreSourceOptions(): Readable<ReportSource[]> {
  const exploreSpecsQuery = createQuery(getValidExploreSpecsQueryOptions());

  return derived(exploreSpecsQuery, (validExploreSpecs) => {
    const exploreNames = validExploreSpecs.data ?? [];
    const exploreSources = exploreNames.map((res) => {
      const exploreName = res.meta?.name?.name ?? "";
      const exploreSpec = res.explore?.state?.validSpec;
      return {
        label: exploreSpec?.displayName ?? exploreName,
        metricsViewName: exploreSpec?.metricsView ?? "",
        exploreName,
        canvasName: "",
      };
    });

    return exploreSources;
  });
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
        metricsViewName: "", // There can be multplie of these. Will be added as sub options
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
      return Object.entries(
        canvasResolvedSpecResp.data?.metricsViews ?? [],
      ).map(([metricsViewName, metricsViewSpec]) => ({
        label:
          metricsViewSpec?.state?.validSpec?.displayName ?? metricsViewName,
        metricsViewName,
        exploreName: "",
        canvasName,
      }));
    },
  );
}
