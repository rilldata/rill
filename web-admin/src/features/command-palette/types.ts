export type SearchableItemType =
  | "project"
  | "explore"
  | "canvas"
  | "report"
  | "alert";

export interface SearchableItem {
  name: string;
  type: SearchableItemType;
  projectName: string;
  orgName: string;
  route: string;
}

export interface GroupedResults {
  projects: SearchableItem[];
  dashboards: SearchableItem[];
  reports: SearchableItem[];
  alerts: SearchableItem[];
}
