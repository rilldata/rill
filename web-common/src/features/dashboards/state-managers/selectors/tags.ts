import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import { allDimensions } from "./dimensions";
import { allMeasures } from "./measures";
import type { DashboardDataSources } from "./types";

export type DimensionTag = {
  name: string;
  displayName: string;
  totalCount: number;
};

export type TagVisibilityState = {
  tagName: string;
  visibleCount: number;
  totalCount: number;
  state: "none" | "partial" | "all";
};

function tagsFromItems(
  items: { tags?: string[] }[],
): { name: string; total: number }[] {
  const seen = new Map<string, number>();
  for (const item of items) {
    if (!item.tags) continue;
    for (const tag of item.tags) {
      if (!tag) continue;
      seen.set(tag, (seen.get(tag) ?? 0) + 1);
    }
  }
  return Array.from(seen, ([name, total]) => ({ name, total }));
}

function deriveTagState(
  visibleCount: number,
  totalCount: number,
): TagVisibilityState["state"] {
  if (visibleCount === 0) return "none";
  if (visibleCount === totalCount) return "all";
  return "partial";
}

export const dimensionTags = (
  dashData: DashboardDataSources,
): DimensionTag[] => {
  const dims = allDimensions(dashData);
  return tagsFromItems(dims).map(({ name, total }) => ({
    name,
    displayName: name,
    totalCount: total,
  }));
};

export const dimensionTagVisibilityState = (
  dashData: DashboardDataSources,
): ((tagName: string) => TagVisibilityState) => {
  const dims = allDimensions(dashData);
  const visible = new Set(dashData.dashboard?.visibleDimensions ?? []);
  return (tagName: string) => {
    let total = 0;
    let visibleCount = 0;
    for (const dim of dims) {
      if (!dim.tags?.includes(tagName)) continue;
      total += 1;
      if (dim.name && visible.has(dim.name)) visibleCount += 1;
    }
    return {
      tagName,
      visibleCount,
      totalCount: total,
      state: deriveTagState(visibleCount, total),
    };
  };
};

export const dimensionsForTag = (
  dashData: DashboardDataSources,
): ((tagName: string) => MetricsViewSpecDimension[]) => {
  const dims = allDimensions(dashData);
  return (tagName: string) => dims.filter((dim) => dim.tags?.includes(tagName));
};

export const measureTags = (dashData: DashboardDataSources): DimensionTag[] => {
  const measures = allMeasures(dashData);
  return tagsFromItems(measures).map(({ name, total }) => ({
    name,
    displayName: name,
    totalCount: total,
  }));
};

export const measureTagVisibilityState = (
  dashData: DashboardDataSources,
): ((tagName: string) => TagVisibilityState) => {
  const measures = allMeasures(dashData);
  const visible = new Set(dashData.dashboard?.visibleMeasures ?? []);
  return (tagName: string) => {
    let total = 0;
    let visibleCount = 0;
    for (const measure of measures) {
      if (!measure.tags?.includes(tagName)) continue;
      total += 1;
      if (measure.name && visible.has(measure.name)) visibleCount += 1;
    }
    return {
      tagName,
      visibleCount,
      totalCount: total,
      state: deriveTagState(visibleCount, total),
    };
  };
};

export const measuresForTag = (
  dashData: DashboardDataSources,
): ((tagName: string) => MetricsViewSpecMeasure[]) => {
  const measures = allMeasures(dashData);
  return (tagName: string) =>
    measures.filter((measure) => measure.tags?.includes(tagName));
};

export const tagSelectors = {
  dimensionTags,
  dimensionTagVisibilityState,
  dimensionsForTag,
  measureTags,
  measureTagVisibilityState,
  measuresForTag,
};
