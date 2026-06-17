import type { SearchableItem, GroupedResults } from "./types";

const MAX_RESULTS_PER_GROUP = 5;
const MIN_QUERY_LENGTH = 2;

export function searchIndex(
  items: SearchableItem[],
  query: string,
): GroupedResults {
  if (query.length < MIN_QUERY_LENGTH) {
    return { projects: [], dashboards: [], reports: [], alerts: [] };
  }

  const q = query.toLowerCase();
  const matched = items.filter(
    (item) =>
      item.name.toLowerCase().includes(q) ||
      item.projectName.toLowerCase().includes(q),
  );

  return groupResults(matched, MAX_RESULTS_PER_GROUP);
}

export function groupResults(
  items: SearchableItem[],
  limit?: number,
): GroupedResults {
  const groups: GroupedResults = {
    projects: [],
    dashboards: [],
    reports: [],
    alerts: [],
  };

  for (const item of items) {
    switch (item.type) {
      case "project":
        groups.projects.push(item);
        break;
      case "explore":
      case "canvas":
        groups.dashboards.push(item);
        break;
      case "report":
        groups.reports.push(item);
        break;
      case "alert":
        groups.alerts.push(item);
        break;
    }
  }

  if (limit) {
    groups.projects = groups.projects.slice(0, limit);
    groups.dashboards = groups.dashboards.slice(0, limit);
    groups.reports = groups.reports.slice(0, limit);
    groups.alerts = groups.alerts.slice(0, limit);
  }

  return groups;
}
