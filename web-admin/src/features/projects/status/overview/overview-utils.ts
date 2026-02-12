import {
  ResourceKind,
  prettyResourceKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

export type ResourceCount = { kind: string; label: string; count: number };

export const displayKinds = [
  ResourceKind.Source,
  ResourceKind.Model,
  ResourceKind.MetricsView,
  ResourceKind.Explore,
  ResourceKind.Canvas,
  ResourceKind.Alert,
  ResourceKind.Report,
  ResourceKind.API,
  ResourceKind.Connector,
];

/**
 * Counts resources by kind, filtered to displayKinds and ordered to match.
 * Only includes kinds with count > 0.
 */
export function countByKind(resources: V1Resource[]): ResourceCount[] {
  const counts = new Map<string, number>();
  for (const r of resources) {
    const kind = r.meta?.name?.kind;
    if (kind) counts.set(kind, (counts.get(kind) ?? 0) + 1);
  }
  return displayKinds
    .filter((kind) => (counts.get(kind) ?? 0) > 0)
    .map((kind) => ({
      kind,
      label: prettyResourceKind(kind),
      count: counts.get(kind) ?? 0,
    }));
}

/**
 * Groups errored resources by kind, sorted by count descending.
 */
export function groupErrorsByKind(resources: V1Resource[]): ResourceCount[] {
  const counts = new Map<string, number>();
  for (const r of resources) {
    const kind = r.meta?.name?.kind;
    if (kind) counts.set(kind, (counts.get(kind) ?? 0) + 1);
  }
  return Array.from(counts.entries())
    .map(([kind, count]) => ({
      kind,
      label: prettyResourceKind(kind),
      count,
    }))
    .sort((a, b) => b.count - a.count);
}
