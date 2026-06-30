import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import {
  UNTAGGED_KEY,
  UNTAGGED_LABEL,
  getResourceTags,
} from "../dashboards/listing/selectors";

// Canvas dashboards sort before explore dashboards, then alphabetically by
// resource name.
export function sortDashboardResources(
  visualizations: V1Resource[],
): V1Resource[] {
  return [...visualizations].sort((a, b) => {
    const aIsCanvas = !!a?.canvas;
    const bIsCanvas = !!b?.canvas;
    if (aIsCanvas !== bIsCanvas) return aIsCanvas ? -1 : 1;
    return (a.meta?.name?.name ?? "").localeCompare(b.meta?.name?.name ?? "");
  });
}

export function getAllDashboardTags(visualizations: V1Resource[]): string[] {
  return Array.from(new Set(visualizations.flatMap(getResourceTags))).sort();
}

export function hasUntaggedDashboards(visualizations: V1Resource[]): boolean {
  return visualizations.some((r) => getResourceTags(r).length === 0);
}

// Buckets dashboards by each of their tags. Multi-tag dashboards appear in
// every bucket they belong to. Untagged dashboards are not added to any bucket.
export function groupDashboardsByTag(
  sortedVisualizations: V1Resource[],
): Map<string, V1Resource[]> {
  const map = new Map<string, V1Resource[]>();
  for (const r of sortedVisualizations) {
    const tags = getResourceTags(r);
    const effectiveTags = tags.length ? tags : [UNTAGGED_KEY];
    for (const tag of effectiveTags) {
      const bucket = map.get(tag) ?? [];
      bucket.push(r);
      map.set(tag, bucket);
    }
  }
  return map;
}

export function buildDashboardHref(
  resource: V1Resource,
  tag: string,
  organization: string,
  project: string,
): string {
  const isMetricsExplorer = !!resource?.explore;
  const slug = isMetricsExplorer ? "explore" : "canvas";
  const name = resource.meta.name.name;
  const base = `/${organization}/${project}/${slug}/${name}`;
  return tag === UNTAGGED_KEY
    ? base
    : `${base}?tags=${encodeURIComponent(tag)}`;
}

export function buildDashboardSubOption(
  resource: V1Resource,
  tag: string,
  organization: string,
  project: string,
): [string, PathOption] {
  const name = resource.meta.name.name;
  const isMetricsExplorer = !!resource?.explore;
  return [
    name.toLowerCase(),
    {
      label:
        (isMetricsExplorer
          ? resource?.explore?.spec?.displayName
          : resource?.canvas?.spec?.displayName) || name,
      href: buildDashboardHref(resource, tag, organization, project),
      resourceKind: isMetricsExplorer
        ? ResourceKind.Explore
        : ResourceKind.Canvas,
    },
  ];
}

// Tag breadcrumb options: each tag shows a submenu of the dashboards that
// carry that tag. UNTAGGED_KEY is included when any dashboard has no tags,
// or when it's the currently active selection (so the breadcrumb renders).
export function buildTagPathsOptions({
  allDashboardTags,
  dashboardsByTag,
  hasUntaggedDashboard,
  activeTag,
  organization,
  project,
}: {
  allDashboardTags: string[];
  dashboardsByTag: Map<string, V1Resource[]>;
  hasUntaggedDashboard: boolean;
  activeTag: string | undefined;
  organization: string;
  project: string;
}): Map<string, PathOption> {
  const map = new Map<string, PathOption>();
  for (const tag of allDashboardTags) {
    const subEntries = (dashboardsByTag.get(tag) ?? []).map((r) =>
      buildDashboardSubOption(r, tag, organization, project),
    );
    map.set(tag.toLowerCase(), {
      label: tag,
      href: `/${organization}/${project}?tags=${encodeURIComponent(tag)}`,
      subOptions: new Map(subEntries),
    });
  }
  if (hasUntaggedDashboard || activeTag === UNTAGGED_KEY) {
    const subEntries = (dashboardsByTag.get(UNTAGGED_KEY) ?? []).map((r) =>
      buildDashboardSubOption(r, UNTAGGED_KEY, organization, project),
    );
    map.set(UNTAGGED_KEY, {
      label: UNTAGGED_LABEL,
      href: `/${organization}/${project}?tags=${encodeURIComponent(UNTAGGED_KEY)}`,
      subOptions: new Map(subEntries),
    });
  }
  return map;
}

export function buildDashboardPathOption(resource: V1Resource): PathOption {
  const name = resource.meta.name.name;
  const isMetricsExplorer = !!resource?.explore;
  return {
    label:
      (isMetricsExplorer
        ? resource?.explore?.spec?.displayName
        : resource?.canvas?.spec?.displayName) || name,
    // depth: 2 ensures path generation always anchors at the project
    // level, even when a tag segment is inserted before this one.
    depth: 2,
    section: isMetricsExplorer ? "explore" : "canvas",
    resourceKind: isMetricsExplorer
      ? ResourceKind.Explore
      : ResourceKind.Canvas,
  };
}

// Dashboard breadcrumb options. When tagAsFolders is on and a tag folder is
// active, the dropdown is scoped to just that tag's dashboards.
export function buildVisualizationOptions({
  sortedVisualizations,
  dashboardsByTag,
  activeTag,
}: {
  sortedVisualizations: V1Resource[];
  dashboardsByTag: Map<string, V1Resource[]>;
  activeTag: string | undefined;
}): Map<string, PathOption> {
  const map = new Map<string, PathOption>();
  const scopedResources = activeTag
    ? (dashboardsByTag.get(activeTag) ?? [])
    : sortedVisualizations;
  for (const resource of scopedResources) {
    map.set(
      resource.meta.name.name.toLowerCase(),
      buildDashboardPathOption(resource),
    );
  }
  return map;
}
