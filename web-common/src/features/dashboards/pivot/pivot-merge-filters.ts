/**
 * Merge two filters together.
 * This might change later when move to the newer
 * filter format.
 */

import type {
  MetricsViewFilterCond,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";

function mergeArrays<T>(arr1: T[], arr2: T[]): T[] {
  return Array.from(new Set([...arr1, ...arr2]));
}

function mergeFilterConds(
  cond1: MetricsViewFilterCond[],
  cond2: MetricsViewFilterCond[],
): MetricsViewFilterCond[] {
  const merged: MetricsViewFilterCond[] = [];
  const allNames = new Set([
    ...cond1.map((c) => c.name),
    ...cond2.map((c) => c.name),
  ]);

  allNames.forEach((name) => {
    const cond1Entry = cond1.find((c) => c.name === name);
    const cond2Entry = cond2.find((c) => c.name === name);

    if (cond1Entry && cond2Entry) {
      merged.push({
        name,
        in: mergeArrays(cond1Entry.in || [], cond2Entry.in || []),
        like: cond1Entry.like || [],
      });
    } else {
      // If the condition only exists in one of the filters, add it directly.
      merged.push((cond1Entry || cond2Entry) as MetricsViewFilterCond);
    }
  });

  return merged;
}

export function mergeFilters(
  filter1: V1MetricsViewFilter,
  filter2: V1MetricsViewFilter,
): V1MetricsViewFilter {
  return {
    include: mergeFilterConds(filter1.include || [], filter2.include || []),
    exclude: mergeFilterConds(filter1.exclude || [], filter2.exclude || []),
  };
}
