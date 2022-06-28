import type { ActiveValues } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";

export function getFilterFromFilters(filters: ActiveValues): string {
  return Object.keys(filters)
    .map((field) => {
      return filters[field]
        .map(([value, filterType]) =>
          filterType ? `"${field}" = '${value}'` : `"${field}" != '${value}'`
        )
        .join(" OR ");
    })
    .join(" AND ");
}
