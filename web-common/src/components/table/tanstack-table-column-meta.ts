import type { RowData } from "tanstack-table-8-svelte-5";
import type { ComponentType, SvelteComponent } from "svelte";
import type { PivotMeasureFormatting } from "@rilldata/web-common/features/dashboards/pivot/types";

declare module "tanstack-table-8-svelte-5" {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData extends RowData, TValue> {
    widthPercent?: number;
    marginLeft?: string;
    icon?: ComponentType<SvelteComponent>;
    /** Full path of dimension name/value pairs from root to this header */
    dimensionPath?: Record<string, string>;
    /** Description shown in the header name tooltip */
    description?: string;
    tooltipFormatter?: (value: unknown) => string | null | undefined;
    /** Conditional formatting config for a main measure column (heatmap/data bar) */
    conditionalFormat?: PivotMeasureFormatting;
    /** Base measure name, used to look up the per-measure color domain */
    measureName?: string;
    /** True for the row-totals column leaves, which are excluded from formatting */
    isRowTotal?: boolean;
  }
}
