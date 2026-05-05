import { getContext, setContext } from "svelte";

const GRAPH_NAV_KEY = Symbol("graph-navigation");

export interface GraphNavigationCallbacks {
  /** Navigate to show the lineage tree for a specific resource */
  viewLineage: (kind: string | null, resourceName: string) => void;
  /** Navigate to open the resource's file/definition */
  openFile?: (filePath: string) => void;
}

export function setGraphNavigation(callbacks: GraphNavigationCallbacks) {
  setContext(GRAPH_NAV_KEY, callbacks);
}

export function getGraphNavigation(): GraphNavigationCallbacks | undefined {
  return getContext<GraphNavigationCallbacks | undefined>(GRAPH_NAV_KEY);
}
