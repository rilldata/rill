import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";

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
