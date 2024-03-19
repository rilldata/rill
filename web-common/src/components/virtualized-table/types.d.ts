export type HeaderPosition = "top" | "left" | "top-left";

export type PinnedColumnSide = "left" | "right";

import type { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { SvelteComponent } from "svelte";

export interface VirtualizedTableColumns {
  name: string;
  type: string;
  largestStringLength?: number;
  summary?: ProfileColumnSummary | any;
  label?: string | typeof SvelteComponent<any>;
  total?: number;
  description?: string;
  enableResize?: boolean;
  // is this column highlighted in the table
  highlight?: boolean;
  // Is this the table sorted by this column, and if so, in what direction?
  // Leave undefined if the table is not sorted by this column.
  sorted?: SortDirection;
  format?: string;
}
