import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";

// prepare the activeFilters to be sent to the server
export function prune(actives) {
  const filters: ActiveValues = {};
  for (const activeColumn in actives) {
    if (!actives[activeColumn].length) continue;
    filters[activeColumn] = actives[activeColumn];
  }
  return filters;
}
