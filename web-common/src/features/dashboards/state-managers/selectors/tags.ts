import {
  buildTagIndex,
  type TagIndex,
} from "@rilldata/web-common/components/menu/tag-utils";
import { allDimensions } from "./dimensions";
import { allMeasures } from "./measures";
import type { DashboardDataSources } from "./types";

export const dimensionTagIndex = (dashData: DashboardDataSources): TagIndex =>
  buildTagIndex(allDimensions(dashData));

export const measureTagIndex = (dashData: DashboardDataSources): TagIndex =>
  buildTagIndex(allMeasures(dashData));

// Combined index over dimensions + measures, used by surfaces (e.g. the pivot
// sidebar) where both kinds participate in a single tag-filtered list.
export const combinedTagIndex = (dashData: DashboardDataSources): TagIndex =>
  buildTagIndex([...allDimensions(dashData), ...allMeasures(dashData)]);

export const tagSelectors = {
  dimensionTagIndex,
  measureTagIndex,
  combinedTagIndex,
};
