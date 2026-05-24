import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import type { DashboardMutables } from "./types";

type Taggable = { name?: string; tags?: string[] };

function namesInTag(items: Taggable[], tagName: string): string[] {
  return items
    .filter((item) => item.name && item.tags?.includes(tagName))
    .map((item) => item.name!);
}

function orderedByList(names: string[], orderedAllNames: string[]): string[] {
  const set = new Set(names);
  return orderedAllNames.filter((n) => set.has(n));
}

function clampAtLeastOne(
  next: string[],
  current: string[],
  orderedAllNames: string[],
): string[] {
  if (next.length > 0) return next;
  // Preserve the first currently visible item if possible, otherwise the first overall.
  const fallback = current[0] ?? orderedAllNames[0];
  return fallback ? [fallback] : [];
}

export const showAllInDimensionTag = (
  { dashboard }: DashboardMutables,
  allDimensions: MetricsViewSpecDimension[],
  tagName: string,
) => {
  const orderedAllNames = allDimensions
    .map((d) => d.name)
    .filter((n): n is string => !!n);
  const inTag = new Set(namesInTag(allDimensions, tagName));
  const union = new Set([...dashboard.visibleDimensions, ...inTag]);
  dashboard.visibleDimensions = orderedByList(
    Array.from(union),
    orderedAllNames,
  );
  dashboard.allDimensionsVisible =
    dashboard.visibleDimensions.length === orderedAllNames.length;
};

export const hideAllInDimensionTag = (
  { dashboard }: DashboardMutables,
  allDimensions: MetricsViewSpecDimension[],
  tagName: string,
) => {
  const orderedAllNames = allDimensions
    .map((d) => d.name)
    .filter((n): n is string => !!n);
  const inTag = new Set(namesInTag(allDimensions, tagName));
  const remaining = dashboard.visibleDimensions.filter((n) => !inTag.has(n));
  dashboard.visibleDimensions = clampAtLeastOne(
    orderedByList(remaining, orderedAllNames),
    dashboard.visibleDimensions,
    orderedAllNames,
  );
  dashboard.allDimensionsVisible =
    dashboard.visibleDimensions.length === orderedAllNames.length;
};

export const onlyShowDimensionTag = (
  { dashboard }: DashboardMutables,
  allDimensions: MetricsViewSpecDimension[],
  tagName: string,
) => {
  const orderedAllNames = allDimensions
    .map((d) => d.name)
    .filter((n): n is string => !!n);
  const inTagNames = namesInTag(allDimensions, tagName);
  dashboard.visibleDimensions = clampAtLeastOne(
    orderedByList(inTagNames, orderedAllNames),
    dashboard.visibleDimensions,
    orderedAllNames,
  );
  dashboard.allDimensionsVisible =
    dashboard.visibleDimensions.length === orderedAllNames.length;
};

export const showAllInMeasureTag = (
  { dashboard }: DashboardMutables,
  allMeasures: MetricsViewSpecMeasure[],
  tagName: string,
) => {
  const orderedAllNames = allMeasures
    .map((m) => m.name)
    .filter((n): n is string => !!n);
  const inTag = new Set(namesInTag(allMeasures, tagName));
  const union = new Set([...dashboard.visibleMeasures, ...inTag]);
  dashboard.visibleMeasures = orderedByList(Array.from(union), orderedAllNames);
  dashboard.allMeasuresVisible =
    dashboard.visibleMeasures.length === orderedAllNames.length;
};

export const hideAllInMeasureTag = (
  { dashboard }: DashboardMutables,
  allMeasures: MetricsViewSpecMeasure[],
  tagName: string,
) => {
  const orderedAllNames = allMeasures
    .map((m) => m.name)
    .filter((n): n is string => !!n);
  const inTag = new Set(namesInTag(allMeasures, tagName));
  const remaining = dashboard.visibleMeasures.filter((n) => !inTag.has(n));
  dashboard.visibleMeasures = clampAtLeastOne(
    orderedByList(remaining, orderedAllNames),
    dashboard.visibleMeasures,
    orderedAllNames,
  );
  // Keep leaderboard sort measure valid if it was hidden.
  if (
    dashboard.visibleMeasures.length > 0 &&
    !dashboard.visibleMeasures.includes(dashboard.leaderboardSortByMeasureName)
  ) {
    dashboard.leaderboardSortByMeasureName = dashboard.visibleMeasures[0];
  }
  dashboard.allMeasuresVisible =
    dashboard.visibleMeasures.length === orderedAllNames.length;
};

export const onlyShowMeasureTag = (
  { dashboard }: DashboardMutables,
  allMeasures: MetricsViewSpecMeasure[],
  tagName: string,
) => {
  const orderedAllNames = allMeasures
    .map((m) => m.name)
    .filter((n): n is string => !!n);
  const inTagNames = namesInTag(allMeasures, tagName);
  dashboard.visibleMeasures = clampAtLeastOne(
    orderedByList(inTagNames, orderedAllNames),
    dashboard.visibleMeasures,
    orderedAllNames,
  );
  if (
    dashboard.visibleMeasures.length > 0 &&
    !dashboard.visibleMeasures.includes(dashboard.leaderboardSortByMeasureName)
  ) {
    dashboard.leaderboardSortByMeasureName = dashboard.visibleMeasures[0];
  }
  dashboard.allMeasuresVisible =
    dashboard.visibleMeasures.length === orderedAllNames.length;
};

export const tagActions = {
  showAllInDimensionTag,
  hideAllInDimensionTag,
  onlyShowDimensionTag,
  showAllInMeasureTag,
  hideAllInMeasureTag,
  onlyShowMeasureTag,
};
